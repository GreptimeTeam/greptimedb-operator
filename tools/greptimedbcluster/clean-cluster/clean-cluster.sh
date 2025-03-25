#!/usr/bin/env bash

set -e

# Color definitions
RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
MAGENTA='\033[1;35m'
CYAN='\033[1;36m'
RESET='\033[0m'

# Input the GreptimeDBCluster name and namespace
read -p "$(echo -e "${GREEN}Enter the name of the GreptimeDBCluster: ${RESET}")" NAME
read -p "$(echo -e "${GREEN}Enter the namespace of the GreptimeDBCluster: ${RESET}")" NAMESPACE

# Input the ETCD name and namespace
read -p "$(echo -e "${GREEN}Enter the name of the ETCD StatefulSet: ${RESET}")" ETCD_NAME
read -p "$(echo -e "${GREEN}Enter the namespace of the ETCD StatefulSet: ${RESET}")" ETCD_NAMESPACE

# Delete the GreptimeDBCluster
if kubectl get greptimedbclusters "$NAME" -n "$NAMESPACE" &>/dev/null; then
    kubectl delete greptimedbclusters "$NAME" --namespace="$NAMESPACE"

    # Wait for the resource to be deleted with a timeout of 100 seconds
    TIMEOUT=100  # timeout in seconds
    INTERVAL=2   # check every 2 seconds
    ELAPSED=0    # elapsed time in seconds

    while kubectl get greptimedbclusters "$NAME" -n "$NAMESPACE" &>/dev/null; do
        sleep $INTERVAL
        ELAPSED=$((ELAPSED + INTERVAL))

        if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
            echo -e "${RED}Timeout: GreptimeDBCluster '$NAME' was not deleted in $TIMEOUT seconds.${RESET}"
            exit 1
        fi

        echo "Waiting for the GreptimeDBCluster '$NAME' to be deleted..."
    done

    echo -e "${BLUE}Successfully deleted GreptimeDBCluster '$NAME' in namespace '$NAMESPACE'.${RESET}"
else
    echo -e "${RED}GreptimeDBCluster '${NAME}' does not exist in namespace '${NAMESPACE}'.${RESET}"
fi

# Delete the datanode PVCs
PVC_LABEL="app.greptime.io/component=${NAME}-datanode"

# Find and delete PVCs with the specified label
PVCS=$(kubectl get pvc -n "$NAMESPACE" -l "$PVC_LABEL" -o jsonpath='{.items[*].metadata.name}')

if [ -z "$PVCS" ]; then
    echo -e "${RED}No PVCs found.${RESET}"
else
    for PVC in $PVCS; do
        echo -e "${YELLOW}Deleting PVC: ${PVC}${RESET}"
        kubectl delete pvc "$PVC" -n "$NAMESPACE"

        # Wait for the PVC to be deleted
        ELAPSED=0
        while kubectl get pvc "$PVC" -n "$NAMESPACE" &>/dev/null; do
            sleep $INTERVAL
            ELAPSED=$((ELAPSED + INTERVAL))
            if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
                echo -e "${RED}Timeout: PVC '$PVC' was not deleted in $TIMEOUT seconds.${RESET}"
                break
            fi
            echo "Waiting for PVC '$PVC' to be deleted..."
        done

        echo -e "${BLUE}Successfully deleted PVC '$PVC'.${RESET}"
    done
fi

# Clean up etcd data
echo -e "${YELLOW}Starting etcd data cleanup...${RESET}"

# Get etcd StatefulSet pods
ETCD_SELECTOR=$(kubectl get statefulset -n "$ETCD_NAMESPACE" "$ETCD_NAME" -o jsonpath='{.spec.selector.matchLabels}' | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|join(",")')

if [ -z "$ETCD_SELECTOR" ]; then
    echo -e "${RED}Failed to get StatefulSet ${ETCD_NAME} in namespace ${ETCD_NAMESPACE}${RESET}"
    exit 1
fi

# Get only Running etcd pods using the StatefulSet's selector
ETCD_PODS=$(kubectl get pods -n "$ETCD_NAMESPACE" -l "$ETCD_SELECTOR" --field-selector status.phase=Running -o jsonpath='{.items[*].metadata.name}')

if [ -z "$ETCD_PODS" ]; then
    echo -e "${RED}No Running etcd pods found for StatefulSet '${ETCD_NAME}' in namespace '${ETCD_NAMESPACE}'.${RESET}"
    exit 1
fi

detect_etcd_leader() {
    local leader=""
    for pod in $ETCD_PODS; do
        echo -e "${CYAN}Checking pod $pod for leader status...${RESET}" >&2
        local endpoint_status=$(kubectl exec -n "$ETCD_NAMESPACE" "$pod" -- etcdctl endpoint status --write-out=json 2>/dev/null)

        if [ -z "$endpoint_status" ]; then
            echo -e "${RED}Failed to get status from $pod. Skipping.${RESET}" >&2
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
    echo "$leader"  # Return just the pod name without any formatting
}
ETCD_LEADER=$(detect_etcd_leader)

if [ -z "$ETCD_LEADER" ]; then
    echo -e "${RED}Failed to detect etcd leader.${RESET}"
    exit 1
fi

# Delete all data from etcd
echo -e "${YELLOW}Deleting all data from etcd...${RESET}"
if kubectl exec -n "$ETCD_NAMESPACE" "$ETCD_LEADER" -- etcdctl del --prefix ""; then
    echo -e "${BLUE}Successfully deleted all data from etcd.${RESET}"
else
    echo -e "${RED}Failed to delete data from etcd.${RESET}"
    exit 1
fi
