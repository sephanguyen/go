set 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS SINK_SCHEDULER_PUBLIC_INFO_V1;
DROP CONNECTOR IF EXISTS SINK_DAY_INFO_PUBLIC_INFO_V1;
DROP CONNECTOR IF EXISTS SINK_DAY_TYPE_PUBLIC_INFO_V1;

DROP STREAM IF EXISTS SCHEDULER_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS SCHEDULER_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS DAY_INFO_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS DAY_INFO_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS DAY_TYPE_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS DAY_TYPE_STREAM_ORIGIN_V1;

CREATE STREAM IF NOT EXISTS SCHEDULER_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.scheduler', value_format='AVRO');
CREATE STREAM IF NOT EXISTS SCHEDULER_STREAM_FORMATED_V1 AS 
    SELECT
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->SCHEDULER_ID AS rowkey,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->FREQ AS FREQ,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        SCHEDULER_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        AS_VALUE(SCHEDULER_STREAM_ORIGIN_V1.AFTER->SCHEDULER_ID) AS SCHEDULER_ID
    FROM SCHEDULER_STREAM_ORIGIN_V1
    WHERE SCHEDULER_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';


CREATE SINK CONNECTOR IF NOT EXISTS SINK_SCHEDULER_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHEDULER_STREAM_FORMATED_V1',
      'fields.whitelist'='scheduler_id,start_date,end_date,freq,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.scheduler',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SCHEDULER_ID:scheduler_id,START_DATE:start_date,END_DATE:end_date,FREQ:freq,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='scheduler_id'
);

CREATE STREAM IF NOT EXISTS DAY_INFO_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.day_info', value_format='AVRO');
CREATE STREAM IF NOT EXISTS DAY_INFO_STREAM_FORMATED_V1 AS 
    SELECT
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DATE AS DATE,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID AS DAY_TYPE_ID,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->OPENING_TIME AS OPENING_TIME,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->STATUS AS STATUS,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->TIME_ZONE AS TIME_ZONE,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM DAY_INFO_STREAM_ORIGIN_V1
    WHERE DAY_INFO_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_DAY_INFO_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}DAY_INFO_STREAM_FORMATED_V1',
      'fields.whitelist'='location_id,date,day_type_id,opening_time,status,time_zone,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.day_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LOCATION_ID:location_id,DATE:date,DAY_TYPE_ID:day_type_id,OPENING_TIME:opening_time,STATUS:status,TIME_ZONE:time_zone,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='location_id,date'
);

CREATE STREAM IF NOT EXISTS DAY_TYPE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.day_type', value_format='AVRO');
CREATE STREAM IF NOT EXISTS DAY_TYPE_STREAM_FORMATED_V1 AS 
    SELECT
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID AS rowkey,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DISPLAY_NAME AS DISPLAY_NAME,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        AS_VALUE(DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID) AS DAY_TYPE_ID
    FROM DAY_TYPE_STREAM_ORIGIN_V1
    WHERE DAY_TYPE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_DAY_TYPE_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}DAY_TYPE_STREAM_FORMATED_V1',
      'fields.whitelist'='day_type_id,display_name,is_archived,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.day_type',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='DAY_TYPE_ID:day_type_id,DISPLAY_NAME:display_name,IS_ARCHIVED:is_archived,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='day_type_id'
);
