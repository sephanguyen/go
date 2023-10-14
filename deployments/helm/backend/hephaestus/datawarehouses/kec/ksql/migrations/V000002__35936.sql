set 'auto.offset.reset' = 'earliest';
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
      'table.name.format'='bob.scheduler_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SCHEDULER_ID:scheduler_id,START_DATE:start_date,END_DATE:end_date,FREQ:freq,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='scheduler_id'
);
