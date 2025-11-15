#!/bin/bash
ENV_FILE=".env"

if [[ -f "${ENV_FILE}" ]]; then 
    set -a
    source "${ENV_FILE}"
    set +a
else 
    echo ".env not found at ${ENV_FILE}"
fi

if [[ -z "${DATABASE_DSN}" ]]; then 
    echo "DATABASE_DSN is required" >&2; exit 1; fi


cmd="${1:-up}"
shift || true
args=("$@")

echo "migrate -database ${DATABASE_DSN} -path ./migrations ${cmd} ${args[1]}"
migrate -database "${DATABASE_DSN}" -path ./migrations "${cmd}" "${args[@]}"