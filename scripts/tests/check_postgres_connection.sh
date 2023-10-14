#!/bin/bash
#
# This script checks whether "SQL_PROXY_CONN_NAME" and "DB_PREFIX" value is allowed
# for environemnt "ENV" and partner "ORG". It exits 1 if the check fails.
# - ENV can be: stag, uat, dorp, prod
# - ORG can be: manabie, jprep, aic, ga, renseikai, synersia, tokyo
#
# Usage:
#   ENV=stag ORG=manabie SQL_PROXY_CONN_NAME=<cloudsql-conn> DB_PREFIX=<prefix> ./check_postgres_connection.sh

set -euo pipefail

if [ -z "${ENV}" ]; then
  >&2 echo "Environment variable \"ENV\" cannot be empty"
  exit 1
fi
if [ -z "${ORG}" ]; then
  >&2 echo "Environment variable \"ORG\" cannot be empty"
  exit 1
fi
if [ -z "${SQL_PROXY_CONN_NAME}" ]; then
  >&2 echo "Environment variable \"SQL_PROXY_CONN_NAME\" cannot be empty"
  exit 1
fi
if [ -z "${DB_PREFIX+x}" ]; then
  >&2 echo "Environment variable \"DB_PREFIX\" must be set (it can be empty)"
  exit 1
fi

echo "ENV: \"${ENV}\""
echo "ORG: \"${ORG}\""
echo "SQL_PROXY_CONN_NAME: \"${SQL_PROXY_CONN_NAME}\""
echo "DB_PREFIX: \"${DB_PREFIX}\""

# Simple check to prevent mistakenly connecting to prod from preprod
if [[ "${ENV}" == "dorp" ]]; then
  if [[ "${SQL_PROXY_CONN_NAME}" != *"clone"* ]]; then
    >&2 echo "${SQL_PROXY_CONN_NAME} is an invalid Cloud SQL connection name for preproduction (should contain \"clone\" prefix)"
    exit 1
  fi
fi

vfp="${BASH_SOURCE%/*}/../../deployments/helm/backend/${ENV}-${ORG}-values.yaml"
if [ ! -f "${vfp}" ]; then
  >&2 echo "File ${vfp} does not exist"
  exit 1
fi

check_cloudsql_conn() {
  # This is the list of keys in helm value file that
  # contains the Cloud SQL connection strings. We want SQL_PROXY_CONN_NAME
  # to match one of the values from these keys.
  cloudSQLKeys=(
    "cloudSQLCommonInstance"
    "cloudSQLLMSInstance"
    "cloudSQLAuthInstance"
  )
  for k in "${cloudSQLKeys[@]}"; do
    conn=$(k=$k yq '.global.[env(k)]' "${vfp}")
    if [[ "${conn}" == "null" ]]; then
      >&2 echo "Field \"global.$k\" does not exist in ${vfp}"
      return 1
    fi
    echo "Found allowed Cloud SQL instance: ${conn}"
    if [[ "${SQL_PROXY_CONN_NAME}" == "${conn}" ]]; then
      echo "\"SQL_PROXY_CONN_NAME\" successfully matches value of field \"global.$k\""
      return 0
    fi
  done

  >&2 echo "\"SQL_PROXY_CONN_NAME\" does not match any known and allowed Cloud SQL instance"
  return 1
}

check_dbprefix() {
  v=$(yq '.global.dbPrefix' "${vfp}")
  if [[ "${v}" == "null" ]]; then
    >&2 echo "Field \"global.dbPrefix\" does not exist in ${vfp}"
    return 1
  fi
  echo "Found allowed db prefix: \"${v}\""
  if [[ "${DB_PREFIX}" != "${v}" ]]; then
    >&2 echo "DB_PREFIX is invalid (expected \"${v}\", got \"${DB_PREFIX})\""
    return 1
  fi
  echo "\"DB_PREFIX\" value is allowed"
  return 0
}

check_cloudsql_conn
check_dbprefix
