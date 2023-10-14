#!/bin/bash

# This file sets up the Docker image caching for running in local and CI.
#
# Each and every image used in Kubernetes deployment should be added to this file,
# so as to speed up the deployment in local and save networking cost on CI.
# Your helm chart's local profile (e.g. local-manabie-values.yaml) should also
# customize the deployments to pull from local registry.
#
# In production, images are always pulled from upstream source.
#
# LOCAL: In local, all images are cached in a local registry:
#
# ```
#   $ docker ps
#   CONTAINER ID   IMAGE        COMMAND                  CREATED        STATUS       PORTS                      NAMES
#   30451533defc   registry:2   "/entrypoint.sh /etc…"   14 hours ago   Up 4 hours   127.0.0.1:5001->5000/tcp   kind-registry
# ```
# Check `./deployments/kind_with_registry.bash` to see how the registry is set up.
#
# You can push/pull images to this registry with `docker pull localhost:5001/<image>`.
# This script pulls all the required images and re-pushes them to the local registry.
#
# ```
#   $ docker pull postgres:13.1 # pull the original image
#   $ docker tag postgres:13.1 localhost:5001/postgres:13.1
#   $ docker push localhost:5001/postgres:13.1  # push the image to local registry
# ```
# Then, inside local k8s cluster, containers can use those images without having to pull them again.
#
#
# CI: On CI, the environment is different than in local. Thus, the caching
# strategy is different. Depending on the size and usage frequency of the image,
# we cache the image in 1 out of 2 locations:
#
#   1) A shared registry running in the same k8s cluster.
#   Note that each CI runs in a k8s pod in `staging-2` cluster. Thus, the registry
#   can be accessed via a normal k8s service: kind-reg.actions-runner-system.svc.
#   Thus, use `kind-reg.actions-runner-system.svc/<image>` in your deployments.
#   This script automatically updates CI's registry with new images.
#   You can use this option by default.
#
#   2) Google's Artifact Registry (GAR). Images are stored in the same
#   region (asia-southeast1), so pulling from GAR incurs no additional networking cost.
#   In this case, use `asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci/<image>` in
#   your deployments. However, this script does not automatically push to GAR.
#   You will need to do that yourselves.
#   You can consider this option if the image is used in a lot of deployments,
#   or if the image's size is large.
#   
#   3) Pull through registry running in k8s hcmo local cluster. Pull through registry acts
#   a proxy between GAR and runners. Runners pull images from the pull through registry.
#   If pull through registry doesn't have the image. It will pull from GAR and cached the image.
#   Internet bandwidth in HCMO is so limited. Pull through registry helps us reduce internet traffic
#   and ensure runners don't need to wait too long to pull images.
#
# To keep the k8s manifests of local and CI environment as identical as possible,
# in local, the registry domains `kind-reg.actions-runner-system.svc` and
# `asia-southeast1-docker.pkg.dev` are redirected to `localhost:5001`.
# Therefore, choose 1 option from above and do not use `localhost:5001`.
set -u

ci=${CI:-false}
use_shared_registry="${USE_SHARED_REGISTRY:-false}"
aphelios_deployment_enabled=${APHELIOS_DEPLOYMENT_ENABLED:-false}
scheduling_deployment_enabled=${SCHEDULING_DEPLOYMENT_ENABLED:-false}
redash_deployment_enabled=${REDASH_DEPLOYMENT_ENABLED:-false}
monitoring_deployment_enabled=${INSTALL_MONITORING_STACKS:-false}
MANABIE_DEPLOYER_ENABLED=${MANABIE_DEPLOYER_ENABLED:-false}
appsmith_deployment_enabled=${APPSMITH_DEPLOYMENT_ENABLED:-false}

