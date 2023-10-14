SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS COMPANY_DETAIL_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.company_detail', value_format='AVRO');

CREATE STREAM IF NOT EXISTS COMPANY_DETAIL_STREAM_FORMATTED_V1
    AS SELECT
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_DETAIL_ID AS KEY,
        AS_VALUE(COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_DETAIL_ID) AS COMPANY_DETAIL_ID,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_NAME AS COMPANY_NAME,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_ADDRESS AS COMPANY_ADDRESS,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_PHONE_NUMBER AS COMPANY_PHONE_NUMBER,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_LOGO_URL AS COMPANY_LOGO_URL,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM COMPANY_DETAIL_STREAM_ORIGIN_V1
    WHERE COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY COMPANY_DETAIL_STREAM_ORIGIN_V1.AFTER->COMPANY_DETAIL_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS COMPANY_DETAIL_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}COMPANY_DETAIL_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS COMPANY_DETAIL_V1
AS SELECT
    COMPANY_DETAIL_TABLE_FORMATTED_V1.KEY as ID,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.COMPANY_DETAIL_ID as COMPANY_DETAIL_ID,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.COMPANY_NAME AS COMPANY_NAME,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.COMPANY_ADDRESS AS COMPANY_ADDRESS,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.COMPANY_PHONE_NUMBER AS COMPANY_PHONE_NUMBER,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.COMPANY_LOGO_URL AS COMPANY_LOGO_URL,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.CREATED_AT AS CREATED_AT,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.UPDATED_AT AS UPDATED_AT,
    COMPANY_DETAIL_TABLE_FORMATTED_V1.DELETED_AT AS DELETED_AT
FROM COMPANY_DETAIL_TABLE_FORMATTED_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_COMPANY_DETAIL_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}COMPANY_DETAIL_V1',
      'fields.whitelist'='company_detail_key,company_detail_id,company_name,company_address,company_phone_number,company_logo_url,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.company_detail',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:company_detail_key,COMPANY_DETAIL_ID:company_detail_id,COMPANY_NAME:company_name,COMPANY_ADDRESS:company_address,COMPANY_PHONE_NUMBER:company_phone_number,COMPANY_LOGO_URL:company_logo_url,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='company_detail_id'
);
