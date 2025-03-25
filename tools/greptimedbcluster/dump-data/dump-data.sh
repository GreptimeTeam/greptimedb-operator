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

# Create output directory
OUTPUT_DIR="/tmp/greptime_dump_$(date +%Y%m%d_%H%M%S)"
mkdir -p "${OUTPUT_DIR}/pod_logs"

# Input the GreptimeDBCluster name and namespace
read -p "$(echo -e "${GREEN}Enter GreptimeDBCluster name: ${RESET}")" NAME
read -p "$(echo -e "${GREEN}Enter GreptimeDBCluster namespace: ${RESET}")" NAMESPACE

# Input the ETCD name and namespace
read -p "$(echo -e "${GREEN}Enter the name of the ETCD StatefulSet: ${RESET}")" ETCD_NAME
read -p "$(echo -e "${GREEN}Enter the namespace of the ETCD StatefulSet: ${RESET}")" ETCD_NAMESPACE

echo -e "${GREEN}Starting GreptimeDB environment information collection...${RESET}"

# Collect global resource information
echo -e "${CYAN}[1/6] Collecting global resources...${RESET}"
{
  echo "===== GreptimeDB Clusters ====="
  kubectl get greptimedbclusters "${NAME}" -n "${NAMESPACE}"
} > "${OUTPUT_DIR}/1_global_resources.txt"

# Collect specific cluster information
echo -e "${CYAN}[2/6] Collecting cluster (${NAMESPACE}/${NAME}) information...${RESET}"
{
  echo "===== GreptimeDBCluster YAML ====="
  kubectl get greptimedbcluster "${NAME}" -n "${NAMESPACE}" -oyaml

  echo -e "\n\n===== Related Pods ====="
  kubectl get pod -n "${NAMESPACE}" -owide

  echo -e "\n\n===== Related PVCs ====="
  kubectl get pvc -n "${NAMESPACE}"

  echo -e "\n\n===== Related Services ====="
  kubectl get svc -n "${NAMESPACE}"

  echo -e "\n\n===== Related Events ====="
  kubectl get events -n "${NAMESPACE}" --sort-by='.metadata.creationTimestamp'
} > "${OUTPUT_DIR}/2_cluster_${NAMESPACE}_${NAME}.txt"

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

# Collect ETCD status
echo -e "${CYAN}[3/6] Collecting ETCD status...${RESET}"
{
  echo "===== ETCD Pod Status ====="
  kubectl get pods -l "$ETCD_SELECTOR" -n "$ETCD_NAMESPACE" -owide

  echo -e "\n\n===== ETCD Member Status ====="
  for pod in $ETCD_PODS; do
    echo "Using Pod: ${ETCD_NAMESPACE}/${pod}"
    kubectl exec -n "${ETCD_NAMESPACE}" "${pod}" -- etcdctl endpoint status --write-out=table
    echo -e "\n\n===== ETCD Health Status ====="
    kubectl exec -n "${ETCD_NAMESPACE}" "${pod}" -- etcdctl endpoint health --write-out=table

    timeout 30 kubectl logs -n "${ETCD_NAMESPACE}" "${pod}" > "${OUTPUT_DIR}/pod_logs/${pod}.log" 2>&1
  done
} > "${OUTPUT_DIR}/3_etcd_status.txt"