# List of images to be cached in local (or with shared registry on CI).
# To cache a new image, simply add it to this `required_images` variable.
# Reference to the cached image with `kind-reg.actions-runner-system.svc/<image>`.
#
# Check https://kubernetes.io/docs/tasks/access-application-cluster/list-all-running-container-images/
# to get the list of images being used in a minikube cluster.
required_images=(
  asia.gcr.io/student-coach-e1e95/decrypt-secret:20220219
  asia.gcr.io/student-coach-e1e95/decrypt-secret:20220517
  asia.gcr.io/student-coach-e1e95/fake_firebase_token:1.1.0 # for emulators' firebase
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2-20230411
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v2.8.1.cli-migrations-v3
  istio/proxyv2:1.18.0 # for istio
  istio/pilot:1.18.0   # for istio
  letsencrypt/pebble:v2.3.1      # for emulators' letsencrypt
  minio/mc:RELEASE.2020-12-18T10-53-53Z
  minio/minio:RELEASE.2020-12-23T02-24-12Z
  postgres:13.11-bookworm   # for emulators' postgres
  mozilla/sops:v3.7.3-alpine # for sops decrypt
  nats:2.8.4-alpine3.15
  natsio/nats-box:0.13.8
  natsio/prometheus-nats-exporter:0.9.3
  natsio/nats-server-config-reloader:0.7.0
  asia.gcr.io/student-coach-e1e95/graphql-mesh:0.0.1
  asia.gcr.io/student-coach-e1e95/wait-for:0.0.2
  asia.gcr.io/student-coach-e1e95/hasura-metric-adapter:0.0.1
  redis:7.2.0-alpine3.18

  asia.gcr.io/student-coach-e1e95/customized_debezium_kafka:1.9.0 # for kafka
  asia.gcr.io/student-coach-e1e95/customized_debezium_connect:1.9.6
  asia.gcr.io/student-coach-e1e95/customized_cp_schema_registry:7.1.2
  asia.gcr.io/student-coach-e1e95/kafkatools:0.0.2
  provectuslabs/kafka-ui:latest
  danielqsj/kafka-exporter:latest

  # import-map-deployer
  asia.gcr.io/student-coach-e1e95/import-map-deployer:0.0.2
  google/cloud-sdk:323.0.0-alpine

  # elastic
  amazon/opendistro-for-elasticsearch-kibana:1.13.1
  asia.gcr.io/student-coach-e1e95/customized_elastic:1.13.1
  quay.io/prometheuscommunity/elasticsearch-exporter:v1.2.1

  # unleash
  unleashorg/unleash-server:4.19.1
  unleashorg/unleash-proxy:0.13.1
  node:14-alpine

  # cert-manager
  quay.io/jetstack/cert-manager-acmesolver:v1.7.1
  quay.io/jetstack/cert-manager-cainjector:v1.7.1
  quay.io/jetstack/cert-manager-controller:v1.7.1
  quay.io/jetstack/cert-manager-ctl:v1.7.1
  quay.io/jetstack/cert-manager-webhook:v1.7.1

  # ksql
  jbergknoff/postgresql-client:latest # this image only has "latest" tag
  confluentinc/cp-ksqldb-server:7.3.0
  confluentinc/ksqldb-cli:0.28.2
  confluentinc/ksqldb-server:0.28.2
  asia-docker.pkg.dev/student-coach-e1e95/manaverse/kafka-cronjob-restart-connector:2023081001 # for kafka
)

# appsmith
if [[ "${appsmith_deployment_enabled}" == "true" ]]; then
  required_images+=(
    asia.gcr.io/student-coach-e1e95/appsmith-custom:1.0.25
    asia.gcr.io/student-coach-e1e95/mongodb-custom:0.0.1
  )
fi

if [[ "${scheduling_deployment_enabled}" == "true" ]]; then
  required_images+=(
    asia-docker.pkg.dev/student-coach-e1e95/manaverse/auto-scheduling-grpc:2023062103
    asia-docker.pkg.dev/student-coach-e1e95/manaverse/auto-scheduling-http:2023062102
    asia-docker.pkg.dev/student-coach-e1e95/manaverse/auto-scheduling-job:2023062100
  )
fi

