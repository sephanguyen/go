#!/bin/bash

# This script get the kubecontext specified by PROJECT, REGION,
# and CLUSTER environment variable.

set -eu

gcloud container clusters get-credentials --project="$PROJECT" --region="$REGION" "$CLUSTER"

# Sanity check
# If we want to checking cluster status, we need grant `container.namespaces.get`
# We don't need checking namespaces.
# ./.github/tests/k8s_cluster.test.bash