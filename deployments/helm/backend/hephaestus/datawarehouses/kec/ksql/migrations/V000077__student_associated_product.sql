SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.student_associated_product', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENT_ASSOCIATED_PRODUCT_STREAM_FORMATTED_V1
    AS SELECT
        STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID + STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->ASSOCIATED_PRODUCT_ID as KEY,
        STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID AS STUDENT_PRODUCT_ID,
        STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->ASSOCIATED_PRODUCT_ID AS ASSOCIATED_PRODUCT_ID,
        STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1
    WHERE STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID + STUDENT_ASSOCIATED_PRODUCT_STREAM_ORIGIN_V1.AFTER->ASSOCIATED_PRODUCT_ID
    EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS STUDENT_ASSOCIATED_PRODUCT_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_ASSOCIATED_PRODUCT_STREAM_FORMATTED_V1',
      'fields.whitelist'='student_product_id,associated_product_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.student_associated_product',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_PRODUCT_ID:student_product_id,ASSOCIATED_PRODUCT_ID:associated_product_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='student_product_id,associated_product_id'
);
