SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS BANK_MAPPING_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.bank_mapping', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BANK_MAPPING_STREAM_FORMATTED_V1
    AS SELECT
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->BANK_MAPPING_ID AS KEY,
        AS_VALUE(BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->BANK_MAPPING_ID) AS BANK_MAPPING_ID,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->BANK_ID AS BANK_ID,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->PARTNER_BANK_ID AS PARTNER_BANK_ID,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->REMARKS AS BANK_MAPPING_REMARKS,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS BANK_MAPPING_IS_ARCHIVED,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS BANK_MAPPING_CREATED_AT,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS BANK_MAPPING_UPDATED_AT,
        BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS BANK_MAPPING_DELETED_AT
    FROM BANK_MAPPING_STREAM_ORIGIN_V1
    WHERE BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY BANK_MAPPING_STREAM_ORIGIN_V1.AFTER->BANK_MAPPING_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BANK_MAPPING_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BANK_MAPPING_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS BANK_MAPPING_WITH_BANK_V1
AS SELECT
    BANK_MAPPING_TABLE_FORMATTED_V1.KEY as ROW_KEY,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_ID as BANK_MAPPING_ID,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_ID as BANK_ID,
    BANK_MAPPING_TABLE_FORMATTED_V1.PARTNER_BANK_ID AS PARTNER_BANK_ID,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_REMARKS AS BANK_MAPPING_REMARKS,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_IS_ARCHIVED AS BANK_MAPPING_IS_ARCHIVED,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_CREATED_AT AS BANK_MAPPING_CREATED_AT,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_UPDATED_AT AS BANK_MAPPING_UPDATED_AT,
    BANK_MAPPING_TABLE_FORMATTED_V1.BANK_MAPPING_DELETED_AT AS BANK_MAPPING_DELETED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_CODE AS BANK_BANK_CODE,
    BANK_TABLE_FORMATTED_V1.BANK_NAME AS BANK_BANK_NAME,
    BANK_TABLE_FORMATTED_V1.BANK_NAME_PHONETIC AS BANK_BANK_NAME_PHONETIC,
    BANK_TABLE_FORMATTED_V1.BANK_IS_ARCHIVED AS BANK_IS_ARCHIVED,
    BANK_TABLE_FORMATTED_V1.BANK_CREATED_AT AS BANK_CREATED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_UPDATED_AT AS BANK_UPDATED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_DELETED_AT AS BANK_DELETED_AT
FROM BANK_MAPPING_TABLE_FORMATTED_V1
JOIN BANK_TABLE_FORMATTED_V1
    ON BANK_MAPPING_TABLE_FORMATTED_V1.BANK_ID = BANK_TABLE_FORMATTED_V1.KEY;

CREATE TABLE IF NOT EXISTS BANK_MAPPING_WITH_BANK_AND_PARTNER_BANK_V1
AS SELECT
    BANK_MAPPING_WITH_BANK_V1.ROW_KEY as ROW_KEY,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_ID as BANK_MAPPING_ID,
    BANK_MAPPING_WITH_BANK_V1.BANK_ID as BANK_ID,
    BANK_MAPPING_WITH_BANK_V1.PARTNER_BANK_ID AS PARTNER_BANK_ID,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_REMARKS AS BANK_MAPPING_REMARKS,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_IS_ARCHIVED AS BANK_MAPPING_IS_ARCHIVED,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_CREATED_AT AS BANK_MAPPING_CREATED_AT,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_UPDATED_AT AS BANK_MAPPING_UPDATED_AT,
    BANK_MAPPING_WITH_BANK_V1.BANK_MAPPING_DELETED_AT AS BANK_MAPPING_DELETED_AT,
    BANK_MAPPING_WITH_BANK_V1.BANK_BANK_CODE AS BANK_BANK_CODE,
    BANK_MAPPING_WITH_BANK_V1.BANK_BANK_NAME AS BANK_BANK_NAME,
    BANK_MAPPING_WITH_BANK_V1.BANK_BANK_NAME_PHONETIC AS BANK_BANK_NAME_PHONETIC,
    BANK_MAPPING_WITH_BANK_V1.BANK_IS_ARCHIVED AS BANK_IS_ARCHIVED,
    BANK_MAPPING_WITH_BANK_V1.BANK_CREATED_AT AS BANK_CREATED_AT,
    BANK_MAPPING_WITH_BANK_V1.BANK_UPDATED_AT AS BANK_UPDATED_AT,
    BANK_MAPPING_WITH_BANK_V1.BANK_DELETED_AT AS BANK_DELETED_AT,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_BANK_NUMBER AS PARTNER_BANK_BANK_NUMBER,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_BANK_NAME AS PARTNER_BANK_BANK_NAME,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_BANK_BRANCH_NUMBER AS PARTNER_BANK_BANK_BRANCH_NUMBER,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_BANK_BRANCH_NAME AS PARTNER_BANK_BANK_BRANCH_NAME,
    PARTNER_BANK_TABLE_FORMATTED_V1.DEPOSIT_ITEMS AS DEPOSIT_ITEMS,
    PARTNER_BANK_TABLE_FORMATTED_V1.ACCOUNT_NUMBER AS ACCOUNT_NUMBER,
    PARTNER_BANK_TABLE_FORMATTED_V1.CONSIGNOR_CODE AS CONSIGNOR_CODE,
    PARTNER_BANK_TABLE_FORMATTED_V1.CONSIGNOR_NAME AS CONSIGNOR_NAME,
    PARTNER_BANK_TABLE_FORMATTED_V1.IS_DEFAULT AS IS_DEFAULT,
    PARTNER_BANK_TABLE_FORMATTED_V1.RECORD_LIMIT AS RECORD_LIMIT,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_REMARKS AS PARTNER_BANK_REMARKS,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_IS_ARCHIVED AS PARTNER_BANK_IS_ARCHIVED,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_CREATED_AT AS PARTNER_BANK_CREATED_AT,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_UPDATED_AT AS PARTNER_BANK_UPDATED_AT,
    PARTNER_BANK_TABLE_FORMATTED_V1.PARTNER_BANK_DELETED_AT AS PARTNER_BANK_DELETED_AT
