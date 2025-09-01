#!/usr/bin/env bash

set -e

# Color definitions
RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
BLUE='\033[1;34m'
CYAN='\033[1;36m'
WHITE='\033[0;37m'
BOLD_WHITE='\033[1;37m'
RESET='\033[0m'

# Variables
NAME=""
NAMESPACE=""
ETCD_NAME=""
ETCD_NAMESPACE=""
DEFAULT_KUBECONFIG="${HOME}/.kube/config"
KUBECTL_IMAGE="greptime/kubectl:1.32.3"
KUBECTL=""
OUTPUT_BASE_DIR="/tmp"
OUTPUT_DIR=""

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
        --output-dir)
        OUTPUT_BASE_DIR="$2"
        shift
        shift
        ;;
        *)
        echo -e "${RED}Unknown option: $1${RESET}"
        exit 1
        ;;
    esac
done

initialize_output_dir() {
    OUTPUT_DIR="${OUTPUT_BASE_DIR}/greptime_dump_$(date +%Y%m%d_%H%M%S)"
    mkdir -p "${OUTPUT_DIR}/pod_logs"
    echo -e "${GREEN}Output directory: ${OUTPUT_DIR}${RESET}"
}

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

collect_global_resources() {
    echo -e "${CYAN}[1/6] Collecting global resources...${RESET}"
    {
        echo "===== GreptimeDB Clusters ====="
        $KUBECTL get greptimedbclusters "${NAME}" -n "${NAMESPACE}"
    } > "${OUTPUT_DIR}/1_global_resources.txt"
}

collect_cluster_info() {
    echo -e "${CYAN}[2/6] Collecting cluster (${NAMESPACE}/${NAME}) information...${RESET}"
    {
        echo "===== GreptimeDBCluster YAML ====="
        $KUBECTL get greptimedbcluster "${NAME}" -n "${NAMESPACE}" -oyaml

        echo -e "\n\n===== Related Pods ====="
        $KUBECTL get pod -n "${NAMESPACE}" -owide

        echo -e "\n\n===== Related PVCs ====="
        $KUBECTL get pvc -n "${NAMESPACE}"

        echo -e "\n\n===== Related Services ====="
        $KUBECTL get svc -n "${NAMESPACE}"

        echo -e "\n\n===== Related Events ====="
        $KUBECTL get events -n "${NAMESPACE}" --sort-by='.metadata.creationTimestamp'
    } > "${OUTPUT_DIR}/2_cluster_${NAMESPACE}_${NAME}.txt"
}

collect_etcd_info() {
    if [[ -z "$ETCD_NAME" || -z "$ETCD_NAMESPACE" ]]; then
        echo -e "${YELLOW}Skipping ETCD info collection as ETCD name or namespace not provided${RESET}"
        return
    fi

    echo -e "${CYAN}[3/6] Collecting ETCD info...${RESET}"

    ETCD_SELECTOR=$($KUBECTL get statefulset -n "$ETCD_NAMESPACE" "$ETCD_NAME" -o jsonpath='{.spec.selector.matchLabels}' | jq -r 'to_entries|map("\(.key)=\(.value|tostring)")|join(",")')
    if [ -z "$ETCD_SELECTOR" ]; then
        echo -e "${RED}Failed to get ETCD StatefulSet ${ETCD_NAME} in namespace ${ETCD_NAMESPACE}${RESET}"
        return
    fi

    ETCD_PODS=$($KUBECTL get pods -n "$ETCD_NAMESPACE" -l "$ETCD_SELECTOR" --field-selector status.phase=Running -o jsonpath='{.items[*].metadata.name}')
    if [ -z "$ETCD_PODS" ]; then
        echo -e "${RED}No Running etcd pods found for StatefulSet '${ETCD_NAME}' in namespace '${ETCD_NAMESPACE}'.${RESET}"
        return
    fi

    {
        echo "===== ETCD StatefulSet info ====="
        $KUBECTL get sts "$ETCD_NAME" -n "$ETCD_NAMESPACE" -owide

        echo -e "\n\n===== ETCD StatefulSet YAML ====="
        $KUBECTL get sts "$ETCD_NAME" -n "$ETCD_NAMESPACE" -oyaml

        echo -e "\n\n===== ETCD Pod Status ====="
        $KUBECTL get pods -l "$ETCD_SELECTOR" -n "$ETCD_NAMESPACE" -owide

        for pod in $ETCD_PODS; do
            echo -e "\n\n=====Using Pod: ${ETCD_NAMESPACE}/${pod}"
            echo -e "\n\n===== ETCD Member Status ====="
            $KUBECTL exec -n "${ETCD_NAMESPACE}" "${pod}" -- etcdctl endpoint status --write-out=table
            echo -e "\n\n===== ETCD Health Status ====="
            $KUBECTL exec -n "${ETCD_NAMESPACE}" "${pod}" -- etcdctl endpoint health --write-out=table

            timeout 30 $KUBECTL logs -n "${ETCD_NAMESPACE}" "${pod}" > "${OUTPUT_DIR}/pod_logs/${pod}.log" 2>&1
        done
    } > "${OUTPUT_DIR}/3_etcd_info.txt"
}

