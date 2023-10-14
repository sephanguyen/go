#!/bin/bash
# required: kcctl, curl, jq, awk
# required: connected to kafka connect in server
set -e
echo "Checking for missing topics in KSQL"
source_connectors=($(kcctl get connectors | grep source | awk '{print $1}'))
list_topics=()
for source_connector in "${source_connectors[@]}"
do
  list_topics+=($(curl -X GET http://localhost:8083/connectors/$source_connector/topics | jq ".$source_connector.topics" | jq -r '.[]' | awk -F'.' '{split($0,arr,"."); print arr[4]"."arr[5]}' | sed 's/\"//g'))
done

migration_dir=../../backend/hephaestus/datalake/migrations
table_names=()
for sql_file in "$migration_dir"/*.sql; do
    table_names+=($(grep -i "CREATE TABLE" "$sql_file" | awk '{print $6}' | sed 's/\"//g'))
done

echo "-------------------------------------"
echo "missing topics"
missing_tables_data=()
for table_name in "${table_names[@]}"
do
    found=0
    for topic in "${list_topics[@]}"
    do
        if [[ $topic == "$table_name" ]]; then
            found=1
            break
        fi
    done

    if [[ $found -eq 0 ]]; then
        missing_tables_data+=($table_name)
    fi
done


function build_snapshot_command() {
    snap_table_name=$2
    delimiter=","
    snp_query=$(printf "${delimiter}%s" "${snap_table_name[@]}")
    snp_query=${snp_query:${#delimiter}}  # remove leading delimiterecho "$joined"
    echo $snp_query
    date_now=$(date +"%Y-%m-%d-%H-%M-%S-%N")
    echo "-------------------------------------"
    echo "SNAPSHOT COMMAND" 
    echo "INSERT INTO public.alloydb_dbz_signal (id,"type","data") VALUES
    ('id-$date_now-multi','execute-snapshot'
    ,'{"data-collections": [$snp_query]}');"
    echo "-------------------------------------"
}

sorted_list=($(printf "%s\n" "${missing_tables_data[@]}" | sort))
snapshot_query=()
prv_schema=""
for table_name in "${sorted_list[@]}"
do
    if [[ $table_name == "public.snapshot_datawarehouse_signal" ]]; then
        continue
    fi
    schema=$(echo $table_name | awk -F'.' '{split($0,arr,"."); print arr[1]}')
    table=$(echo $table_name | awk -F'.' '{split($0,arr,"."); print arr[2]}')
    echo $schema
    echo $table
    if [[ $schema != $prv_schema ]]; then
        if [[ $prv_schema != "" ]]; then
            echo "-------------------------------------"
            build_snapshot_command $prv_schema $snapshot_query
            snapshot_query=()
        fi
        prv_schema=$schema
    fi
    snapshot_query+=("\"public.$table\"") 
done
build_snapshot_command $prv_schema $snapshot_query

query=()
for table_name in "${sorted_list[@]}"
do
    if [[ $table_name == "public.snapshot_datawarehouse_signal" ]]; then
        continue
    fi
    echo $table_name
    query+=("select '$table_name' as tablename, count(*) as record_count from $table_name \n") 
done
echo "-------------------------------------"
delimiter=" UNION ALL "
joined=$(printf "${delimiter}%s" "${query[@]}")
joined=${joined:${#delimiter}}  # remove leading delimiterecho "$joined"

echo -e $joined
