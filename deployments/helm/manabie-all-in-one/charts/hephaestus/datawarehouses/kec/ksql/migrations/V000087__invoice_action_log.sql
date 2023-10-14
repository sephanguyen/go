SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS INVOICE_ACTION_LOG_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.invoice_action_log', value_format='AVRO');

CREATE STREAM IF NOT EXISTS INVOICE_ACTION_LOG_STREAM_FORMATTED_V1
    AS SELECT
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->INVOICE_ACTION_ID AS KEY,
        AS_VALUE(INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->INVOICE_ACTION_ID) AS INVOICE_ACTION_ID,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->PAYMENT_SEQUENCE_NUMBER AS PAYMENT_SEQUENCE_NUMBER,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->INVOICE_ID AS INVOICE_ID,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->USER_ID AS STAFF_ID,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ACTION AS ACTION,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ACTION_DETAIL AS ACTION_DETAIL,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ACTION_COMMENT AS ACTION_COMMENT,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM INVOICE_ACTION_LOG_STREAM_ORIGIN_V1
    WHERE INVOICE_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->INVOICE_ACTION_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS INVOICE_ACTION_LOG_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}INVOICE_ACTION_LOG_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS INVOICE_ACTION_LOG_PUBLIC_INFO_V1
AS SELECT
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.KEY AS INVOICE_ACTION_ID,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.PAYMENT_SEQUENCE_NUMBER AS PAYMENT_SEQUENCE_NUMBER,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.INVOICE_ID AS INVOICE_ID,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.STAFF_ID AS STAFF_ID,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.ACTION AS ACTION,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.ACTION_DETAIL AS ACTION_DETAIL,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.ACTION_COMMENT AS ACTION_COMMENT,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.CREATED_AT AS CREATED_AT,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.UPDATED_AT AS UPDATED_AT,
    INVOICE_ACTION_LOG_TABLE_FORMATTED_V1.DELETED_AT AS DELETED_AT
FROM INVOICE_ACTION_LOG_TABLE_FORMATTED_V1;


CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_ACTION_LOG_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_ACTION_LOG_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_action_id,payment_sequence_number,invoice_id,staff_id,action,action_detail,action_comment,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.invoice_action_log',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='INVOICE_ACTION_ID:invoice_action_id,PAYMENT_SEQUENCE_NUMBER:payment_sequence_number,INVOICE_ID:invoice_id,STAFF_ID:staff_id,ACTION:action,ACTION_DETAIL:action_detail,ACTION_COMMENT:action_comment,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='invoice_action_id'
);
