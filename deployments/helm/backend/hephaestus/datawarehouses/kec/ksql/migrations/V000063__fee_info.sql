SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS FEE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.fee', value_format='AVRO');

CREATE STREAM IF NOT EXISTS FEE_STREAM_FORMATTED_V1
    AS SELECT
        FEE_STREAM_ORIGIN_V1.AFTER->FEE_ID AS KEY,
        AS_VALUE(FEE_STREAM_ORIGIN_V1.AFTER->FEE_ID) AS FEE_ID,
        FEE_STREAM_ORIGIN_V1.AFTER->FEE_TYPE AS FEE_TYPE,
        CAST(NULL AS VARCHAR) AS PACKAGE_CREATED_AT,
        CAST(NULL AS VARCHAR) AS PACKAGE_UPDATED_AT,
        CAST(NULL AS VARCHAR) AS PACKAGE_DELETED_AT
    FROM FEE_STREAM_ORIGIN_V1
    WHERE FEE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->FEE_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS FEE_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}FEE_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS FEE_PUBLIC_INFO_V1
AS SELECT
    PRODUCT_TABLE_FORMATTED_V1.KEY AS ROW_KEY,
    PRODUCT_TABLE_FORMATTED_V1.PRODUCT_ID AS PRODUCT_ID,
    PRODUCT_TABLE_FORMATTED_V1.NAME AS NAME,
    PRODUCT_TABLE_FORMATTED_V1.PRODUCT_TYPE AS PRODUCT_TYPE,
    PRODUCT_TABLE_FORMATTED_V1.TAX_ID AS TAX_ID,
    PRODUCT_TABLE_FORMATTED_V1.AVAILABLE_FROM AS AVAILABLE_FROM,
    PRODUCT_TABLE_FORMATTED_V1.AVAILABLE_UNTIL AS AVAILABLE_UNTIL,
    PRODUCT_TABLE_FORMATTED_V1.REMARKS AS REMARKS,
    PRODUCT_TABLE_FORMATTED_V1.CUSTOM_BILLING_PERIOD AS CUSTOM_BILLING_PERIOD,
    PRODUCT_TABLE_FORMATTED_V1.BILLING_SCHEDULE_ID AS BILLING_SCHEDULE_ID,
    PRODUCT_TABLE_FORMATTED_V1.DISABLE_PRO_RATING_FLAG AS DISABLE_PRO_RATING_FLAG,
    PRODUCT_TABLE_FORMATTED_V1.IS_ARCHIVED AS IS_ARCHIVED,
    PRODUCT_TABLE_FORMATTED_V1.IS_UNIQUE AS IS_UNIQUE,
    PRODUCT_TABLE_FORMATTED_V1.PRODUCT_CREATED_AT AS PRODUCT_CREATED_AT,
    PRODUCT_TABLE_FORMATTED_V1.PRODUCT_UPDATED_AT AS PRODUCT_UPDATED_AT,
    PRODUCT_TABLE_FORMATTED_V1.PRODUCT_DELETED_AT AS PRODUCT_DELETED_AT,
    AS_VALUE(FEE_TABLE_FORMATTED_V1.KEY) AS FEE_ID,
    FEE_TABLE_FORMATTED_V1.FEE_TYPE AS FEE_TYPE,
    FEE_TABLE_FORMATTED_V1.PACKAGE_CREATED_AT AS PACKAGE_CREATED_AT,
    FEE_TABLE_FORMATTED_V1.PACKAGE_UPDATED_AT AS PACKAGE_UPDATED_AT,
    FEE_TABLE_FORMATTED_V1.PACKAGE_DELETED_AT AS PACKAGE_DELETED_AT
FROM PRODUCT_TABLE_FORMATTED_V1
JOIN FEE_TABLE_FORMATTED_V1
ON PRODUCT_TABLE_FORMATTED_V1.PRODUCT_ID = FEE_TABLE_FORMATTED_V1.KEY;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_FEE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}FEE_PUBLIC_INFO_V1',
      'fields.whitelist'='fee_id,fee_type,fee_created_at,fee_updated_at,package_deleted_at,name,product_type,tax_id,remarks,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,is_archived,created_at,updated_at,is_unique,product_created_at,product_updated_at,product_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.fee',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='FEE_ID:fee_id,FEE_TYPE:fee_type,FEE_CREATED_AT:fee_created_at,FEE_UPDATED_AT:fee_updated_at,FEE_DELETED_AT:fee_deleted_at,NAME:name,PRODUCT_TYPE:product_type,TAX_ID:tax_id,AVAILABLE_FROM:available_from,AVAILABLE_UNTIL:available_until,REMARKS:remarks,CUSTOM_BILLING_PERIOD:custom_billing_period,BILLING_SCHEDULE_ID:billing_schedule_id,DISABLE_PRO_RATING_FLAG:disable_pro_rating_flag,IS_ARCHIVED:is_archived,IS_UNIQUE:is_unique,PRODUCT_CREATED_AT:product_created_at,PRODUCT_UPDATED_AT:product_updated_at,PRODUCT_DELETED_AT:product_deleted_at',
      'pk.fields'='fee_id'
);
