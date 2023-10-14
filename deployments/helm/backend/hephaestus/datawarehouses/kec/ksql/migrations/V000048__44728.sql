SET 'auto.offset.reset' = 'earliest';

DROP STREAM IF EXISTS SCHOOL_COURSE_SCHOOL_INFO_PUBLIC_INFO_V1 DELETE TOPIC;
DROP STREAM IF EXISTS SCHOOL_COURSE_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS SCHOOL_INFO_STREAM_FORMATED_V1 DELETE TOPIC;


CREATE STREAM IF NOT EXISTS SCHOOL_COURSE_STREAM_FORMATED_V1
    AS SELECT
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_ID AS KEY,
        AS_VALUE(SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_ID) AS SCHOOL_COURSE_ID,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_NAME AS SCHOOL_COURSE_NAME,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_NAME_PHONETIC AS SCHOOL_COURSE_NAME_PHONETIC,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID AS SCHOOL_ID,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_PARTNER_ID AS SCHOOL_COURSE_PARTNER_ID,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM SCHOOL_COURSE_STREAM_ORIGIN_V1
    WHERE SCHOOL_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->SCHOOL_COURSE_ID
    EMIT CHANGES;
CREATE TABLE IF NOT EXISTS SCHOOL_COURSE_TABLE_FORMATED_V1 (ID VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}SCHOOL_COURSE_STREAM_FORMATED_V1', value_format='AVRO');

CREATE STREAM IF NOT EXISTS SCHOOL_INFO_STREAM_FORMATED_V1
    AS SELECT
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID AS KEY,
        AS_VALUE(SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID) AS SCHOOL_ID,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_NAME AS SCHOOL_NAME,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_NAME_PHONETIC AS SCHOOL_NAME_PHONETIC,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_LEVEL_ID AS SCHOOL_LEVEL_ID,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->SCHOOL_PARTNER_ID AS SCHOOL_PARTNER_ID,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->ADDRESS AS ADDRESS,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        '' AS SCHOOL_COURSE_ID

    FROM SCHOOL_INFO_STREAM_ORIGIN_V1
    WHERE SCHOOL_INFO_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->SCHOOL_ID
    EMIT CHANGES;
CREATE TABLE IF NOT EXISTS SCHOOL_INFO_TABLE_FORMATED_V1 (ID VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}SCHOOL_INFO_STREAM_FORMATED_V1', value_format='AVRO');


CREATE TABLE IF NOT EXISTS SCHOOL_COURSE_SCHOOL_INFO_PUBLIC_INFO_V1 
AS SELECT
    SCHOOL_COURSE_TABLE_FORMATED_V1.ID AS SCHOOL_COURSE_SCHOOL_INFO_ID,
    SCHOOL_COURSE_TABLE_FORMATED_V1.SCHOOL_COURSE_ID AS SCHOOL_COURSE_ID,
    SCHOOL_COURSE_TABLE_FORMATED_V1.SCHOOL_COURSE_NAME AS SCHOOL_COURSE_NAME,
    SCHOOL_COURSE_TABLE_FORMATED_V1.SCHOOL_COURSE_NAME_PHONETIC AS SCHOOL_COURSE_NAME_PHONETIC,
    SCHOOL_COURSE_TABLE_FORMATED_V1.IS_ARCHIVED AS SCHOOL_COURSE_IS_ARCHIVED,
    SCHOOL_COURSE_TABLE_FORMATED_V1.SCHOOL_COURSE_PARTNER_ID AS SCHOOL_COURSE_PARTNER_ID,
    SCHOOL_COURSE_TABLE_FORMATED_V1.CREATED_AT AS SCHOOL_COURSE_CREATED_AT,
    SCHOOL_COURSE_TABLE_FORMATED_V1.UPDATED_AT AS SCHOOL_COURSE_UPDATED_AT,
    SCHOOL_COURSE_TABLE_FORMATED_V1.DELETED_AT AS SCHOOL_COURSE_DELETED_AT,

    SCHOOL_INFO_TABLE_FORMATED_V1.ID AS SCHOOL_ID,
    SCHOOL_INFO_TABLE_FORMATED_V1.SCHOOL_NAME AS SCHOOL_NAME,
    SCHOOL_INFO_TABLE_FORMATED_V1.SCHOOL_NAME_PHONETIC AS SCHOOL_NAME_PHONETIC,
    SCHOOL_INFO_TABLE_FORMATED_V1.SCHOOL_LEVEL_ID AS SCHOOL_LEVEL_ID,
    SCHOOL_INFO_TABLE_FORMATED_V1.IS_ARCHIVED AS SCHOOL_INFO_IS_ARCHIVED,
    SCHOOL_INFO_TABLE_FORMATED_V1.SCHOOL_PARTNER_ID AS SCHOOL_PARTNER_ID,
    SCHOOL_INFO_TABLE_FORMATED_V1.ADDRESS AS ADDRESS,
    SCHOOL_INFO_TABLE_FORMATED_V1.CREATED_AT AS SCHOOL_INFO_CREATED_AT,
    SCHOOL_INFO_TABLE_FORMATED_V1.UPDATED_AT AS SCHOOL_INFO_UPDATED_AT,
    SCHOOL_INFO_TABLE_FORMATED_V1.DELETED_AT AS SCHOOL_INFO_DELETED_AT

