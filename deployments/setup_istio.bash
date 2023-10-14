#!/bin/bash

set -euo pipefail

kubecontext="kind-kind"

ROOT=~/.manabie
if [[ ! -d $ROOT ]]; then
  mkdir -p $ROOT
fi

ISTIO_VERSION=${ISTIO_VERSION:-1.15.4}
ISTIO_VERSION_NAME=${ISTIO_VERSION//./-}

install_istio() {
  ISTIO_DIR=$ROOT/istio-$ISTIO_VERSION
  if [[ ! -d $ISTIO_DIR ]]; then
    # run in subshell to keep the current directory
    (cd $ROOT; curl -L https://istio.io/downloadIstio | ISTIO_VERSION=$ISTIO_VERSION sh -)
  fi

  istioctl=$ISTIO_DIR/bin/istioctl
  $istioctl x precheck

  rev=$(kubectl -n istio-system get pod -l app=istio-ingressgateway -o "jsonpath={.items..metadata.labels['istio\.io/rev']}")
  if [[ "$rev" != "${ISTIO_VERSION//./-}" ]]; then
    echo "Installing istio version $ISTIO_VERSION"
    $istioctl install -y -f ./deployments/istio/$ENV/config-${ISTIO_VERSION_NAME}.yaml \
      --set revision=${ISTIO_VERSION_NAME} \
      --set values.global.imagePullPolicy=IfNotPresent

    $istioctl tag set local --revision $ISTIO_VERSION_NAME --overwrite
  fi
}

install_istio_local() {
  local ENV=${ENV:-local}
  if [[ "${ENV}" != "local" ]]; then
    echo >&2 "This function is supposed to run in only local environment (currently: $ENV)"
    exit 1
  fi

  ISTIO_DIR=$ROOT/istio-$ISTIO_VERSION
  if [[ ! -d $ISTIO_DIR ]]; then
    # run in subshell to keep the current directory
    (cd $ROOT; curl -L https://istio.io/downloadIstio | ISTIO_VERSION=$ISTIO_VERSION sh -)
  fi

  istioctl=$ISTIO_DIR/bin/istioctl
  $istioctl --context=${kubecontext} x precheck

  artifact_registry_domain="asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci"
  rev=$(kubectl --context=${kubecontext} -n istio-system get pod -l app=istio-ingressgateway -o "jsonpath={.items..metadata.labels['istio\.io/rev']}")
  if [[ "$rev" != "${ISTIO_VERSION//./-}" ]]; then
    extra_args="--set=hub=${artifact_registry_domain}/docker.io/istio" # in local, we pull images from local registry
    echo "Installing istio version $ISTIO_VERSION"
    $istioctl install -y -f "./deployments/istio/$ENV/config-${ISTIO_VERSION_NAME}.yaml" \
      --context=${kubecontext} \
      --set=revision="${ISTIO_VERSION_NAME}" \
      --set=values.global.imagePullPolicy=IfNotPresent \
      --set=values.global.imagePullSecrets[0]=regcred ${extra_args}

    $istioctl --context=${kubecontext} tag set local --revision "$ISTIO_VERSION_NAME" --overwrite
  fi
}

setup_backend_namespace() {
  target_namespace="$1"
  if ! kubectl --context=${kubecontext} get namespace "${target_namespace}" >/dev/null; then
    kubectl --context=${kubecontext} create namespace "${target_namespace}"
  fi

  istio_rev="$(kubectl --context=${kubecontext} get namespace "${target_namespace}" -o "jsonpath={.metadata.labels['istio\.io/rev']}")"
  if [[ -z $istio_rev ]]; then
    kubectl --context=${kubecontext} label namespace "${target_namespace}" istio.io/rev=local
  elif [[ "$istio_rev" != local ]]; then
    kubectl --context=${kubecontext} label namespace "${target_namespace}" istio.io/rev-
    kubectl --context=${kubecontext} label namespace "${target_namespace}" istio.io/rev=local
  else
    echo "==== namespace ${target_namespace} already has istio.io/rev=$istio_rev"
  fi
}

install_cert_manager() {
  local version="v1.7.1"
  if [[ "$(helm search repo jetstack/cert-manager --versions | grep -c $version)" != "1" ]]; then
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
  fi

  helm upgrade --install --create-namespace cert-manager jetstack/cert-manager \
    --wait \
    --namespace cert-manager \
    --version $version \
    --set 'cainjector.extraArgs[0]=--leader-elect=false' \
    --set installCRDs=true
}

install_istio_gateway() {
  helm upgrade --wait -n istio-system --install --timeout 1m30s \
    ${ENV}-${ORG}-gateway ./deployments/helm/platforms/gateway \
    --values ./deployments/helm/platforms/gateway/${ENV}-${ORG}-values.yaml \
    --set=org=${ORG} \
    --set=environment=${ENV}
}

setup_localhost() {
  # I feel bad for copy and pasting this block from below.
  # But it would be more dangerous to refactor it, as we will eventually remove minikube.
  is_host_configured=$(grep -c -E "^127.0.0.1\s+api.local-green.manabie.io" /etc/hosts) || true
  if [[ "$is_host_configured" == "0" ]]; then
    echo "  Add entry '127.0.0.1 api.local-green.manabie.io' to your /etc/hosts.";
  fi

  count=$(grep -cE '^[a-z0-9\.]+\s+api.local-green.manabie.io' /etc/hosts) || true
  if [ "$count" -gt "1" ]; then
    echo "Found more than one entry for api.local-green.manabie.io in /etc/hosts"
    echo "Please remove all except the '127.0.0.1 api.local-green.manabie.io' then run this script again."
    return 1
  fi
}

update_coredns() {
  ip=$(kubectl --context=${kubecontext} get svc istio-ingressgateway -n istio-system --no-headers | awk '{print$3}')

  echo "Updating coredns..."

  cat <<EOF | kubectl --context=${kubecontext} apply -f -
apiVersion: v1
data:
  Corefile: |
    .:53 {
        errors
        health {
           lameduck 5s
        }
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           fallthrough in-addr.arpa ip6.arpa
           ttl 30
        }
        prometheus :9153
        forward . 1.1.1.1 /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
    teacher.local.manabie.io web-api.local-green.manabie.io api.local-green.manabie.io backoffice.local.manabie.io backoffice-mfe.local.manabie.io admin.local-green.manabie.io {
       hosts {
         $ip teacher.local.manabie.io web-api.local-green.manabie.io api.local-green.manabie.io backoffice.local.manabie.io backoffice-mfe.local.manabie.io admin.local-green.manabie.io
         fallthrough
       }
       whoami
    }
    learner.local.manabie.io {
      hosts {
         $ip learner.local.manabie.io
         fallthrough
      }
      whoami
    }
    web-api.local-blue.manabie.io api.local-blue.manabie.io admin.local-blue.manabie.io {
       hosts {
         $ip web-api.local-blue.manabie.io api.local-blue.manabie.io admin.local-blue.manabie.io
         fallthrough
       }
       whoami
    }
    grafana.local.manabie.io {
       hosts {
         $ip grafana.local.manabie.io
         fallthrough
       }
       whoami
    }
    kiali.local.manabie.io {
       hosts {
         $ip kiali.local.manabie.io
         fallthrough
       }
       whoami
    }
    redash.local.manabie.io {
       hosts {
         $ip redash.local.manabie.io
         fallthrough
       }
       whoami
    }
    appsmith.local-green.manabie.io {
       hosts {
         $ip appsmith.local-green.manabie.io
         fallthrough
       }
       whoami
    }
    minio.local.manabie.io {
       hosts {
         $ip minio.local.manabie.io
         fallthrough
       }
       whoami
    }
    learnosity-web-view.local.manabie.io {
       hosts {
         $ip learnosity-web-view.local.manabie.io
         fallthrough
       }
       whoami
    }
    internal.local.manabie.io internal.local-green.manabie.io {
      hosts {
        $ip internal.local.manabie.io internal.local-green.manabie.io
        fallthrough
      }
      whoami
    }
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
EOF
  kubectl --context=${kubecontext} delete pod -n kube-system --wait $(kubectl --context=${kubecontext} get pods -n kube-system | grep coredns | awk '{print$1}')
}

wait_for_cert() {
  acme=$(kubectl --context=${kubecontext} get pods -n istio-system | grep -c cm-acme-http-solver) || true
  echo "Waiting for acme http solver to complete..."
  while [[ "$acme" != "0" ]]; do
    sleep 2
    acme=$(kubectl --context=${kubecontext} get pods -n istio-system | grep -c cm-acme-http-solver) || true
  done
}
