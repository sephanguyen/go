#!/bin/bash

# Reference https://kind.sigs.k8s.io/docs/user/local-registry/

set -e

# skip creating kind cluster if it already exists
if kind get clusters | grep '^kind$' >/dev/null; then
  echo "kind cluster already created. To re-create it, run \"kind delete cluster\" to delete the existing cluster first, then run this script again."
  RED='\033[0;31m'
  NC='\033[0m' # No Color

  if kubectl version | grep 'v1.23' >/dev/null; then
    printf "${RED}=======================${NC}\n"
    printf "${RED}Using deprecated kind version v1.23, please run ./deployments/sk.bash -d to delete cluster and reinstall kind again${NC}\n"
    printf "${RED}=======================${NC}\n"
  fi
  exit 0
fi

ci="${CI:-false}"
use_shared_registry="${USE_SHARED_REGISTRY:-false}"
use_hcmo_runner=false

# check using hcmo runners or not
if [[ $RUNNER_LABELS == *"arc-runner-hcm"* ]]; then
  echo "Using HCMO runner !!!"
  use_hcmo_runner=true
fi

# create a cluster with the local registry enabled in containerd
# explicitly `docker pull` to avoid image being cleaned up by `docker image prune`
kind_image="kindest/node:v1.24.12@sha256:1e12918b8bc3d4253bc08f640a231bb0d3b2c5a9b28aa3f2ca1aee93e1e8db16" # kind v0.18.0
if [[ "${ci}" == "true" ]]; then
  # On CI, use the cached image in Artifact Registry instead to save networking time and cost
  kind_image="asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci/kindest/node:v1.24.12-v0.18.0"
fi
docker pull "${kind_image}"
reg_name='kind-registry'
reg_port='5001'
reg_host='localhost:5001'
containerdConfigPatches=$(cat <<EOF
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."kind-registry:5000"]
    endpoint = ["http://${reg_name}:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."kind-reg.actions-runner-system.svc"]
    endpoint = ["http://${reg_name}:5000"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."asia-southeast1-docker.pkg.dev"]
    endpoint = ["http://${reg_name}:5000"]
EOF
)

function setup_local_registry() {
  if [ "$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)" != 'true' ]; then
    docker run \
      -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
      registry:2
  fi
}

# connect the registry to the cluster network if not already connected
function connect_registry_to_cluster() {
  if [ "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "${reg_name}")" = 'null' ]; then
    docker network connect "kind" "${reg_name}"
  fi
}

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
function setup_local_registry_hosting() {

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "${reg_host}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF

}

function setup_kind() {

  cat <<EOF | kind create cluster \
  --image ${kind_image} \
  --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
${extraMounts}
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  - |
    kind: ClusterConfiguration
    etcd:
      local:
        dataDir: /tmp/etcd
        extraArgs:
          unsafe-no-fsync: "true"
  - |
    kind: KubeletConfiguration
    maxPods: 510
  extraPortMappings:
  - containerPort: 31400
    hostPort: 31400
    protocol: TCP
  - containerPort: 31500
    hostPort: 31500
    protocol: TCP
  - containerPort: 31600
    hostPort: 31600
    protocol: TCP
${containerdConfigPatches}
EOF
}

function remove_kind_cluster() {
  kind delete cluster
}

# case 1: ci=false
if [[ "$ci" == "false" ]]; then

setup_local_registry
setup_kind
connect_registry_to_cluster
setup_local_registry_hosting

else

  # case 2: ci=true, use_shared_registry=true
  if [[ "$use_shared_registry" == "true" ]]; then

reg_name='kind-reg.actions-runner-system.svc'
reg_port='443'
reg_host='kind-reg.actions-runner-system.svc'
containerdConfigPatches=$(cat <<EOF
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."kind-reg.actions-runner-system.svc"]
    endpoint = ["https://${reg_name}"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."kind-reg.actions-runner-system.svc".tls]
      cert_file = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/tls.cert"
      key_file  = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/tls.key"
      ca_file = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/ca.crt"
EOF
)

extraMounts=$(cat <<EOF
  extraMounts:
    - containerPath: /etc/docker/certs.d/kind-reg.actions-runner-system.svc
      hostPath: /etc/docker/certs.d/kind-reg.actions-runner-system.svc
EOF
)

if [[ "$use_hcmo_runner" == "true" ]]; then

containerdConfigPatches=$(cat <<EOF
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."asia-southeast1-docker.pkg.dev"]
    endpoint = ["https://pull-through-registry.actions-runner-system.svc"]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."kind-reg.actions-runner-system.svc"]
    endpoint = ["https://${reg_name}"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."kind-reg.actions-runner-system.svc".tls]
      cert_file = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/tls.cert"
      key_file  = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/tls.key"
      ca_file = "/etc/docker/certs.d/kind-reg.actions-runner-system.svc/ca.crt"
  [plugins."io.containerd.grpc.v1.cri".registry.configs."pull-through-registry.actions-runner-system.svc".tls]
      cert_file = "/etc/docker/certs.d/pull-through-registry.actions-runner-system.svc/tls.cert"
      key_file  = "/etc/docker/certs.d/pull-through-registry.actions-runner-system.svc/tls.key"
      ca_file = "/etc/docker/certs.d/pull-through-registry.actions-runner-system.svc/ca.crt"
EOF
)

extraMounts=$(cat <<EOF
  extraMounts:
    - containerPath: /etc/docker/certs.d/kind-reg.actions-runner-system.svc
      hostPath: /etc/docker/certs.d/kind-reg.actions-runner-system.svc
    - containerPath: /etc/docker/certs.d/pull-through-registry.actions-runner-system.svc
      hostPath: /etc/docker/certs.d/pull-through-registry.actions-runner-system.svc
    - containerPath: /var/lib/containerd
      hostPath: /var/lib/containerd
EOF
)

fi

remove_kind_cluster
setup_kind
setup_local_registry_hosting

  # case 3: ci=true, use_shared_registry=false
  else

remove_kind_cluster
setup_local_registry
setup_kind
connect_registry_to_cluster
setup_local_registry_hosting

  fi

fi