FROM SCHOOL_COURSE_TABLE_FORMATED_V1 
JOIN SCHOOL_INFO_TABLE_FORMATED_V1 ON SCHOOL_INFO_TABLE_FORMATED_V1.ID = SCHOOL_COURSE_TABLE_FORMATED_V1.SCHOOL_ID;


DROP CONNECTOR IF EXISTS SCHOOL_COURSE_SCHOOL_INFO_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_SCHOOL_COURSE_SCHOOL_INFO_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHOOL_COURSE_SCHOOL_INFO_PUBLIC_INFO_V1',
      'fields.whitelist'='school_course_id,school_course_name,school_course_name_phonetic,school_id,school_course_is_archived,school_course_partner_id,school_course_created_at,school_course_updated_at,school_course_deleted_at,school_name,school_name_phonetic,school_level_id,school_info_is_archived,school_info_created_at,school_info_updated_at,school_info_deleted_at,school_partner_id,address',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.school_course_school_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SCHOOL_COURSE_ID:school_course_id,SCHOOL_COURSE_NAME:school_course_name,SCHOOL_COURSE_NAME_PHONETIC:school_course_name_phonetic,SCHOOL_ID:school_id,SCHOOL_COURSE_IS_ARCHIVED:school_course_is_archived,SCHOOL_COURSE_PARTNER_ID:school_course_partner_id,SCHOOL_COURSE_CREATED_AT:school_course_created_at,SCHOOL_COURSE_UPDATED_AT:school_course_updated_at,SCHOOL_COURSE_DELETED_AT:school_course_deleted_at,SCHOOL_NAME:school_name,SCHOOL_NAME_PHONETIC:school_name_phonetic,SCHOOL_LEVEL_ID:school_level_id,SCHOOL_INFO_IS_ARCHIVED:school_info_is_archived,SCHOOL_INFO_CREATED_AT:school_info_created_at,SCHOOL_INFO_UPDATED_AT:school_info_updated_at,SCHOOL_INFO_DELETED_AT:school_info_deleted_at,SCHOOL_PARTNER_ID:school_partner_id,ADDRESS:address',
      'pk.fields'='school_id,school_course_id'
);

CREATE SINK CONNECTOR IF NOT EXISTS SINK_SCHOOL_INFO_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHOOL_INFO_STREAM_FORMATED_V1',
      'fields.whitelist'='school_course_id,school_id,school_name,school_name_phonetic,school_level_id,school_info_is_archived,school_info_created_at,school_info_updated_at,school_info_deleted_at,school_partner_id,address',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.school_course_school_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SCHOOL_COURSE_ID:school_course_id,SCHOOL_ID:school_id,SCHOOL_NAME:school_name,SCHOOL_NAME_PHONETIC:school_name_phonetic,SCHOOL_LEVEL_ID:school_level_id,IS_ARCHIVED:school_info_is_archived,CREATED_AT:school_info_created_at,UPDATED_AT:school_info_updated_at,DELETED_AT:school_info_deleted_at,SCHOOL_PARTNER_ID:school_partner_id,ADDRESS:address',
      'pk.fields'='school_id,school_course_id'
);