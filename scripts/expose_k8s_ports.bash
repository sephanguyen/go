#!/bin/bash

valid_namespaces=("emulator" "backend", "local-manabie-appsmith")
if [[ -z "$1" ]]; then
  NAMESPACES=(${valid_namespaces[@]}) 
else
  NAMESPACES="$1"
  if [[ ! " ${valid_namespaces[@]} " =~ " ${NAMESPACES} " ]]; then
      echo "Invalid value \"$NAMESPACES\" for namespace, must be a subset of: ${NAMESPACES[@]}"
      exit 1
  fi
fi

echo "Exposing port for: ${NAMESPACES[@]}"
for ns in "${NAMESPACES[@]}"; do
  # This script searchs for all pods in 'ns' namespace and forward their ports to localhost
  cmds=$(kubectl -n $ns get svc -o \
    go-template="{{range \$item := .items}}{{range.spec.ports}}kubectl -n $ns port-forward service/{{ \$item.metadata.name }} {{.port}}:{{.port}}{{\"\\n\"}}{{end}}{{end}}")
  while IFS= read -r line; do
    if [[ $line == *8080:8080 ]]; then
      echo "$line (skipped)"  # hasura always use port 8080, causing conflict
    else
      echo $line
      $line &
    fi
  done <<< "$cmds"
done

wait