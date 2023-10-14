#!/bin/bash

set -euo pipefail

mode="checkout" # use `git checkout` when rendering
if [[ "$(whoami)" == "anhpngt" ]]; then
  mode="multirepo" # use a different repo for rendering
fi

./deployments/sk.bash -- render --digest-source=none "$@" >new.yaml

if [[ "${mode}" == "checkout" ]]; then
  cur_branch=$(git rev-parse --abbrev-ref HEAD)
  git checkout develop
  ./deployments/sk.bash -- render --digest-source=none "$@" >old.yaml

  git checkout "${cur_branch}"
else
  # this assumes that you are in "backend" repo, while having "backend-ro" repo
  # with the same parent directory
  currentrepo=$(basename "${PWD}")
  developrepo="../backend-ro/"
  (cd "${developrepo}" && ./deployments/sk.bash -- render --digest-source=none "$@" >"../${currentrepo}/old.yaml")
fi

code --diff old.yaml new.yaml