# Collect pod details and logs
echo -e "${CYAN}[4/6] Collecting pod details and logs...${RESET}"
kubectl get pods -n "${NAMESPACE}" -o name | while read -r pod; do
  pod_name=${pod#pod/}
  echo "Processing Pod: ${pod_name}"

  {
    echo "===== Pod Description ====="
    kubectl describe -n "${NAMESPACE}" "${pod}"

    echo -e "\n\n===== Pod Configuration ====="
    kubectl get -n "${NAMESPACE}" "${pod}" -oyaml
  } > "${OUTPUT_DIR}/pod_${pod_name}_info.txt"

  # Collect logs with a timeout to prevent hanging
  timeout 30 kubectl logs -n "${NAMESPACE}" "${pod}" --all-containers=true --timestamps > "${OUTPUT_DIR}/pod_logs/${pod_name}.log" 2>&1 || true

  # Collect previous logs if available
  timeout 30 kubectl logs -n "${NAMESPACE}" "${pod}" --all-containers=true --timestamps --previous > "${OUTPUT_DIR}/pod_logs/${pod_name}_previous.log" 2>&1 || true
done

sql_query() {
  local query="$1"

  # Get the frontend pod name
  FRONTEND=$(kubectl get pods -n "$NAMESPACE" --no-headers | grep "^$NAME-frontend" | awk '{print $1}' | head -n 1)

  if [ -z "$FRONTEND" ]; then
    echo -e "${RED}No frontend pod found matching the pattern ${NAME}-frontend in namespace ${NAMESPACE}.${RESET}"
    exit 1
  fi

  # Get the raw JSON data
  response=$(kubectl exec -it "$FRONTEND" -n "$NAMESPACE" -- /bin/bash -c "curl -s -X POST -H 'Content-Type: application/x-www-form-urlencoded' -d 'sql=${query}' http://localhost:4000/v1/sql")

  # Extract column names
  columns=$(echo "$response" | tr -d '\n' | grep -o '"column_schemas":\[.*\]' | grep -o '"name":"[^"]*"' | sed 's/"name":"//;s/"//g')

  # Extract row data
  rows=$(echo "$response" | tr -d '\n' | grep -o '"rows":\[\[.*\]\]' | sed 's/"rows":\[\[//;s/\]\]//' | sed 's/\],\[/\n/g' | sed 's/"//g')

  # Get the maximum width of each column
  max_widths=()
  for col in $columns; do
    max_widths+=(${#col})
  done

  # Calculate the maximum length for each row
  while IFS= read -r line; do
    IFS=',' read -ra values <<< "$line"
    for i in "${!values[@]}"; do
      len=${#values[i]}
      if [ "$len" -gt "${max_widths[i]:-0}" ]; then
        max_widths[$i]=$len
      fi
    done
  done <<< "$rows"

  # Generate the separator line
  divider="+"
  for w in "${max_widths[@]}"; do
    divider="$divider$(printf '%*s' $((w+2)) | tr ' ' '-')+"
  done

  # Print the header
  echo "$divider"
  printf "|"
  i=1
  for col in $columns; do
    w=${max_widths[i-1]:-0}
    printf " %-*s |" "$w" "$col"
    i=$((i+1))
  done
  printf "\n%s\n" "$divider"

  # Print the data rows
  echo "$rows" | while IFS= read -r line; do
    printf "|"
    IFS=',' read -ra values <<< "$line"
    for i in "${!values[@]}"; do
      w=${max_widths[i]:-0}
      printf " %-*s |" "$w" "${values[i]}"
    done
    echo
  done

  # Print the footer
  echo "$divider"
  row_count=$(echo "$rows" | wc -l)
  printf "%d rows in set\n" "$row_count"
}

# Collect information_schema table
echo -e "${CYAN}[5/6] Collecting information_schema table...${RESET}"
{
    echo -e "\n\n===== build_info Table ====="
    sql_query 'SELECT * FROM build_info'

    echo -e "\n\n===== cluster_info Table ====="
    sql_query 'SELECT * FROM cluster_info'

} > "${OUTPUT_DIR}/5_information_schema_table.txt" 2>&1

# Package all collected information
echo -e "${CYAN}[6/6] Packaging collected information...${RESET}"
tar czf "${OUTPUT_DIR}.tar.gz" "${OUTPUT_DIR}"

echo -e "${GREEN}Collection complete! All data saved to: ${OUTPUT_DIR}.tar.gz${RESET}"
echo -e "Contents:"
ls -lh "${OUTPUT_DIR}"/*
