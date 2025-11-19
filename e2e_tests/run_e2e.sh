#!/usr/bin/env bash

PORT="${1:-19500}"
EXPORTER_TLS="${2:-false}"
EXPORTER_AUTH="${3:-false}"
CERT_PATH="${4:-}"
MODE="${5:-}"

# Users for test basic auth.
AUTH_USER="test"
AUTH_PASSWORD="test"

# Cert auth.
AUTH_CERT="user.pem"
AUTH_KEY="user.key"

# Use http or https.
case ${EXPORTER_TLS} in
    "false")
        EXPORTER_URL="http://localhost:${PORT}/metrics"
        CURL_FLAGS=""
        ;;
    "true")
        EXPORTER_URL="https://localhost:${PORT}/metrics"
        CURL_FLAGS="-k"
        ;;
    *)
        echo "[ERROR] incorrect value: get=${EXPORTER_TLS}, want=true or false"
        exit 1
        ;;
esac

# Use basic auth, cert or not.
case ${EXPORTER_AUTH} in
    "false")
        ;;
    "basic")
        CURL_FLAGS+=" -u ${AUTH_USER}:${AUTH_PASSWORD}"
        ;;
    "cert")
        CURL_FLAGS+=" --cert ${CERT_PATH}/${AUTH_CERT} --key ${CERT_PATH}/${AUTH_KEY}"
        ;;
    *)
        echo "[ERROR] incorect value: get=${EXPORTER_AUTH}, want=false, basic or cert"
        exit 1
        ;;
esac

# Check if exporter is available.
MAX_ATTEMPTS=10
ATTEMPT=0
SLEEP_SECONDS=5

while [ ${ATTEMPT} -lt ${MAX_ATTEMPTS} ]; do
    ATTEMPT=$((ATTEMPT + 1))
    if curl -s ${CURL_FLAGS} -o /dev/null -w "%{http_code}" ${EXPORTER_URL} | grep -q "200"; then
        break
    else
        echo "[INFO] exporter not available yet (attempt ${ATTEMPT})"
        if [ ${ATTEMPT} -lt ${MAX_ATTEMPTS} ]; then
            sleep ${SLEEP_SECONDS}
        else
            echo "[WARNING] exporter is not available after ${MAX_ATTEMPTS} attempts"
        fi
    fi
done

