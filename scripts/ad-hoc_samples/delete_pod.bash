#!/bin/bash
set -eu

kubectl delete pod -n $NAMESPACE $1