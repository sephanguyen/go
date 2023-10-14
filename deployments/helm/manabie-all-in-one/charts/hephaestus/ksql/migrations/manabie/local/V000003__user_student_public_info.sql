set 'auto.offset.reset' = 'earliest';

DROP STREAM IF EXISTS STUDENT_PUBLIC_INFO_V3;

DROP STREAM IF EXISTS USERS_STREAM_FORMATED_V3;
CREATE STREAM IF NOT EXISTS USERS_STREAM_FORMATED_V4
    AS SELECT
           USERS_STREAM_ORIGIN_V3.AFTER->USER_ID as USER_ID,
           USERS_STREAM_ORIGIN_V3.AFTER->NAME as NAME,
           USERS_STREAM_ORIGIN_V3.AFTER->COUNTRY as COUNTRY,
           USERS_STREAM_ORIGIN_V3.AFTER->CREATED_AT as CREATED_AT,
           USERS_STREAM_ORIGIN_V3.AFTER->UPDATED_AT as UPDATED_AT
    FROM USERS_STREAM_ORIGIN_V3
    PARTITION BY AFTER->USER_ID
    EMIT CHANGES;
CREATE TABLE IF NOT EXISTS USERS_TABLE_FORMATED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='USERS_STREAM_FORMATED_V4', value_format='AVRO');


DROP STREAM IF EXISTS STUDENT_STREAM_FORMATED_V3;
CREATE STREAM IF NOT EXISTS STUDENT_STREAM_FORMATED_V4
    AS SELECT
           STUDENT_STREAM_ORIGIN_V3.AFTER->STUDENT_ID as STUDENT_ID,
           STUDENT_STREAM_ORIGIN_V3.AFTER->CURRENT_GRADE as CURRENT_GRADE,
           STUDENT_STREAM_ORIGIN_V3.AFTER->GRADE_ID as GRADE_ID
       FROM STUDENT_STREAM_ORIGIN_V3
       PARTITION BY AFTER->STUDENT_ID
       EMIT CHANGES;
CREATE TABLE IF NOT EXISTS STUDENTS_TABLE_FORMATED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='STUDENT_STREAM_FORMATED_V4', value_format='AVRO');


CREATE TABLE IF NOT EXISTS STUDENT_PUBLIC_INFO_V4 AS
SELECT
    USERS_TABLE_FORMATED_V1.KEY AS USER_ID,
    USERS_TABLE_FORMATED_V1.NAME as NAME,
    STUDENTS_TABLE_FORMATED_V1.CURRENT_GRADE as CURRENT_GRADE,
    STUDENTS_TABLE_FORMATED_V1.GRADE_ID as GRADE_ID,
    USERS_TABLE_FORMATED_V1.CREATED_AT as CREATED_AT,
    USERS_TABLE_FORMATED_V1.UPDATED_AT as UPDATED_AT
FROM USERS_TABLE_FORMATED_V1
JOIN STUDENTS_TABLE_FORMATED_V1 ON
    USERS_TABLE_FORMATED_V1.KEY = STUDENTS_TABLE_FORMATED_V1.KEY;



DROP CONNECTOR IF EXISTS SINK_BASIC_USER_INFO_NEW_V4;
CREATE SINK CONNECTOR SINK_BASIC_USER_INFO_NEW_V6 WITH (
    'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
    'transforms.unwrap.delete.handling.mode'='drop',
    'tasks.max'='1',
    'topics'='STUDENT_PUBLIC_INFO_V4',
    'fields.whitelist'='user_id,name,current_grade,grade_id,created_at,updated_at',
    'key.converter'='org.apache.kafka.connect.storage.StringConverter',
    'value.converter'='io.confluent.connect.avro.AvroConverter',
    'value.converter.schema.registry.url'='http://cp-schema-registry:8081',
    'delete.enabled'='false',
    'transforms.unwrap.drop.tombstones'='true',
    'auto.create'='false',
    'connection.url'='${file:/config/kafka-connect-config.properties:bob_url}',
    'insert.mode'='upsert',
    'table.name.format'='public.user_basic_info',
    'pk.mode'='record_key',
    'transforms'='RenameField',
    'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
    'transforms.RenameField.renames'='USER_ID:user_id,NAME:name,CURRENT_GRADE:current_grade,GRADE_ID:grade_id,CREATED_AT:created_at,UPDATED_AT:updated_at',
    'pk.fields'='user_id'
);
