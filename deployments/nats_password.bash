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
base_text="nats.secrets.conf.encrypted.yaml"
plain_text="nats.secrets.conf.yaml"


IFS="
"
for path in ./deployments/helm/platforms/nats-jetstream/secrets/$ORG/$ENV;
do
    ENCRYPTED_FILE="$path/$base_text"
    decrypted_text="$(sops -d $ENCRYPTED_FILE)"
    SAVE_FOLDER="$path/"
    for val in $decrypted_text;
    do
        service=$(echo $val | grep -Po "[a-z]+_password")
        if [ "$service" != "" ];
        then
            s=$(echo $service | grep -Po "[a-z]+_")
            filepath="$SAVE_FOLDER$s$base_text"
            plainpath="$SAVE_FOLDER$s$plain_text"
            echo "data: |" > "$plainpath"
            echo "$val" >> "$plainpath"
            echo "$s$base_text"
            sops -e --output "$filepath" "$plainpath"
            rm "$plainpath"
        else
            echo "$val" >> "$SAVE_FOLDER$plain_text"
        fi
    done
    sops -e --output "$SAVE_FOLDER$base_text" "$SAVE_FOLDER$plain_text"
    rm "$SAVE_FOLDER$plain_text"
done