FROM BANK_MAPPING_WITH_BANK_V1
JOIN PARTNER_BANK_TABLE_FORMATTED_V1
    ON BANK_MAPPING_WITH_BANK_V1.PARTNER_BANK_ID = PARTNER_BANK_TABLE_FORMATTED_V1.KEY;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_BANK_MAPPING_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}BANK_MAPPING_WITH_BANK_AND_PARTNER_BANK_V1',
      'fields.whitelist'='bank_mapping_key,bank_mapping_id,bank_id,partner_bank_id,bank_mapping_remarks,bank_mapping_is_archived,bank_mapping_created_at,bank_mapping_updated_at,bank_mapping_deleted_at,bank_bank_code,bank_bank_name,bank_bank_name_phonetic,bank_is_archived,bank_created_at,bank_updated_at,bank_deleted_at,partner_bank_bank_number,partner_bank_bank_name,partner_bank_bank_branch_number,partner_bank_bank_branch_name,deposit_items,account_number,consignor_code,consignor_name,is_default,record_limit,partner_bank_remarks,partner_bank_is_archived,partner_bank_created_at,partner_bank_updated_at,partner_bank_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.bank_mapping',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ROW_KEY:bank_mapping_key,BANK_MAPPING_ID:bank_mapping_id,BANK_ID:bank_id,PARTNER_BANK_ID:partner_bank_id,BANK_MAPPING_REMARKS:bank_mapping_remarks,BANK_MAPPING_IS_ARCHIVED:bank_mapping_is_archived,BANK_MAPPING_CREATED_AT:bank_mapping_created_at,BANK_MAPPING_UPDATED_AT:bank_mapping_updated_at,BANK_MAPPING_DELETED_AT:bank_mapping_deleted_at,BANK_BANK_CODE:bank_bank_code,BANK_BANK_NAME:bank_bank_name,BANK_BANK_NAME_PHONETIC:bank_bank_name_phonetic,BANK_IS_ARCHIVED:bank_is_archived,BANK_CREATED_AT:bank_created_at,BANK_UPDATED_AT:bank_updated_at,BANK_DELETED_AT:bank_deleted_at,PARTNER_BANK_BANK_NUMBER:partner_bank_bank_number,PARTNER_BANK_BANK_NAME:partner_bank_bank_name,PARTNER_BANK_BANK_BRANCH_NUMBER:partner_bank_bank_branch_number,PARTNER_BANK_BANK_BRANCH_NAME:partner_bank_bank_branch_name,DEPOSIT_ITEMS:deposit_items,ACCOUNT_NUMBER:account_number,CONSIGNOR_CODE:consignor_code,CONSIGNOR_NAME:consignor_name,IS_DEFAULT:is_default,RECORD_LIMIT:record_limit,PARTNER_BANK_REMARKS:partner_bank_remarks,PARTNER_BANK_IS_ARCHIVED:partner_bank_is_archived,PARTNER_BANK_CREATED_AT:partner_bank_created_at,PARTNER_BANK_UPDATED_AT:partner_bank_updated_at,PARTNER_BANK_DELETED_AT:partner_bank_deleted_at',
      'pk.fields'='bank_mapping_id'
);
