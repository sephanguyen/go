#!/bin/bash

set -e

install_prometheus() {
  cmd="helm upgrade --install --create-namespace --wait -n monitoring --timeout 2m30s \
    prometheus ./deployments/helm/platforms/monitoring/prometheus/prometheus-15.10.4.tgz \
    --values ./deployments/helm/platforms/monitoring/prometheus/values.yaml"

  if [[ "$ENV" == "stag" ]]; then
    cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/stag-values.yaml"
  fi

  if [[ "$ENV" == "uat" ]]; then
    cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/uat-values.yaml"
  fi

  if [[ "$ENV" == "prod" ]]; then
    cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/prod-values.yaml"
    if [[ "$ORG" == "manabie" ]]; then
      cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/prod-manabie-values.yaml"
    elif [[ "$ORG" == "jprep" ]]; then
      cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/prod-jprep-values.yaml"
    elif [[ "$ORG" == "tokyo" ]]; then
      cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/prod-tokyo-values.yaml"
    else # jp-partners cluster
      cmd="${cmd} --values ./deployments/helm/platforms/monitoring/prometheus/jp-partners-values.yaml"
    fi
  fi

  $cmd
}

install_kiali() {
  helm upgrade --install --create-namespace  --wait -n istio-system --timeout 2m30s \
    --set=environment="$ENV" \
    --set=vendor="$ORG" \
    kiali ./deployments/helm/platforms/monitoring/kiali-server
}

install_grafana() {
  cmd="helm upgrade --install --create-namespace --wait -n monitoring --timeout 3m30s \
    grafana ./deployments/helm/platforms/monitoring/grafana"

  if [[ "$ENV" == "prod" ]]; then
    cmd="${cmd} --values ./deployments/helm/platforms/monitoring/grafana/production-values.yaml"
  fi

  $cmd
}

install_jaeger() {
  cmd="helm upgrade --install --create-namespace -n monitoring jaeger ./deployments/helm/platforms/monitoring/jaeger"

  if [[ "$ENV" == "prod" ]]; then
    cmd="${cmd} --set cassandra.persistence.enabled=true --set cassandra.persistence.storageClass=premium-rwo"
  fi

  $cmd
}

install_jaeger_all_in_one() {
  helm upgrade --install --create-namespace -n monitoring \
    jaeger-all-in-one ./deployments/helm/platforms/monitoring/jaeger-all-in-one
}

install_yugabyte() {
  local defaultKmsPath="projects/dev-manabie-online/locations/global/keyRings/deployments/cryptoKeys/github-actions"
  local kmsPath=${KMS_PATH:-$defaultKmsPath}

  local svcCreds=""
  local createBackendDbs=false
  local storageClass=standard
  local prodValues=""

  if [[ "$ENV" == "local" ]]; then
    createBackendDbs=true
    svcCreds=$(gpg --quiet --batch --yes --decrypt --passphrase="$DECRYPT_KEY" ./deployments/helm/platforms/yugabyte/secrets/$ORG/$ENV/yugabyte_sa.json.base64.gpg)
  else
    prodValues="--set resource.master.requests.cpu=1,\
      resource.master.requests.memory=1Gi,\
      resource.master.limits.cpu=1,\
      resource.master.limits.memory=1Gi,\
      resource.tserver.requests.cpu=8,\
      resource.tserver.requests.memory=9Gi,\
      resource.tserver.limits.cpu=8,\
      resource.tserver.limits.memory=9Gi,\
      storage.tserver.size=256Gi,\
      storage.tserver.storageClass=premium-rwo"
  fi

  echo $prodValues

  helm upgrade --install --create-namespace --wait -n yugabyte --timeout 2m30s \
    yugabyte ./deployments/helm/platforms/yugabyte $prodValues \
    --set vendor=$ORG \
    --set environment=$ENV \
    --set createBackendDbs.enabled="$createBackendDbs" \
    --set configs.kmsPath="$kmsPath" \
    --set secrets.serviceCredential="$svcCreds" \
    --set serviceAccountEmailSuffix="$SERVICE_ACCOUNT_EMAIL_SUFFIX"
}

install_opentelemetry_collector() {
  helm upgrade --install --create-namespace -n monitoring opentelemetry-collector \
    --values ./deployments/helm/platforms/monitoring/opentelemetry-collector/values.yaml \
    ./deployments/helm/platforms/monitoring/opentelemetry-collector/opentelemetry-collector-0.7.0.tgz
}

install_redash() {
  local redashNamespace="redash"


  if [[ "$ENV" == "local" ]]; then
    svcCreds=$(gpg --quiet --batch --yes --decrypt --passphrase="$DECRYPT_KEY" ./developments/serviceaccount.json.base64.gpg)
  fi

  if [[ "$ENV" == "prod" ]]; then
    redashNamespace="prod-analytics-services"
  fi

  helm upgrade --install --timeout 10m30s --wait --create-namespace --namespace $redashNamespace redash ./deployments/helm/platforms/redash \
    --values ./deployments/helm/platforms/redash/values.yaml \
    --values ./deployments/helm/platforms/redash/${ENV}-${ORG}-values.yaml \
    --values ./deployments/helm/platforms/gateway/${ENV}-${ORG}-values.yaml \
    --set=environment=$ENV \
    --set=vendor=$ORG \
    --set=serviceAccountEmailSuffix="$SERVICE_ACCOUNT_EMAIL_SUFFIX" \
    --set=secrets.serviceCredential="$svcCreds"
}

install_runner_controller() {
  echo "Installing actions runner controller !!!"

  local runnerNamespace="actions-runner-system"

  helm upgrade --install --timeout 5m --wait --create-namespace \
    --namespace $runnerNamespace actions-runner-controller ./deployments/helm/platforms/actions-runner-controller/actions-runner-controller-0.23.3.tgz \
    --values ./deployments/helm/platforms/actions-runner-controller/values.yaml
}

uninstall_ml_model() {
  echo "Delete machine learning model server ..."

  local model_namespace="$ENV-$ORG-machine-learning"
  kubectl delete inferenceservice bubble -n $model_namespace
  kubectl delete inferenceservice question-field -n $model_namespace
  kubectl delete inferenceservice answer-sheet -n $model_namespace
  kubectl delete inferenceservice ocr -n $model_namespace
}

install_actions_exporter() {
  echo "Installing actions exporter !!!"

  local runnerNamespace="actions-runner-system"

  helm upgrade --install --timeout 5m --wait --create-namespace \
    --namespace $runnerNamespace github-actions-exporter ./deployments/helm/platforms/github-actions-exporter/github-actions-exporter-0.1.4.tgz \
    --values ./deployments/helm/platforms/github-actions-exporter/values.yaml
}
