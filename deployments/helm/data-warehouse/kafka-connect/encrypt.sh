#!/bin/bash
set -eo pipefail
# used to encrypt secrets and connectors config

org=$1
env=$2
project=""
location="global"
key=""

prepare_kms_values() {
    case "$org" in
        manabie)
            key=github-actions
            case "$env" in
                local | dev)
                    project="dev-manabie-online"
                    ;;
                stag | uat)
                    project="staging-manabie-online"
                    ;;
                *)
                    echo "Invalid <env> value \"$env\" for manabie, must be one of: local dev stag uat"
                    exit 1
                    ;;
            esac
            ;;
        *)
            echo "Invalid <org> value \"$org\", must be one of: manabie"
            exit 1
            ;;
    esac
}

prepare_kms_values

encrypt_secret() {
    local destination=./secrets/$org/$env
    local file_name=$1
    local secret_file_in=${destination}/$file_name
    local secret_file_out=${destination}/${file_name%.*}.encrypted.${file_name##*.}

    echo "Encrypting $secret_file_in to:"
    echo -e "\t$secret_file_out"

    sops --encrypt --gcp-kms projects/$project/locations/$location/keyRings/deployments/cryptoKeys/$key $secret_file_in > $secret_file_out
}

echo ""
echo "'''''''''''''''''''''''''''''''''''''''"
echo "'                                     '"
echo "' Encrypting secret for kafka connect '"
echo "'                                     '"
echo "'''''''''''''''''''''''''''''''''''''''"
echo ""
encrypt_secret kafka-connect.secrets.properties
encrypt_secret kafka-connect.secrets.env.yaml