if [[ "${aphelios_deployment_enabled}" == "true" ]]; then
  required_images+=(
    asia.gcr.io/student-coach-e1e95/aphelios:latest
    gcr.io/kubebuilder/kube-rbac-proxy:v0.4.0
    kserve/kserve-controller:v0.9.0
    kserve/storage-initializer:v0.9.0
    docker.io/seldonio/mlserver:1.0.0
  )
fi

if [[ "${redash_deployment_enabled}" == "true" ]]; then
  required_images+=(
    asia.gcr.io/student-coach-e1e95/customized_redash:10.1.1
  )
fi

if [[ "${monitoring_deployment_enabled}" == "true" ]]; then
  required_images+=(
    docker.io/bitnami/minio:2022.7.8-debian-11-r0
    docker.io/bitnami/thanos:0.27.0-scratch-r3
    quay.io/prometheus/prometheus:v2.36.2
    grafana/grafana:9.0.5
  )
fi
if [[ "${CAMEL_K_ENABLED:-false}" == "true" ]]; then
  required_images+=(
    apache/camel-k:1.12.0
    eclipse-temurin:11.0.19_7-jdk-jammy
  )
fi
if [[ "${NETWORK_POLICY_ENABLED:-false}" == "true" ]]; then
  required_images+=(
    "quay.io/tigera/operator:v1.30.4"

    # Installation
    "calico/node:v3.26.1"
    "calico/pod2daemon-flexvol:v3.26.1"
    "calico/cni:v3.26.1"
    "calico/kube-controllers:v3.26.1"
    "calico/csi:v3.26.1"
    "calico/node-driver-registrar:v3.26.1"
    "calico/typha:v3.26.1"

    # Calico API server
    "calico/apiserver:v3.26.1"

    # Used in integration testing
    "alpine:3.18.2"
    "nginx:1.25.1-alpine3.17-slim"
  )
fi

# shellcheck source=../scripts/log.sh
source ./scripts/log.sh

