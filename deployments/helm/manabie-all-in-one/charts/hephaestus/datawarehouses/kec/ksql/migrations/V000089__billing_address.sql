SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS BILLING_ADDRESS_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.billing_address', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BILLING_ADDRESS_STREAM_FORMATTED_V1
    AS SELECT
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->BILLING_ADDRESS_ID AS KEY,
        AS_VALUE(BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->BILLING_ADDRESS_ID) AS BILLING_ADDRESS_ID,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->STUDENT_PAYMENT_DETAIL_ID AS STUDENT_PAYMENT_DETAIL_ID,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->POSTAL_CODE AS POSTAL_CODE,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->CITY AS CITY,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->STREET1 AS STREET1,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->STREET2 AS STREET2,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->PREFECTURE_CODE AS PREFECTURE_CODE,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM BILLING_ADDRESS_STREAM_ORIGIN_V1
    WHERE BILLING_ADDRESS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->BILLING_ADDRESS_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BILLING_ADDRESS_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BILLING_ADDRESS_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS BILLING_ADDRESS_PUBLIC_INFO_V1
AS SELECT
    BILLING_ADDRESS_TABLE_FORMATTED_V1.KEY AS BILLING_ADDRESS_ID,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.USER_ID AS USER_ID,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.STUDENT_PAYMENT_DETAIL_ID AS STUDENT_PAYMENT_DETAIL_ID,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.POSTAL_CODE AS POSTAL_CODE,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.CITY AS CITY,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.STREET1 AS STREET1,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.STREET2 AS STREET2,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.PREFECTURE_CODE AS PREFECTURE_CODE,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.CREATED_AT AS CREATED_AT,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.UPDATED_AT AS UPDATED_AT,
    BILLING_ADDRESS_TABLE_FORMATTED_V1.DELETED_AT AS DELETED_AT
FROM BILLING_ADDRESS_TABLE_FORMATTED_V1;


CREATE SINK CONNECTOR IF NOT EXISTS SINK_BILLING_ADDRESS_PUBLIC_INFO_V1_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}BILLING_ADDRESS_PUBLIC_INFO_V1',
      'fields.whitelist'='billing_address_id,user_id,student_payment_detail_id,postal_code,city,street1,street2,prefecture_code,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.billing_address',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='BILLING_ADDRESS_ID:billing_address_id,USER_ID:user_id,STUDENT_PAYMENT_DETAIL_ID:student_payment_detail_id,POSTAL_CODE:postal_code,CITY:city,STREET1:street1,STREET2:street2,PREFECTURE_CODE:prefecture_code,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='billing_address_id'
);
