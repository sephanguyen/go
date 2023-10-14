#!/bin/bash
read -p 'Your new service name: ' namevar
read -p 'Your new service port: ' portvar
sed -e "s/yournewservicename/${namevar}/g" \
    -e "s/696969/${portvar}/g" \
    ./scripts/template/add_new_helm_chart.patch | git apply --binary --ignore-whitespace
yq -i ".global.${namevar}.enabled = false" deployments/helm/manabie-all-in-one/disable-all.yaml
yq -i ".global.${namevar}.enabled = true" deployments/helm/manabie-all-in-one/values.yaml
yq -i ".${namevar}.migrationEnabled = false" deployments/helm/manabie-all-in-one/values.yaml
yq -i ".${namevar}.hasuraEnabled = false" deployments/helm/manabie-all-in-one/values.yaml
yq -i ".${namevar}.readinessProbe.enabled = true" deployments/helm/manabie-all-in-one/values.yaml
