#!/usr/bin/env bash

set -e

# Color definitions
RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
WHITE='\033[0;37m'
BOLD_WHITE='\033[1;37m'
RESET='\033[0m'

# Variables
NAME=""
NAMESPACE=""
ETCD_NAME=""
ETCD_NAMESPACE=""
SKIP_CONFIRM=false
DEFAULT_KUBECONFIG="${HOME}/.kube/config"
KUBECTL_IMAGE="bitnami/kubectl:1.30.7"
KUBECTL=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        --cluster)
        NAME="$2"
        shift
        shift
        ;;
        --namespace)
        NAMESPACE="$2"
        shift
        shift
        ;;
        --etcd)
        ETCD_NAME="$2"
        shift
        shift
        ;;
        --etcd-namespace)
        ETCD_NAMESPACE="$2"
        shift
        shift
        ;;
        -y)
        SKIP_CONFIRM=true
        shift
        ;;
        *)
        echo -e "${RED}Unknown option: $1${RESET}"
        exit 1
        ;;
    esac
done

check_prerequisites() {
  echo -e "${GREEN}=> Check prerequisites...${RESET}"

  # Check if kubectl is available
  if command -v kubectl &> /dev/null; then
      KUBECTL="kubectl"
  # Check if docker is available

  elif command -v docker &> /dev/null; then
      # Check if kubeconfig file exists
      if [ ! -f "$DEFAULT_KUBECONFIG" ]; then
          echo -e "${RED}Error: kubeconfig file not found at $DEFAULT_KUBECONFIG${RESET}"
          exit 1
      fi
      KUBECTL="docker run --network host --rm -v $DEFAULT_KUBECONFIG:/.kube/config $KUBECTL_IMAGE"

  else
      echo -e "${RED}Error: Neither kubectl or docker is installed!${RESET}"
      echo -e "${YELLOW}You can install kubectl for following the instructions here: https://kubernetes.io/docs/tasks/tools/${RESET}"
      echo -e "${YELLOW}You can install Docker for the instructions here: https://docs.docker.com/get-docker/${RESET}"
      exit 1
  fi

  echo -e "${GREEN}<= All prerequisites are met.${RESET}"
}

wait_for() {
  local command="$1"
  local timeout="$2"    # timeout in seconds.
  local interval=1      # check every 1 seconds.
  local elapsed_time=0  # elapsed time in seconds.

  while true; do
    # Turn off the exit-on-error flag to avoid exiting the script.
    set +e
    eval "$command" &> /dev/null
    local status=$?
    set -e

    # Check if command was successful.
    if [[ $status -eq 0 ]]; then
      return 0
    fi

    # Update elapsed time.
    ((elapsed_time += interval))

    # Check if the timeout has been reached.
    if [[ $elapsed_time -ge $timeout ]]; then
      echo "Timeout reached. Command failed."
      return 1
    fi

    # Wait for a specified interval before retrying.
    sleep $interval
  done
}

# Delete the GreptimeDBCluster
delete_cluster() {
  if [[ -z "$NAME" || -z "$NAMESPACE" ]]; then
      return 0
  fi

  if $KUBECTL get greptimedbclusters "$NAME" -n "$NAMESPACE" &>/dev/null; then
      echo -e "${GREEN}Deleting GreptimeDBCluster '$NAME' in namespace '$NAMESPACE'...${RESET}"

      # Wait for the resource to be deleted with a timeout of 100 seconds
      TIMEOUT=100  # timeout in seconds
      COMMAND="$KUBECTL delete greptimedbclusters \"$NAME\" -n \"$NAMESPACE\""

      if wait_for "$COMMAND" "$TIMEOUT"; then
          echo -e "${GREEN}Successfully deleted GreptimeDBCluster '$NAME' in namespace '$NAMESPACE'.${RESET}"
      else
          echo -e "${RED}GreptimeDBCluster '$NAME' was not deleted in $TIMEOUT seconds.${RESET}"
          exit 1
      fi
  else
      echo -e "${RED}GreptimeDBCluster '${NAME}' does not exist in namespace '${NAMESPACE}'.${RESET}"
  fi
}

# Delete the datanode PVCs
delete_pvc() {
  if [[ -z "$NAME" || -z "$NAMESPACE" ]]; then
      return 0
  fi

  PVC_LABEL="app.greptime.io/component=${NAME}-datanode"

  # Find and delete PVCs with the specified label
  PVCs=$($KUBECTL get pvc -n "$NAMESPACE" -l "$PVC_LABEL" -o jsonpath='{.items[*].metadata.name}' 2>/dev/null)

  if [ -z "$PVCs" ]; then
      echo -e "${RED}No datanode PVCs found.${RESET}"
      return
  fi

  for PVC in $PVCs; do
      echo -e "${BLUE}Deleting PVC: ${PVC}${RESET}"

      TIMEOUT=30
      COMMAND="$KUBECTL delete pvc \"$PVC\" -n \"$NAMESPACE\""

      if ! wait_for "$COMMAND" "$TIMEOUT"; then
          echo -e "${RED}PVC '$PVC' was not deleted in $TIMEOUT seconds.${RESET}"
      else
          echo -e "${GREEN}Successfully deleted PVC '$PVC'.${RESET}"
      fi
  done
}

