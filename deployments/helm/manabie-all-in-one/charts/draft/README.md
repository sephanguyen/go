# Deploy local using Helm Charts

## File structure

```sh
draft
├── Chart.lock
├── charts
│   └── virtualservice-0.1.0.tgz 
├── Chart.yaml
├── secrets
│   └── manabie
│       └── local
│           ├── draft.config.yaml
│           ├── draft.secrets.yaml.encrypted
│           └── draft.secrets.yaml.encrypted.base64
├── templates
│   ├── configmap.yaml
│   ├── deployment.yaml    
│   ├── _helpers.tpl
│   ├── secret.yaml
│   ├── serviceaccount.yaml
│   ├── service.yaml
│   └── virtualservice.yaml
└── values.yaml
```

* ```.helmignore:``` This holds all the files to ignore when packaging the chart.
* The ```values.yaml```  file is also important to templates. This file contains the *default values* for a chart. These values may be overridden by users during `helm install` or `helm upgrade`.
* The ```templates``` directory is for template files. When Helm evaluates a chart, it will send all of the files in the `templates/` directory through the template rendering engine. It then collects the results of those templates and sends them on to Kubernetes. In service draft, templates contain:
  - ```configmap.yaml``` is the file used to create a ConfigMap. ConfigMaps are Kubernetes objects that allow to seperate configuration data/files from image content to keep containerized applications portable.
  - ```deployment.yaml``` is  basic manifest for creating a Kubernetes [deployment](https://kubernetes.io/docs/user-guide/deployments/). Deployment is a resource object in Kubernetes that provides declarative updates to applications. A deployment allows you to describe an application’s life cycle, such as which images to use for the app, the number of pods there should be, and the way in which they should be updated. 
  - ```service.yaml``` is a basic manifest for creating a [service endpoint](https://kubernetes.io/docs/user-guide/services/) for deployment.
  - ```_helpers.tpl```: A place to put template helpers that you can re-use throughout the chart.
  - ```secret.yaml``` is a file used to create a Secret. Kubernetes Secrets let you store and manage sensitive information, such as passwords, OAuth tokens, and ssh keys. Storing confidential information in a Secret is safer and more flexible than putting it verbatim in a [Pod](https://kubernetes.io/docs/concepts/workloads/pods/) definition or in a [container image](https://kubernetes.io/docs/reference/glossary/?all=true#term-image).
  - ```serviceaccount.yaml``` used to create Service Account for Pod. A service account provides an identity for processes that run in a Pod.
  - ```virtualservice.yaml``` defines a set of traffic routing rules to apply when a host is addressed. Each routing rule defines matching criteria for traffic of a specific protocol. If the traffic is matched, then it is sent to a named destination service (or subset/version of it) defined in the registry.
* The ```Chart.yaml``` file contains a description of the chart. You can access it from within a template. The `charts/` directory *may* contain other charts (which we call *subcharts*). Later in this guide we will see how those work when it comes to template rendering.
* The ```secrets``` directory contain encrypted configuration files.

## Quick start

* Run  ```./deployments/local.bash``` for start [Minikube](https://minikube.sigs.k8s.io/docs/start/), build image and set up istio, postgres ...

* Install helm charts of draft service:

```bash
helm upgrade draft --install ./deployments/helm/manabie-all-in-one/charts/draft \
    -n backend \
    --values ./deployments/helm/platforms/gateway/local-manabie-values.yaml \
    --set=project="local" \
    --set=environment="local" \
    --set=vendor="manabie"
```

* We will get the output:

```shell
Release "draft" has been upgraded. Happy Helming!
NAME: draft
LAST DEPLOYED: Fri Apr 16 14:47:38 2021
NAMESPACE: backend
STATUS: deployed
REVISION: 1
TEST SUITE: None
```
* Check pod status by following command:

```shell
kubectl get pods -n <namespace>
```

* If pod is not running , to find out why the pod is not running, we can use `kubectl describe pod` on the pending Pod:

```shell
kubectl describe pod <pod's name> -n <namespace>
```

