SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS PRODUCT_DISCOUNT_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.product_discount', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PRODUCT_DISCOUNT_STREAM_FORMATTED_V1
    AS SELECT
        PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->DISCOUNT_ID + PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID as KEY,
        PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->DISCOUNT_ID AS DISCOUNT_ID,
        PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
        PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CAST(NULL AS VARCHAR) AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM PRODUCT_DISCOUNT_STREAM_ORIGIN_V1
    WHERE PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->DISCOUNT_ID + PRODUCT_DISCOUNT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID
    EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS PRODUCT_DISCOUNT_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_DISCOUNT_STREAM_FORMATTED_V1',
      'fields.whitelist'='discount_id,product_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_discount',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='DISCOUNT_ID:discount_id,PRODUCT_ID:product_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='discount_id,product_id'
);