for img in "${required_images[@]}"; do
  image_name=${img%:*}
  image_tag=${img##*:}

  if [[ "${use_shared_registry}" == "false" ]]; then

    if [[ "$ci" == "false" ]]; then
      if curl -fs "http://localhost:5001/v2/${image_name}/manifests/${image_tag}" >/dev/null; then
        logdebug "Image \"localhost:5001/${img}\" found in local registry"
        continue
      fi
    else
      if docker image inspect "localhost:5001/$img" >/dev/null 2>&1 >/dev/null; then
        logdebug "Image \"localhost:5001/${img}\" found in local registry"
        continue
      fi
    fi

    loginfo "Image \"localhost:5001/${img}\" cannot be found in local registry. Pulling it now."
    docker pull "${img}"
    docker tag "${img}" "localhost:5001/${img}"
    docker push "localhost:5001/${img}"
    loginfo ""

    if [[ "$ci" == "true" ]]; then
      docker rmi "${img}"
    fi

  else
    echo $image_name:$image_tag
    if curl -kfs "https://kind-reg.actions-runner-system.svc/v2/${image_name}/manifests/${image_tag}" >/dev/null; then
      logdebug "Image \"kind-reg.actions-runner-system.svc/${img}\" found in local registry"
      continue
    else
      loginfo "Image \"kind-reg.actions-runner-system.svc/${img}\" cannot be found in local registry. Pulling it now."
      docker pull "${img}"
      docker tag "${img}" "kind-reg.actions-runner-system.svc/${img}"
      docker push "kind-reg.actions-runner-system.svc/${img}"
      loginfo ""
    fi
  fi
done

# List of the images that are stored on Artifact Registry (AR).
# In local, we use a mirror registry config to redirect `docker pull` to localhost:5001.
# On CI, these images are downloaded directly from AR.
#
# To add a new image:
#   1) add the image name to this `artifact_registry_images` variable
#   2) retag and push the image to Artifact Registry, for example:
#   ```
#     $ docker pull postgres:13.1
#     $ docker tag postgres:13.1 asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci/postgres:13.1
#     $ docker push asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci/postgres:13.1
#   ```
# Reference to the cached image with `asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci/<image>`.
artifact_registry_images=(
  asia.gcr.io/student-coach-e1e95/wait-for:0.0.2

  # hasura
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v1.3.3.cli-migrations-v2-20230411
  asia.gcr.io/student-coach-e1e95/customized-graphql-engine:v2.8.1.cli-migrations-v3

  # istio
  istio/proxyv2:1.18.0
  istio/pilot:1.18.0

  nats:2.8.4-alpine3.15                                               # nats-jetstream
  asia.gcr.io/student-coach-e1e95/decrypt-secret:20220219             # elastic
  asia.gcr.io/student-coach-e1e95/decrypt-secret:20220517             # elastic
  asia.gcr.io/student-coach-e1e95/customized_elastic:1.13.1           # elastic
  asia.gcr.io/student-coach-e1e95/customized_debezium_kafka:1.9.0     # for kafka
  asia.gcr.io/student-coach-e1e95/customized_debezium_connect:1.9.6   # for kafka
  asia.gcr.io/student-coach-e1e95/customized_cp_schema_registry:7.1.2 # for kafka
  confluentinc/cp-ksqldb-server:7.3.0                                 # for kafka
  confluentinc/ksqldb-server:0.28.2                                   # for kafka dwh
  asia-docker.pkg.dev/student-coach-e1e95/manaverse/kafka-cronjob-restart-connector:2023081001 # for kafka

  mozilla/sops:v3.7.3-alpine # for sops decrypt
)
if [[ "${CAMEL_K_ENABLED:-false}" == "true" ]]; then
  artifact_registry_images+=(
    apache/camel-k:1.12.0
  )
fi

if [[ "$ci" != "true" ]]; then # cache in local only
  ar_image_prefix="student-coach-e1e95/ci"
  for img in "${artifact_registry_images[@]}"; do
    image_name="${ar_image_prefix}/${img%:*}"
    image_tag="${img##*:}"
    if curl -fs "http://localhost:5001/v2/${image_name}/manifests/${image_tag}" >/dev/null; then
      logdebug "Image \"localhost:5001/${img}\" found in local registry"
      continue # skip if already exists
    fi

    renamed_img="localhost:5001/${ar_image_prefix}/${img}"
    loginfo "Image \"${renamed_img}\" cannot be found in local registry. Pulling it now."
    docker pull "${img}"
    docker tag "${img}" "${renamed_img}"
    docker push "${renamed_img}"
    loginfo ""

    if [[ "$ci" == "true" ]]; then
      docker rmi "${img}"
    fi
  done
fi

# On CI, we use imagePullSecret as the credential to directly pull image from
# Google's Artifact Registry (AR).
# To make this work, you need to:
#   1. Make sure `regcred` docker-registry secret is created in your desired namespace
#      (by adding your namespace to the list below).
#   2. Set your pod's imagePullSecrets to: [{"name": "regcred"}]
#
# Note that the image upload to AR should be done manually.
if [[ "$ci" == "true" ]]; then
  flags=(
    "docker-registry"
    "regcred"
    "--docker-server" "${DOCKER_SERVER}"
    "--docker-username" "${DOCKER_USERNAME}"
    "--docker-password" "${DOCKER_PASSWORD}"
    "--docker-email" "${DOCKER_EMAIL}"
  )
  org=${ORG:-manabie}
  namespaces=(
    "camel-k"
    "tigera-operator"
    "emulator"
    "istio-system"
    "local-${org}-elastic"
    "local-${org}-appsmith"
    "local-${org}-backend"
    "local-${org}-data-warehouse"
    "local-${org}-kafka"
    "local-${org}-nats-jetstream"
    "local-${org}-unleash"
    "backend"
  )
  for ns in "${namespaces[@]}"; do
    kubectl create namespace "${ns}" --dry-run=client -o yaml | kubectl apply -f - # create namespace if not exists
    kubectl -n "${ns}" create secret "${flags[@]}"
  done
fi
