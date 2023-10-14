#!/bin/bash

# check config of kafka connect for every env
expect_config_file="./expect_config.json"

check_config_values() {
    local org=$1
    local env=$2

    local expect_config=$(cat $expect_config_file | jq  --arg org "$org" --arg env "$env" '.[$org] | .[$env]' )

    echo "Checking config file kafka-connect.env of ORG: $org ENV: $env"
    local kafka_connect_env_file="./configs/$org/$env/kafka-connect.env"
    cat $kafka_connect_env_file | grep -v '^#' | while read -r line; do
        p='^(\w+)_SOURCE_DATABASE_URL:\spostgres:\/\/((\w+-?\.?)+):5432\/(\w+)\?user=((\w+-?)+@?(\w+-?\.?)+)(&password=.*)?&sslmode=disable$'
        if [[ $line =~ $p ]]; then
            # lowercase the variable name
            local SERVICE=${BASH_REMATCH[1],,}
            local DBHOST=${BASH_REMATCH[2],,}
            local DBNAME=${BASH_REMATCH[4],,}
            local DBUSER=${BASH_REMATCH[5],,}

            local EXPECTED_DBHOST=$(echo $expect_config | jq '.dbhost' | tr -d '"')
            local EXPECTED_DBUSER=$(echo $expect_config | jq '.dbuser' | tr -d '"')
            local DBPREFIX=$(echo $expect_config | jq '.dbprefix' | tr -d '"')
            local EXPECTED_DBNAME="${DBPREFIX}${SERVICE}"

            echo -e -n "\t SERVICE: $SERVICE"

            if [ $DBHOST != $EXPECTED_DBHOST ]; then 
                echo " - Error: unexpected database host, expect $EXPECTED_DBHOST but got $DBHOST"
                exit 1
            fi

            if [ $DBUSER != $EXPECTED_DBUSER ]; then 
                echo " - Error: unexpected database user, expect $EXPECTED_DBUSER but got $DBUSER"
                exit 1
            fi

            if [ $DBNAME != $EXPECTED_DBNAME ]; then 
                echo "unexpected database user, expect $EXPECTED_DBNAME but got $DBNAME"
                exit 1
            fi

            echo -e " - OK"
        fi
    done
}

check_config_values manabie local
check_config_values manabie stag
check_config_values manabie uat
check_config_values manabie prod
check_config_values jprep stag
check_config_values jprep uat
check_config_values jprep prod
check_config_values aic prod
check_config_values ga prod
check_config_values renseikai prod
check_config_values synersia prod
check_config_values tokyo prod
