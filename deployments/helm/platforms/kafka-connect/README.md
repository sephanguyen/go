

### Configuration folder

```
connectors
└── manabie
    └── local
        ├── sink
        │   ├── _bob_to_eureka_learning_objectives.json
        │   └── _bob_to_eureka_users.json
        └── source
            └── _bob.json
```


### Source connector config

we only have one config for a source database
example: bob.json, eureka.json, tom.json

#### Config properties

```
{
  "name": "local_manabie_bob_source_connector",
  "config": {
    "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
    "database.dbname": "bob",
    "database.hostname": "${file:/decrypted/kafka-connect.secrets.properties:hostname}",
    "database.user": "${file:/decrypted/kafka-connect.secrets.properties:user}",
    "database.password": "${file:/decrypted/kafka-connect.secrets.properties:password}",
    "database.port": "5432",
    "database.server.name": "bob",
    "database.sslmode": "disable",
    "plugin.name": "pgoutput",
    "tasks.max": "1",
    "key.converter":"io.confluent.connect.avro.AvroConverter",
    "key.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "key.converter.schemas.enable": "false",
    "value.converter":"io.confluent.connect.avro.AvroConverter",
    "value.converter.schema.registry.url":"http://cp-schema-registry:8081",
    "value.converter.schemas.enable": "false",
    "slot.name": "bob",
    "slot.drop.on.stop": "false",
    "publication.autocreate.mode": "disabled",
    "publication.name": "debezium_publication",
    "snapshot.mode": "initial",
    "tombstones.on.delete": "true",
    "heartbeat.interval.ms": "20000",
    "schema.include.list": "public",
    "table.include.list": "public.dbz_signals,public.users,public.learning_objectives",
    "signal.data.collection": "public.dbz_signals",
    "time.precision.mode": "connect",
    "topic.creation.default.replication.factor": 3,  
    "topic.creation.default.partitions": 10,  
    "topic.creation.default.cleanup.policy": "compact",  
    "topic.creation.default.compression.type": "lz4",
    "transforms": "route",
    "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
    "transforms.route.regex": "([^.]+).([^.]+).([^.]+)",
    "transforms.route.replacement": "local.manabie.$1.$2.$3"
  }
}
```

### Add captured table

We use incremental snap shot:
- add table name to field `table.include.list`
- trigger snapshot for this table by inserting to table public.dbz_signal. please put this query in the job file.
```
INSERT INTO myschema.dbz_signals VALUES('ad-hoc-1', 'execute-snapshot', '{"data-collections": ["public.users", "public.learning_objectives"]}') ON CONFLICT DO NOTHING
```


### Sink connector config

#### Config properties
```
{
    "name": "local_manabie_bob_to_eureka_users_sink_connector",
    "config": {
        "connector.class": "io.confluent.connect.jdbc.JdbcSinkConnector",
        "tasks.max": "1",
        "topics": "local.manabie.bob.public.users",
        "connection.url": "${file:/decrypted/kafka-connect.secrets.properties:url}",
        "transforms": "unwrap,route",
        "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
        "transforms.unwrap.drop.tombstones": "true",
        "transforms.unwrap.delete.handling​.mode": "drop",
        "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
        "transforms.route.regex": "([^.]+).([^.]+).([^.]+).([^.]+).([^.]+)",
        "transforms.route.replacement": "$5",
        "auto.create": "true",
        "insert.mode": "upsert",
        "delete.enabled": "true",
        "pk.mode": "record_key"
    }
}
```

the current topic of user has format `local.manabie.bob.public.users`
we use transform to extract last field which is table name `users`
because default behavior of sink connector is using topic name as the table name


### Encrypt secret and connector config

we now only encrypt the secret and leave raw connector config

```
cd ./deployments/helm/platforms/kafka-connect
./encrypt.sh [org] [env]
```

example 
```
./encrypt.sh manabie local 
```

### Data piline definition

Data pipeline sync table from postgresql to postgresql

Example:
```
datapipelines:
- name: bob_locations
  source:
    database: bob
    schema: public
    table: locations
  sinks:
  - database: entryexitmgmt
    schema: public
    table: locations
    deployOrg: [e2e, manabie, jprep, aic, ga, renseikai, synersia, tokyo]
    captureDeleteAll: false
```

| field                                          | description                                                                                                                                                                    |
|------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| datapipelines                                  | list of data pipeline which sync table from postgresql to postgresql                                                                                                           |
| datapipelines[i].name                          | Name contain the source database and table                                                                                                                                     |
| datapipelines[i].source                        | Info of source table                                                                                                                                                           |
| datapipelines[i].source.database               | Name of source database                                                                                                                                                        |
| datepipelines[i].source.schema                 | Name of source database schema                                                                                                                                                 |
| datepipelines[i].source.table                  | Name of source table will be synced                                                                                                                                            |
| datepipelines[i].sinks                         | List of sink table, one source table can be synced to multiple place                                                                                                           |
| datepipelines[i].sinks[i].database             | Name of sink database                                                                                                                                                          |
| datepipelines[i].sinks[i].schema               | Name of sink database schema                                                                                                                                                   |
| datepipelines[i].sinks[i].table                | Name of sink database table (normally it will be the same as source table name)                                                                                                |
| datepipelines[i].sinks[i].deployOrg            | Specify which organization will be deployed. Org: [manabie, jprep, aic, ga, renseikai, synersia]. Leaving empty means all organizations                                        |
| datepipelines[i].sinks[i].deployEnv            | Specify which environment will be deployed. Env: [stag, uat, prod]. Leaving empty means all environments                                                                       |
| datepipelines[i].sinks[i].captureDeleteAll | Enable capture hard delete event for all envs. Default is false. Usually, our features're implementing soft delete rather than hard delete. So in most case, you will only set it to false. |
| datepipelines[i].sinks[i].excludeColumns       | exclude sensitive columns, if you don't want these column to be write to sink table.                                                                                           |


Run command:
```
make gen-data-pipeline
```

this tool read input from file `deployments/helm/platforms/kafka-connect/datapipeline_postgresql2postgresql.yaml` to generate kafka connect - sink connector config to folder `deployments/helm/platforms/kafka-connect/generated_connectors/sink`

Improve: 
- will test and refactor for our deployment to use the these generate sink connector config
- Update tool be able to delete connector config which not defined

