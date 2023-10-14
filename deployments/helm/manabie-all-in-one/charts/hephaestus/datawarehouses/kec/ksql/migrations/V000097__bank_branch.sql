SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS BANK_BRANCH_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.invoicemgmt.bank_branch', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BANK_BRANCH_STREAM_FORMATTED_V1
    AS SELECT
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_ID AS KEY,
        AS_VALUE(BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_ID) AS BANK_BRANCH_ID,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_ID AS BANK_ID,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_CODE AS BANK_BRANCH_CODE,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_NAME AS BANK_BRANCH_NAME,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_PHONETIC_NAME AS BANK_BRANCH_PHONETIC_NAME,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS BANK_BRANCH_IS_ARCHIVED,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS BANK_BRANCH_CREATED_AT,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS BANK_BRANCH_UPDATED_AT,
        BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS BANK_BRANCH_DELETED_AT
    FROM BANK_BRANCH_STREAM_ORIGIN_V1
    WHERE BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY BANK_BRANCH_STREAM_ORIGIN_V1.AFTER->BANK_BRANCH_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BANK_BRANCH_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BANK_BRANCH_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS BANK_BRANCH_V1
AS SELECT
    BANK_BRANCH_TABLE_FORMATTED_V1.KEY as ID,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_ID as BANK_BRANCH_ID,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_ID as BANK_ID,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_CODE AS BANK_BRANCH_CODE,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_NAME AS BANK_BRANCH_NAME,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_PHONETIC_NAME AS BANK_BRANCH_PHONETIC_NAME,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_IS_ARCHIVED AS BANK_BRANCH_IS_ARCHIVED,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_CREATED_AT AS BANK_BRANCH_CREATED_AT,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_UPDATED_AT AS BANK_BRANCH_UPDATED_AT,
    BANK_BRANCH_TABLE_FORMATTED_V1.BANK_BRANCH_DELETED_AT AS BANK_BRANCH_DELETED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_CODE AS BANK_CODE,
    BANK_TABLE_FORMATTED_V1.BANK_NAME AS BANK_NAME,
    BANK_TABLE_FORMATTED_V1.BANK_NAME_PHONETIC AS BANK_NAME_PHONETIC,
    BANK_TABLE_FORMATTED_V1.BANK_IS_ARCHIVED AS BANK_IS_ARCHIVED,
    BANK_TABLE_FORMATTED_V1.BANK_CREATED_AT AS BANK_CREATED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_UPDATED_AT AS BANK_UPDATED_AT,
    BANK_TABLE_FORMATTED_V1.BANK_DELETED_AT AS BANK_DELETED_AT
FROM BANK_BRANCH_TABLE_FORMATTED_V1
JOIN BANK_TABLE_FORMATTED_V1
ON BANK_BRANCH_TABLE_FORMATTED_V1.BANK_ID = BANK_TABLE_FORMATTED_V1.KEY;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_BANK_BRANCH_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}BANK_BRANCH_V1',
      'fields.whitelist'='bank_branch_key,bank_branch_id,bank_id,bank_branch_code,bank_branch_name,bank_branch_phonetic_name,bank_branch_is_archived,bank_branch_created_at,bank_branch_updated_at,bank_branch_deleted_at,bank_code,bank_name,bank_name_phonetic,bank_is_archived,bank_created_at,bank_updated_at,bank_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.bank_branch',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:bank_branch_key,BANK_BRANCH_ID:bank_branch_id,BANK_ID:bank_id,BANK_BRANCH_CODE:bank_branch_code,BANK_BRANCH_NAME:bank_branch_name,BANK_BRANCH_PHONETIC_NAME:bank_branch_phonetic_name,BANK_BRANCH_IS_ARCHIVED:bank_branch_is_archived,BANK_BRANCH_CREATED_AT:bank_branch_created_at,BANK_BRANCH_UPDATED_AT:bank_branch_updated_at,BANK_BRANCH_DELETED_AT:bank_branch_deleted_at,BANK_CODE:bank_code,BANK_NAME:bank_name,BANK_NAME_PHONETIC:bank_name_phonetic,BANK_IS_ARCHIVED:bank_is_archived,BANK_CREATED_AT:bank_created_at,BANK_UPDATED_AT:bank_updated_at,BANK_DELETED_AT:bank_deleted_at',
      'pk.fields'='bank_branch_id'
);
