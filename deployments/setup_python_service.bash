#!/bin/bash

set -eu

install_aphelios(){
  echo "Installing Aphelios..."

  local namespace="$ENV-$ORG-machine-learning"
  svcCreds=${SVC_CRED:-}  # already available from env.bash
  helm upgrade --install aphelios --create-namespace -n "$namespace" ./deployments/helm/platforms/aphelios \
    --set=environment="$ENV" \
    --set=vendor="$ORG" \
    --set=serviceAccountEmailSuffix="$SERVICE_ACCOUNT_EMAIL_SUFFIX"
}

install_auto_scheduling_locally() {
  echo "Installing auto-scheduling..."

  current_date=$(date +%Y%m%d)
  current_time=$(date +%s%1N)
  result="$current_date$current_time"
  result=${result:0:10}

  docker build \
    -f ./internal/scheduling/job/bestco/scheduling.Dockerfile . \
    -t asia.gcr.io/student-coach-e1e95/scheduling:"$result"

  docker tag asia.gcr.io/student-coach-e1e95/scheduling:"$result" \
    localhost:5001/asia.gcr.io/student-coach-e1e95/scheduling:latest

  docker push localhost:5001/asia.gcr.io/student-coach-e1e95/scheduling:latest
}

install_auto_scheduling_locally
