SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS INVOICE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.invoice', value_format='AVRO');

CREATE STREAM IF NOT EXISTS INVOICE_STREAM_FORMATTED_V1 with (kafka_topic='{{ .Values.topicPrefix }}INVOICE_STREAM_FORMATTED_V1', value_format='AVRO')
    AS SELECT
        INVOICE_STREAM_ORIGIN_V1.AFTER->INVOICE_ID AS KEY,
        AS_VALUE(INVOICE_STREAM_ORIGIN_V1.AFTER->INVOICE_ID) AS INVOICE_ID,
        INVOICE_STREAM_ORIGIN_V1.AFTER->INVOICE_SEQUENCE_NUMBER AS INVOICE_SEQUENCE_NUMBER,
        INVOICE_STREAM_ORIGIN_V1.AFTER->TYPE AS INVOICE_TYPE,
        INVOICE_STREAM_ORIGIN_V1.AFTER->STATUS AS INVOICE_STATUS,
        INVOICE_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        INVOICE_STREAM_ORIGIN_V1.AFTER->SUB_TOTAL AS SUB_TOTAL,
        INVOICE_STREAM_ORIGIN_V1.AFTER->TOTAL AS TOTAL,
        INVOICE_STREAM_ORIGIN_V1.AFTER->OUTSTANDING_BALANCE AS OUTSTANDING_BALANCE,
        INVOICE_STREAM_ORIGIN_V1.AFTER->AMOUNT_PAID AS AMOUNT_PAID,
        INVOICE_STREAM_ORIGIN_V1.AFTER->AMOUNT_REFUNDED AS AMOUNT_REFUNDED,
        INVOICE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        INVOICE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        INVOICE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM INVOICE_STREAM_ORIGIN_V1
    WHERE INVOICE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY INVOICE_STREAM_ORIGIN_V1.AFTER->INVOICE_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS INVOICE_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}INVOICE_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS INVOICE_PUBLIC_INFO_V1
AS SELECT
    INVOICE_TABLE_FORMATTED_V1.KEY as ID,
    INVOICE_TABLE_FORMATTED_V1.INVOICE_ID as INVOICE_ID,
    INVOICE_TABLE_FORMATTED_V1.INVOICE_SEQUENCE_NUMBER AS INVOICE_SEQUENCE_NUMBER,
    INVOICE_TABLE_FORMATTED_V1.INVOICE_TYPE AS INVOICE_TYPE,
    INVOICE_TABLE_FORMATTED_V1.INVOICE_STATUS AS INVOICE_STATUS,
    INVOICE_TABLE_FORMATTED_V1.STUDENT_ID AS STUDENT_ID,
    INVOICE_TABLE_FORMATTED_V1.SUB_TOTAL AS SUB_TOTAL,
    INVOICE_TABLE_FORMATTED_V1.TOTAL AS TOTAL,
    INVOICE_TABLE_FORMATTED_V1.OUTSTANDING_BALANCE AS OUTSTANDING_BALANCE,
    INVOICE_TABLE_FORMATTED_V1.AMOUNT_PAID AS AMOUNT_PAID,
    INVOICE_TABLE_FORMATTED_V1.AMOUNT_REFUNDED AS AMOUNT_REFUNDED,
    INVOICE_TABLE_FORMATTED_V1.CREATED_AT AS CREATED_AT,
    INVOICE_TABLE_FORMATTED_V1.UPDATED_AT AS UPDATED_AT,
    INVOICE_TABLE_FORMATTED_V1.DELETED_AT AS DELETED_AT
FROM INVOICE_TABLE_FORMATTED_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_key,invoice_id,invoice_sequence_number,type,status,student_id,sub_total,total,outstanding_balance,amount_paid,amount_refunded,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='invoicemgmt.invoice_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:invoice_key,INVOICE_ID:invoice_id,INVOICE_SEQUENCE_NUMBER:invoice_sequence_number,INVOICE_TYPE:type,INVOICE_STATUS:status,STUDENT_ID:student_id,SUB_TOTAL:sub_total,TOTAL:total,OUTSTANDING_BALANCE:outstanding_balance,AMOUNT_PAID:amount_paid,AMOUNT_REFUNDED:amount_refunded,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='invoice_id'
);
