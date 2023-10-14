#!/bin/bash

set -eux

URL="https://admin.$ENV.manabie.io/imd"

if [[ "$ENV" == "production" ]]; then
   URL="https://admin.prod.tokyo.manabie.io/imd"
fi
if [[ "$ENV" == "preproduction" ]]; then
   URL="https://admin.prep.tokyo.manabie.io/imd"
fi


IMPORTMAP_PATH="import-map.json"

if [ -e "$IMPORTMAP_PATH" ]; then
    echo "File $IMPORTMAP_PATH exists."
    cat $IMPORTMAP_PATH

    curl -u admin:$IMD_PASSWORD -X PATCH "$URL/import-map.json?env=$ORG" \
     -H "Accept: application/json" -H "Content-Type: application/json" \
        --data "@$IMPORTMAP_PATH" 
       
fi

if [[ "$DELETE_SERVICE" != "" ]]; then

    # Convert the string to an array
    IFS=", " read -r -a array <<< "$DELETE_SERVICE"

    # Loop through the array
    for value in "${array[@]}"
    do
        echo "Deleting svc: $value"
        curl -u admin:$IMD_PASSWORD -X DELETE "$URL/services/$value?env=$ORG"
    done
fi
