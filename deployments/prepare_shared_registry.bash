#!/bin/bash

find ./skaffold.* -type f
find ./skaffold.* -type f -exec sed -i 's/localhost:5001/kind-reg.actions-runner-system.svc/g' {} \;
find ./deployments/setup_istio.bash -type f
find ./deployments/setup_istio.bash -type f -exec sed -i 's/localhost:5001/kind-reg.actions-runner-system.svc/g' {} \;
find ./scripts/clean_docker_daemon.bash -type f 
find ./scripts/clean_docker_daemon.bash -type f -exec sed -i 's/localhost:5001/kind-reg.actions-runner-system.svc/g' {} \;
find ./deployments/helm/backend/bob/local-manabie-values.yaml -type f -exec sed -i 's/localhost:5001/kind-reg.actions-runner-system.svc/g' {} \;