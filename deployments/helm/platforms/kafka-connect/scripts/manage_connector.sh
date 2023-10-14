
#!/bin/bash

wait_for() {
    local retry=0
    local max_retry=100
    until $(curl -s -o /dev/null -f $1); do
        echo "waiting for $1 goes live ..."
        if [ $retry -gt $max_retry ]; then
            echo "ERROR: max retry to wait for $1 goes live"
            exit 1
        fi
        let "retry += 1"
        sleep 2
    done
}

create_connector() {
    local connector_name=$1
    echo "Not found $connector_name. Creating $connector_name"
    local http_code=$(curl -s -w '%{http_code}' -o /dev/null \
    -X POST \
    -H "Accept:application/json" \
    -H "Content-Type:application/json" \
    $KAFKA_CONNECT/connectors/ \
    -d @$file)
    if [ $http_code -ne 201 ]; then
        echo "ERROR: Cannot create $connector_name"
        exit 1
    fi
    echo -e "\tCreate $connector_name successfully\n"
}

update_connector() {
    echo "Found $connector_name. Updating $connector_name "
    local connector_name=$1
    local http_code=$(curl -s -w '%{http_code}' -o /dev/null \
    -X PUT \
    -H "Accept:application/json" \
    -H "Content-Type:application/json" \
    $KAFKA_CONNECT/connectors/$connector_name/config \
    -d "$connector_config")

    if [ $http_code -ne 200 ]; then
        echo "ERROR: Cannot update $connector_name"
        exit 1
    fi
    echo -e "\tUpdate $connector_name successfully\n"
}


upsert_source_connector() {
    local dir=$1
    echo $dir
    for file in $dir/*
    do
        if [ -s $file ]; then
            connector_name=$(cat $file | jq -r ".name")
            connector_config=$(cat $file | jq -r ".config")
            publication_name=$(cat $file | jq -r ".config.\"publication.name\"")

            # prepare publication for tables
            dbname=$(cat $file | jq -r ".config.\"database.server.name\"")
            db_url=""

            kafka_connect_env_file="/config/kafka-connect.env"
            case "$dbname" in

            *bob)
                db_url=$(grep BOB_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
            ;;

            *calendar)
                db_url=$(grep CALENDAR_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *draft)
                db_url=$(grep DRAFT_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *eureka)
                db_url=$(grep EUREKA_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
            ;;

            *entryexitmgmt)
                db_url=$(grep ENTRYEXITMGMT_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
            ;;

            *fatima)
                db_url=$(grep FATIMA_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *invoicemgmt)
                db_url=$(grep INVOICEMGMT_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;
            
            *lessonmgmt)
                db_url=$(grep LESSONMGMT_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *timesheet)
                db_url=$(grep TIMESHEET_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *tom)
                db_url=$(grep TOM_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *zeus)
                db_url=$(grep ZEUS_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;

            *mastermgmt)
                db_url=$(grep MASTERMGMT_SOURCE_DATABASE_URL $kafka_connect_env_file | awk '{print$2}')
                ;;
    
            *)
                echo -n "unknown"
                exit 1
                ;;
            esac

            if [ "$ENV" == "local" ]; then
                # We need to add table to publication upfront
                # in local we can insert it here with super user privilege
                # but in other evironment, we need to manually create it before the deployment
                is_publication_existed=$(psql $db_url -XAt -c "SELECT count(*) FROM pg_catalog.pg_publication WHERE pubname='$publication_name';")
                if [ $is_publication_existed -eq 0 ]; then
                    echo "Create publication $publication_name"
                    psql $db_url -c "CREATE PUBLICATION $publication_name;"
                fi

                captured_tables=$(cat $file | jq -r ".config.\"table.include.list\"")
                captured_tables_arr=$(echo $captured_tables | sed -E 's/,/ /g')
                for table in ${captured_tables_arr[@]}
                do
                is_table_in_publication=$(psql $db_url -XAt -c "SELECT count(*) FROM pg_catalog.pg_publication_tables WHERE pubname='$publication_name' AND schemaname || '.' || tablename='$table';")
                if [ $is_table_in_publication -eq 0 ]; then
                    echo "Add table $table to publication $publication_name"
                    until psql $db_url -c "ALTER PUBLICATION $publication_name ADD TABLE $table"
                    do
                        # this behavior only in the local, need to wait for migration finished
                        # so keep waiting until it's done
                        echo "wait for table $table to be available"
                        sleep 2
                    done
                fi
                done
            fi

            http_code=$(curl -s -w '%{http_code}' -o /dev/null -X GET $KAFKA_CONNECT/connectors/$connector_name/status)

            if [ $http_code -eq 404 ]; then
                create_connector $connector_name
            else
                prev_tables=$(curl -s $KAFKA_CONNECT/connectors/$connector_name/config | jq -r ".\"table.include.list\"")
                prev_tables_arr=$(echo $prev_tables | sed -E 's/,/ /g')
                cur_tables=$(cat $file | jq -r ".config.\"table.include.list\"")
                cur_tables_arr=$(echo $cur_tables | sed -E 's/,/ /g')
                received_signal_table=$(cat $file | jq -r ".config.\"signal.data.collection\"")
                added_tables=()
                for table in ${cur_tables_arr[@]}
                do
                    if [[ ! "${prev_tables_arr[*]}" =~ "${table}" ]]; then
                        added_tables+=($table)
                    fi
                done

                update_connector $connector_name

                if [ ${#added_tables[@]} -ne 0 ]; then
                    echo "Incremental snapshot tables: $added_tables"
                    id=$(uuidgen)
                    dataCollections=""
                    delim=""
                    for table in ${added_tables[@]}
                    do
                        dataCollections+="$delim\"$table\""
                        delim=","
                    done
                    stmt="INSERT INTO $received_signal_table VALUES ('$id', 'execute-snapshot', '{\"data-collections\": [$dataCollections]}')"
                    psql $db_url -c "$stmt"
                fi
            fi
        fi
    done
}

upsert_sink_connector() {
    local dir=$1
    for file in $dir/*
    do
        if [ -s $file ]; then
            connector_name=$(cat $file | jq -r ".name")
            connector_config=$(cat $file | jq -r ".config")

            http_code=$(curl -s -w '%{http_code}' -o /dev/null -X GET $KAFKA_CONNECT/connectors/$connector_name/status)

            if [ $http_code -eq 404 ]; then
                create_connector $connector_name
            else
                update_connector $connector_name
            fi
        fi
    done
}


wait_for "$KAFKA_CONNECT"
wait_for "$SCHEMA_REGISTRY"

upsert_source_connector /etc/debezium/source_connectors
upsert_sink_connector /etc/debezium/sink_connectors

# Create connector sync to alloydb
if [ "$SYNC_ALLOYDB_ENABLED" == "true" ]; then 
    echo "sync alloydb enabled"
    upsert_source_connector /etc/debezium/alloydb_source_connectors
    upsert_sink_connector /etc/debezium/alloydb_sink_connectors
fi