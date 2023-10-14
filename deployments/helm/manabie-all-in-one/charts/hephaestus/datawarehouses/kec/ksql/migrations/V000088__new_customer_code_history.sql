SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.new_customer_code_history', value_format='AVRO');

CREATE STREAM IF NOT EXISTS NEW_CUSTOMER_CODE_HISTORY_STREAM_FORMATTED_V1
    AS SELECT
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->NEW_CUSTOMER_CODE_HISTORY_ID AS KEY,
        AS_VALUE(NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->NEW_CUSTOMER_CODE_HISTORY_ID) AS NEW_CUSTOMER_CODE_HISTORY_ID,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->NEW_CUSTOMER_CODE AS NEW_CUSTOMER_CODE,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->BANK_ACCOUNT_NUMBER AS BANK_ACCOUNT_NUMBER,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1
    WHERE NEW_CUSTOMER_CODE_HISTORY_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->NEW_CUSTOMER_CODE_HISTORY_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}NEW_CUSTOMER_CODE_HISTORY_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS NEW_CUSTOMER_CODE_HISTORY_PUBLIC_INFO_V1
AS SELECT
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.KEY AS NEW_CUSTOMER_CODE_HISTORY_ID,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.NEW_CUSTOMER_CODE AS NEW_CUSTOMER_CODE,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.STUDENT_ID AS STUDENT_ID,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.BANK_ACCOUNT_NUMBER AS BANK_ACCOUNT_NUMBER,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.CREATED_AT AS CREATED_AT,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.UPDATED_AT AS UPDATED_AT,
    NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1.DELETED_AT AS DELETED_AT
FROM NEW_CUSTOMER_CODE_HISTORY_TABLE_FORMATTED_V1;


CREATE SINK CONNECTOR IF NOT EXISTS SINK_NEW_CUSTOMER_CODE_HISTORY_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}NEW_CUSTOMER_CODE_HISTORY_PUBLIC_INFO_V1',
      'fields.whitelist'='new_customer_code_history_id,new_customer_code,student_id,bank_account_number,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.new_customer_code_history',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='NEW_CUSTOMER_CODE_HISTORY_ID:new_customer_code_history_id,NEW_CUSTOMER_CODE:new_customer_code,STUDENT_ID:student_id,BANK_ACCOUNT_NUMBER:bank_account_number,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='new_customer_code_history_id'
);