collect_pod_details() {
    echo -e "${CYAN}[4/6] Collecting GreptimeDB pod details and logs...${RESET}"
    $KUBECTL get pods -n "${NAMESPACE}" -o name | while read -r pod; do
        pod_name=${pod#pod/}
        {
            echo "===== Pod Description ====="
            $KUBECTL describe -n "${NAMESPACE}" "${pod}"

            echo -e "\n\n===== Pod Configuration ====="
            $KUBECTL get -n "${NAMESPACE}" "${pod}" -oyaml
        } > "${OUTPUT_DIR}/4_pod_${pod_name}_info.txt"

        # Collect logs with a timeout to prevent hanging
        timeout 30 $KUBECTL logs -n "${NAMESPACE}" "${pod}" --all-containers=true --timestamps > "${OUTPUT_DIR}/pod_logs/${pod_name}.log" 2>&1 || true

        # Collect previous logs if available
        timeout 30 $KUBECTL logs -n "${NAMESPACE}" "${pod}" --all-containers=true --timestamps --previous > "${OUTPUT_DIR}/pod_logs/${pod_name}_previous.log" 2>&1 || true
    done
}

sql_query() {
    local query=$1

    # Get the frontend pod name
    FRONTEND=$($KUBECTL get pods -n "$NAMESPACE" --no-headers --field-selector=status.phase=Running | grep "^$NAME-frontend" | awk '{print $1}' | head -n 1)

    if [ -z "$FRONTEND" ]; then
        echo -e "${RED}No frontend pod found matching the pattern ${NAME}-frontend in namespace ${NAMESPACE}.${RESET}"
        exit 1
    fi

    # Get the raw JSON data
    response=$($KUBECTL exec -it "$FRONTEND" -n "$NAMESPACE" -- /bin/bash -c "curl -s -X POST -H 'Content-Type: application/x-www-form-urlencoded' -d 'sql=$query' http://localhost:4000/v1/sql")

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

collect_information_schema() {
    echo -e "${CYAN}[5/6] Collecting information_schema table...${RESET}"
    {
        echo -e "===== build_info Table ====="
        sql_query "SELECT * FROM information_schema.build_info"

        echo -e "\n\n===== cluster_info Table ====="
        sql_query "SELECT * FROM information_schema.cluster_info"
    } > "${OUTPUT_DIR}/5_information_schema_table.txt" 2>&1
}

package_results() {
    echo -e "${CYAN}[6/6] Packaging collected information...${RESET}"

    # Create tar from the parent directory
    tar czf "${OUTPUT_DIR}.tar.gz" -C "${OUTPUT_BASE_DIR}" "$(basename "${OUTPUT_DIR}")"

    # Verify the archive was created
    if [ -f "${OUTPUT_DIR}.tar.gz" ]; then
        echo -e "${GREEN}Collection complete! All data saved to: ${OUTPUT_DIR}.tar.gz${RESET}"
        echo -e "Contents:"
        ls -lh "${OUTPUT_DIR}"/*
    else
        echo -e "${RED}Error: Failed to create archive ${OUTPUT_DIR}.tar.gz${RESET}"
        exit 1
    fi
}

main() {
    initialize_output_dir
    check_prerequisites
    collect_global_resources
    collect_cluster_info
    collect_etcd_info
    collect_pod_details
    collect_information_schema
    package_results
}

main