detect_etcd_leader() {
    local leader=""
    for pod in $ETCD_PODS; do
        echo -e "${BLUE}Checking pod $pod for leader status...${RESET}" >&2
        local endpoint_status=$($KUBECTL exec -n "$ETCD_NAMESPACE" "$pod" -- etcdctl endpoint status --write-out=json 2>/dev/null)

        if [ -z "$endpoint_status" ]; then
            echo -e "${BLUE}Failed to get status from $pod. Skipping.${RESET}" >&2
            continue
        fi

        # Handle both array and single object response formats
        local member_id=$(echo "$endpoint_status" | jq -r 'if type == "array" then .[0].Status.header.member_id else .Status.header.member_id end')
        local leader_id=$(echo "$endpoint_status" | jq -r 'if type == "array" then .[0].Status.leader else .Status.leader end')

        if [ -z "$member_id" ] || [ -z "$leader_id" ]; then
            echo -e "${RED}Invalid status format from pod $pod.${RESET}" >&2
            continue
        fi

        if [ "$member_id" = "$leader_id" ]; then
            leader="$pod"
            echo -e "${GREEN}Leader found: $pod (member_id=$member_id)${RESET}" >&2
            break
        fi
    done

    if [ -z "$leader" ]; then
        echo -e "${RED}Error: No etcd leader detected.${RESET}" >&2
        return 1
    fi
    echo "$leader"
}

delete_etcd_data() {
  if [[ -z "$ETCD_NAME" || -z "$ETCD_NAMESPACE" ]]; then
      return 0
  fi

  echo -e "${BLUE}Starting etcd data cleanup...${RESET}"

  # Get etcd StatefulSet pods
  ETCD_SELECTOR=$($KUBECTL get statefulset -n "$ETCD_NAMESPACE" "$ETCD_NAME" -o jsonpath='{.spec.selector.matchLabels}' | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|join(",")')

  if [ -z "$ETCD_SELECTOR" ]; then
      echo -e "${RED}Failed to get ETCD StatefulSet ${ETCD_NAME} in namespace ${ETCD_NAMESPACE}${RESET}"
      exit 1
  fi

  # Get only Running etcd pods using the StatefulSet's selector
  ETCD_PODS=$($KUBECTL get pods -n "$ETCD_NAMESPACE" -l "$ETCD_SELECTOR" --field-selector status.phase=Running -o jsonpath='{.items[*].metadata.name}')

  if [ -z "$ETCD_PODS" ]; then
      echo -e "${RED}No Running etcd pods found for StatefulSet '${ETCD_NAME}' in namespace '${ETCD_NAMESPACE}'.${RESET}"
      exit 1
  fi

  ETCD_LEADER=$(detect_etcd_leader)

  if [ -z "$ETCD_LEADER" ]; then
      echo -e "${RED}Failed to detect etcd leader.${RESET}"
      exit 1
  fi

  # Delete all data from etcd
  echo -e "${BLUE}Deleting all data from etcd...${RESET}"
  if $KUBECTL exec -n "$ETCD_NAMESPACE" "$ETCD_LEADER" -- etcdctl del --prefix ""; then
      echo -e "${GREEN}Successfully deleted all data from etcd.${RESET}"
  else
      echo -e "${RED}Failed to delete data from etcd.${RESET}"
      exit 1
  fi
}

confirm_action() {
    local level="$1"  # 1 for first confirmation, 2 for second confirmation

    if [ "$SKIP_CONFIRM" = true ]; then
        echo -e "${GREEN}Auto-confirmed${RESET}"
        return 0
    fi

    if [ "$level" -eq 1 ]; then
        # First confirmation
        echo -e "${YELLOW}Would you like to cleanup the GreptimeDBCluster '${NAME}' in namespace '${NAMESPACE}', delete the datanode PVCs, and remove the metadata in etcd StatefulSet '${ETCD_NAME}' in namespace '${ETCD_NAMESPACE}'? type 'yes' or 'no': ${RESET}"
        read confirmation
        if [ "$confirmation" != "yes" ]; then
            echo -e "${RED}Operation cancelled.${RESET}"
            exit 0
        fi
    else
        # Second confirmation
        echo -e "${YELLOW}Please confirm again, would you like to cleanup the GreptimeDBCluster '${NAME}' in namespace '${NAMESPACE}', delete the datanode PVCs, and remove the metadata in etcd StatefulSet '${ETCD_NAME}' in namespace '${ETCD_NAMESPACE}'? type 'yes' or 'no': ${RESET}"
        read confirmation
        if [ "$confirmation" != "yes" ]; then
            echo -e "${RED}Operation cancelled.${RESET}"
            exit 0
        fi
    fi
}

main() {
    check_prerequisites

    # First confirmation
    confirm_action 1

    # Second confirmation
    confirm_action 2

    # Execute deletions
    delete_cluster
    delete_pvc
    delete_etcd_data

    echo -e "${GREEN}Cleanup completed successfully.${RESET}"
    echo -e "${YELLOW}Please remember to check and manually delete any associated Object Storage files, RDS and Kafka data if exist.${RESET}"
}

main
