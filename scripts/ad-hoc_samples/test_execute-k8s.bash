#!/bin/bash
set -eu

kubectl exec -it -n $NAMESPACE $1 -- $2