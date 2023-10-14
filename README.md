## Overview

## Setup local environment using kind

<!-- Update this documentation whenever there's a change in setting up backend

  https://manabie.atlassian.net/wiki/spaces/TECH/pages/429655074/Required+Setup+BE -->

### Requirements

- Install [Docker Engine](https://docs.docker.com/engine/install/ubuntu/)
  - DO NOT install *Docker Desktop*
  - Remember to [add Docker as a non-root user](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) when using Linux)
- Make sure your email belong to `dev@manabie.com`, `tech-func-<yourfunction>@manabie.com`, and `tech-squad-<yoursquad>@manabie.com` [groups](https://groups.google.com/my-groups) (by asking your team lead), then [install gcloud CLI](https://cloud.google.com/sdk/docs) and run `gcloud init` to login.
  - Choose `dev-manabie-online` when asked for a cloud project to use (this option does not matter)
  - Pick `No` when asked to configure a default Compute Region and Zone
  - Run `gcloud auth configure-docker`
- Install [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
  - It is recommended to [enable shell autocompletion for kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/#enable-shell-autocompletion)
- Install [helm](https://helm.sh/docs/intro/install/)
  - Similar to `kubectl`, it is recommended to [enable shell autocompletion for helm](https://helm.sh/docs/helm/helm_completion_bash/)
- Install `make`: `sudo apt install build-essential`
- Install [go](https://golang.org/dl/), version at least `1.20.5`
- Add `~/.manabie/bin` to your `PATH`. Persist this change by adding `export PATH=$PATH:~/.manabie/bin` to your `.bashrc`
- You can also manually cache large docker images at home first (automatically done when running `deployments/sk.bash`).
This needs to be done only once.


### Start the local cluster and run tests

```sh
# Init the current repository (this is done once per repository)
make init

# Start the cluster first time. This command will run the entire backend cluster inside kind.
./deployments/sk.bash

# Modify some code, then rebuild and restart the services
./deployments/sk.bash -s bob,tom

# Run the test after the service, or gandalf itself, has restarted
./deployments/k8s_bdd_test.bash
```

You can also manually cache large docker images at home first (automatically done when running `deployments/sk.bash`).


### Alerting

This repository defines two types of alerts. 

- [Prometheus Alerts](/deployments/helm/platforms/monitoring/prometheus/ALERTS.md)
- [GCS Alerts](/deployments/terraform/modules/alert-policies/ALERTS.md)

### Metrics

- [Grafana on Staging](https://grafana.staging.manabie.io/)
- Prometheus on Staging

    ```bash
    gcloud container clusters get-credentials staging --zone asia-southeast1 --project staging-manabie-online # auth staging
    
    kubectl -n monitoring port-forward svc/prometheus-server 8082:80 # port forward to prometheus

    # browse to http://localhost:8082/
    ```

### Working with local cluster as a client (web, mobile)

Run `curl -kv https://api.local-green.manabie.io:31500/image/topic/1231231qwe` and you should get `404` result

### Elasticsearch

To sync (init) documents from database when create new index elastic search, for example `conversation`:

```sh
go run cmd/utils/main.go init conversation --tomdb tomdb_addr --yasuo ysAddr --maxdoc max_record_per_turn
```

### Testing account for UAT and Staging env

```sh
go run cmd/utils/main.go firebase createAccount --userID thu.vo+e2eschool@manabie.com --credentials GOOGLE_APPLICATION_CREDENTIALS_PATH

go run cmd/utils/main.go firebase createAccount --userID thu.vo+e2eadmin@manabie.com --credentials GOOGLE_APPLICATION_CREDENTIALS_PATH

```

### Remote debugging with delve

```sh
export DEBUG_ENABLED=true
./deployments/sk.bash

kubectl -n backend port-forward <<POD_NAME>> 40000

# Debug with delve client:
dlv connect 127.0.0.1:40000

# Debug with vscode:
Ctrl +P => Go:  Install/Update Tools => Update dlv-dap to latest version
Crtl + Shift + D => Connect to server
```
