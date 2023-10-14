#!/bin/bash

# This script is to quick exec into a k8s runner that is running CI.

set -eu

runner_pod_name=$1
echo "Exec-ing into runner ${runner_pod_name}"
kubectl -n actions-runner-system exec -it "${runner_pod_name}" -- bash
