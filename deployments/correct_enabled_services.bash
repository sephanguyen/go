#!/bin/bash

set -euo pipefail

RULES="deployments/helm/manabie-all-in-one/rules.yaml"
VALUES="deployments/helm/manabie-all-in-one/values.yaml"
PORTS="deployments/helm/manabie-all-in-one/ports.yaml"

correct_file() {
    for key in $(yq 'keys' $RULES)
    do
        if [[ "$key" != "-" ]]; then
            if [[ "$(yq ". | has(\"$key\")" $1)" == "true" ]]; then
                yq -i ".${key}.waitForServices = []" $1
            fi
            index=0
            for name in $(yq -r ".${key}[]" $RULES)
            do
                yq -i ".global.${name}.enabled = true" $1
                if [[ "$(yq ". | has(\"$key\")" $1)" == "true" ]]; then
                    port=$(yq -r ".${name}" $PORTS)
                    yq -i ".${key}.waitForServices[${index}].name = \"$name\" |
                        .${key}.waitForServices[${index}].port = $port" $1
                fi
            done
        fi
    done
}
correct_file $VALUES


