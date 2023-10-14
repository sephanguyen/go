#!/bin/bash

if [[ -n "$1" ]]; then
    ENV=$1
else
    ENV="local"
fi
if [[ -n "$2" ]]; then
    ORG=$2
else
    ORG="*"
fi
base_text=".secrets.encrypted.yaml"

for encryptedPath in ./deployments/helm/manabie-all-in-one/charts/*/secrets/$ORG/$ENV/*$base_text;
do
    if [[ "$(cat $encryptedPath | grep storage)" != "" && "$(echo $encryptedPath | grep yasuo)" == "" ]];
    then
        echo $encryptedPath
        decryptedPath="${encryptedPath//encrypted/decrypted}"
        sops -d $encryptedPath > $decryptedPath
        yq -i 'del(.storage)' $decryptedPath
        sops -e $decryptedPath > $encryptedPath
        rm $decryptedPath
    fi
done
