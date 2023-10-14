SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS INVOICE_BILL_ITEM_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.invoice_bill_item', value_format='AVRO');

CREATE STREAM IF NOT EXISTS INVOICE_BILL_ITEM_STREAM_FORMATTED_V1
    AS SELECT
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->INVOICE_BILL_ITEM_ID AS KEY,
        AS_VALUE(INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->INVOICE_BILL_ITEM_ID) AS INVOICE_BILL_ITEM_ID,
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->INVOICE_ID AS INVOICE_ID,
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->BILL_ITEM_SEQUENCE_NUMBER AS BILL_ITEM_SEQUENCE_NUMBER,
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS INVOICE_BILL_ITEM_CREATED_AT,
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS INVOICE_BILL_ITEM_UPDATED_AT,
        INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS INVOICE_BILL_ITEM_DELETED_AT
    FROM INVOICE_BILL_ITEM_STREAM_ORIGIN_V1
    WHERE INVOICE_BILL_ITEM_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->INVOICE_BILL_ITEM_ID
    EMIT CHANGES;


CREATE TABLE IF NOT EXISTS INVOICE_BILL_ITEM_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}INVOICE_BILL_ITEM_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1
AS SELECT
    INVOICE_BILL_ITEM_TABLE_FORMATTED_V1.KEY AS INVOICE_BILL_ITEM_ID,
    AS_VALUE(INVOICE_TABLE_FORMATTED_V1.KEY) AS INVOICE_ID,
    INVOICE_TABLE_FORMATTED_V1.INVOICE_SEQUENCE_NUMBER AS INVOICE_SEQUENCE_NUMBER,
    INVOICE_TABLE_FORMATTED_V1.STUDENT_ID AS STUDENT_ID,
    INVOICE_BILL_ITEM_TABLE_FORMATTED_V1.BILL_ITEM_SEQUENCE_NUMBER AS BILL_ITEM_SEQUENCE_NUMBER,
    INVOICE_BILL_ITEM_TABLE_FORMATTED_V1.INVOICE_BILL_ITEM_CREATED_AT AS INVOICE_BILL_ITEM_CREATED_AT,
    INVOICE_BILL_ITEM_TABLE_FORMATTED_V1.INVOICE_BILL_ITEM_UPDATED_AT AS INVOICE_BILL_ITEM_UPDATED_AT,
    INVOICE_TABLE_FORMATTED_V1.CREATED_AT AS INVOICE_CREATED_AT,
    INVOICE_TABLE_FORMATTED_V1.UPDATED_AT AS INVOICE_UPDATED_AT
FROM INVOICE_BILL_ITEM_TABLE_FORMATTED_V1
JOIN INVOICE_TABLE_FORMATTED_V1
ON INVOICE_BILL_ITEM_TABLE_FORMATTED_V1.INVOICE_ID = INVOICE_TABLE_FORMATTED_V1.KEY;


CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_bill_item_id,invoice_id,invoice_sequence_number,student_id,bill_item_sequence_number,invoice_bill_item_created_at,invoice_bill_item_updated_at,invoice_created_at,invoice_updated_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='invoicemgmt.invoice_bill_item_list_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:invoice_bill_item_id,INVOICE_ID:invoice_id,INVOICE_SEQUENCE_NUMBER:invoice_sequence_number,STUDENT_ID:student_id,BILL_ITEM_SEQUENCE_NUMBER:bill_item_sequence_number,INVOICE_BILL_ITEM_CREATED_AT:invoice_bill_item_created_at,INVOICE_BILL_ITEM_UPDATED_AT:invoice_bill_item_updated_at,INVOICE_CREATED_AT:invoice_created_at,INVOICE_UPDATED_AT:invoice_updated_at',
      'pk.fields'='invoice_bill_item_id'
);
