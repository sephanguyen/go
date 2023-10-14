set 'auto.offset.reset' = 'earliest';

-- ORIGINAL

CREATE STREAM IF NOT EXISTS PREFECTURE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.prefecture', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PREFECTURE_STREAM_FORMATED_V1
AS SELECT
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->PREFECTURE_ID AS KEY,
    AS_VALUE(PREFECTURE_STREAM_ORIGIN_V1.AFTER->PREFECTURE_ID) AS PREFECTURE_ID,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->PREFECTURE_CODE AS PREFECTURE_CODE,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->COUNTRY AS COUNTRY,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
    PREFECTURE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
FROM PREFECTURE_STREAM_ORIGIN_V1
PARTITION BY AFTER->PREFECTURE_ID
EMIT CHANGES;

CREATE STREAM IF NOT EXISTS PREFECTURE_PUBLIC_INFO_V1
AS SELECT
    PREFECTURE_STREAM_FORMATED_V1.KEY AS PREFECTURE_ID,
    PREFECTURE_STREAM_FORMATED_V1.NAME AS NAME,
    PREFECTURE_STREAM_FORMATED_V1.PREFECTURE_CODE AS PREFECTURE_CODE,
    PREFECTURE_STREAM_FORMATED_V1.COUNTRY AS COUNTRY,
    PREFECTURE_STREAM_FORMATED_V1.CREATED_AT AS CREATED_AT,
    PREFECTURE_STREAM_FORMATED_V1.UPDATED_AT AS UPDATED_AT,
    PREFECTURE_STREAM_FORMATED_V1.DELETED_AT AS DELETED_AT
FROM PREFECTURE_STREAM_FORMATED_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_PREFECTURE_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PREFECTURE_PUBLIC_INFO_V1',
      'fields.whitelist'='prefecture_code,country,name,created_at,updated_at,deleted_at,prefecture_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.prefecture',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PREFECTURE_CODE:prefecture_code,COUNTRY:country,NAME:name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,PREFECTURE_ID:prefecture_id',
      'pk.fields'='prefecture_id'
);