# A simple test to check the number of metrics.
# Format: regex for metric | repetitions.
case "${MODE}" in
    "only_diff_prefix")
    # Check inly part of metrics.
    # No full backup with this prefix.
        declare -a REGEX_LIST=(
    '^medusa_backup_info{.*,backup_type="differential",prefix="only_diff_prefix",.*} 1$|1'
    '^medusa_backup_info{.*,backup_type="full",prefix="only_diff_prefix",.*}|0'
    '^medusa_backup_last_duration_seconds{backup_type="differential"}|1'
    '^medusa_backup_last_duration_seconds{backup_type="full"}|0'
    '^medusa_backup_last_objects{backup_type="differential"}|1'
    '^medusa_backup_last_objects{backup_type="full"}|0'
    '^medusa_backup_last_size_bytes{backup_type="differential"}|1'
    '^medusa_backup_last_size_bytes{backup_type="full"}|0'
    '^medusa_backup_status{.*,backup_type="differential"} 1$|1'
    '^medusa_exporter_status{prefix="only_diff_prefix"} 1$|1'
    '^medusa_node_backup_info{.*,backup_type="differential",.*} 1$|1'
    '^medusa_node_backup_status{.*,backup_type="differential",.*} 1$|1'
        )
        ;;
    "only_full_prefix")
    # Check inly part of metrics.
    # No differential backup with this prefix.
        declare -a REGEX_LIST=(
    '^medusa_backup_info{.*,backup_type="differential",prefix="only_full_prefix",.*}|0'
    '^medusa_backup_info{.*,backup_type="full",prefix="only_full_prefix",.*} 1$|1'
    '^medusa_backup_last_duration_seconds{backup_type="differential"}|0'
    '^medusa_backup_last_duration_seconds{backup_type="full"}|1'
    '^medusa_backup_last_objects{backup_type="differential"}|0'
    '^medusa_backup_last_objects{backup_type="full"}|1'
    '^medusa_backup_last_size_bytes{backup_type="differential"}|0'
    '^medusa_backup_last_size_bytes{backup_type="full"}|1'
    '^medusa_backup_status{.*,backup_type="full"} 1$|1'
    '^medusa_exporter_status{prefix="only_full_prefix"} 1$|1'
    '^medusa_node_backup_info{.*,backup_type="full",.*} 1$|1'
    '^medusa_node_backup_status{.*,backup_type="full",.*} 1$|1'
        )
        ;;
    *)
        declare -a REGEX_LIST=(
    '^medusa_backup_completed_nodes{.*,backup_type="differential"} 1$|1'
    '^medusa_backup_completed_nodes{.*,backup_type="full"} 1$|1'
    '^medusa_backup_duration_seconds{.*,backup_type="differential",.*}|1'
    '^medusa_backup_duration_seconds{.*,backup_type="full",.*}|1'
    '^medusa_backup_incomplete_nodes{.*,backup_type="differential"} 0$|1'
    '^medusa_backup_incomplete_nodes{.*,backup_type="full"} 0$|1'
    '^medusa_backup_info{.*,backup_type="differential",prefix="no-prefix",.*} 1$|1'
    '^medusa_backup_info{.*,backup_type="full",prefix="no-prefix",.*} 1$|1'
    '^medusa_backup_last_duration_seconds{backup_type="differential"}|1'
    '^medusa_backup_last_duration_seconds{backup_type="full"}|1'
    '^medusa_backup_last_objects{backup_type="differential"}|1'
    '^medusa_backup_last_objects{backup_type="full"}|1'
    '^medusa_backup_last_size_bytes{backup_type="differential"}|1'
    '^medusa_backup_last_size_bytes{backup_type="full"}|1'
    '^medusa_backup_missing_nodes{.*,backup_type="differential"} 0$|1'
    '^medusa_backup_missing_nodes{.*,backup_type="full"} 0$|1'
    '^medusa_backup_objects{.*,backup_type="differential"}|1'
    '^medusa_backup_objects{.*,backup_type="full"}|1'
    '^medusa_backup_since_last_completion_seconds{backup_type="differential"}|1'
    '^medusa_backup_since_last_completion_seconds{backup_type="full"}|1'
    '^medusa_backup_size_bytes{.*,backup_type="differential"}|1'
    '^medusa_backup_size_bytes{.*,backup_type="full"}|1'
    '^medusa_backup_status{.*,backup_type="differential"} 1$|1'
    '^medusa_backup_status{.*,backup_type="full"} 1$|1'
    '^medusa_exporter_build_info{.*} 1$|1'
    '^medusa_exporter_status{prefix="no-prefix"} 1$|1'
    '^medusa_node_backup_duration_seconds{.*,backup_type="differential",.*"}|1'
    '^medusa_node_backup_duration_seconds{.*,backup_type="full",.*}|1'
    '^medusa_node_backup_info{.*,backup_type="differential",.*} 1$|1'
    '^medusa_node_backup_info{.*,backup_type="full",.*} 1$|1'
    '^medusa_node_backup_objects{.*,backup_type="differential",.*}|1'
    '^medusa_node_backup_objects{.*,backup_type="full",.*}|1'
    '^medusa_node_backup_size_bytes{.*,backup_type="differential",.*}|1'
    '^medusa_node_backup_size_bytes{.*,backup_type="full",.*}|1'
    '^medusa_node_backup_status{.*,backup_type="differential",.*} 1$|1'
    '^medusa_node_backup_status{.*,backup_type="full",.*} 1$|1'
        )
        ;;
esac

# Check results.
for i in "${REGEX_LIST[@]}"
do
    regex=$(echo ${i} | cut -f1 -d'|')
    cnt=$(echo ${i} | cut -f2 -d'|')
    metric_cnt=$(curl -s ${CURL_FLAGS} ${EXPORTER_URL} | grep -E "${regex}" | wc -l | tr -d ' ')
    if [[ ${metric_cnt} != ${cnt} ]]; then
        echo "[ERROR] on regex '${regex}': get=${metric_cnt}, want=${cnt}"
        exit 1
    fi
done

echo "[INFO] all tests passed"
exit 0
