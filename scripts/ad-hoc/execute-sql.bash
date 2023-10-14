#!/bin/bash

set -u

ENV=${ENV:-local}
ORG=${ORG:-manabie}
SQL_PROXY_CONN_NAME=${SQL_PROXY_CONN_NAME:-}
COMMAND=${COMMAND:-}
DB_PREFIX=${DB_PREFIX:-}
PROJECT_ID=${PROJECT:-}

# Install postgresql client (psql)
if ! command -v psql; then
  echo "psql not found, installing it..."
  apt update && apt install -y postgresql-client
fi
psql --version

# Install cloud_sql_proxy
if ! command -v cloud_sql_proxy; then 
  echo "cloud_sql_proxy not found, installing it..."
  if ! command -v curl; then
    apt update && apt install -y curl
  fi
  curl -o /usr/local/bin/cloud_sql_proxy -L https://storage.googleapis.com/cloudsql-proxy/v1.33.8/cloud_sql_proxy.linux.amd64
  chmod +x /usr/local/bin/cloud_sql_proxy
fi
cloud_sql_proxy --version

# Install netcat for nc command
if ! command -v netcat; then
  echo "nc not found, installing it..."
  apt update && apt install -y netcat
fi

# Install psmisc for fuser command
if ! command -v fuser; then
  echo "fuser not found, installing it..."
  apt update && apt install -y psmisc
fi
fuser --version

# Sanity check
if [[ "${ENV}" == "dorp" || "${ENV}" == "preproduction" ]]; then
  echo "Running for preproduction (ENV: ${ENV}), executing sanity check..."
  if [[ "${SQL_PROXY_CONN_NAME}" != *"clone"* ]]; then
    >&2 echo "Invalid Cloud SQL instance name for preproduction: expected \"clone\" in name, got ${SQL_PROXY_CONN_NAME}"
    exit 1
  fi
fi

count=0
cloud_sql_proxy -instances="${SQL_PROXY_CONN_NAME}=tcp:5432" \
  -structured_logs \
  -log_debug_stdout=true \
  -enable_iam_login &
echo "SQL_PROXY_CONN_NAME: ${SQL_PROXY_CONN_NAME}"
until nc -z 127.0.0.1 5432; do
  echo "Waiting for the proxy to run..."
  sleep 2
  if [[ $count == 10 ]]; then
    echo $count
    echo "Timed out waiting for cloud_sql_proxy connection"
    exit 1
  fi
  ((count=count+1))
done

serviceAccount="$ENV-$ORG-ad-hoc@$PROJECT_ID.iam"

export SA=$serviceAccount

# TODO @anhpngt: This is not really the canonical way for this
# I think it's better to use `bash -c` instead.
IFS=' ' read -r -a FILE <<< "$COMMAND"

# Grant permission for file
chmod 700 $FILE
# Run script file
echo "===================================================RESULT=================================================="
eval $COMMAND
exitcode=$?
echo "===================================================RESULT=================================================="
echo "Sending SIGTERM to cloud-sql-proxy process"
fuser -k -TERM 5432/tcp
exit $exitcode
