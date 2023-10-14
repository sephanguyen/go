set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS USERS_STREAM_ORIGIN_V3   WITH (kafka_topic='stag.manabie.bob.public.users', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USERS_STREAM_FORMATED_V3
    AS SELECT
         USERS_STREAM_ORIGIN_V3.AFTER->USER_ID as USER_ID,
         USERS_STREAM_ORIGIN_V3.AFTER->NAME as NAME,
         USERS_STREAM_ORIGIN_V3.AFTER->COUNTRY as COUNTRY,
         USERS_STREAM_ORIGIN_V3.AFTER->CREATED_AT as CREATED_AT,
         USERS_STREAM_ORIGIN_V3.AFTER->UPDATED_AT as UPDATED_AT
    FROM USERS_STREAM_ORIGIN_V3;

CREATE  STREAM IF NOT EXISTS STUDENT_STREAM_ORIGIN_V3  WITH (kafka_topic='stag.manabie.bob.public.students', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENT_STREAM_FORMATED_V3
    AS SELECT
              STUDENT_STREAM_ORIGIN_V3.AFTER-> STUDENT_ID as STUDENT_ID,
              STUDENT_STREAM_ORIGIN_V3.AFTER->CURRENT_GRADE as CURRENT_GRADE,
              STUDENT_STREAM_ORIGIN_V3.AFTER->GRADE_ID as GRADE_ID
       FROM STUDENT_STREAM_ORIGIN_V3;

CREATE STREAM IF NOT EXISTS STUDENT_PUBLIC_INFO_V3 AS
SELECT
    USERS_STREAM_FORMATED_V3.USER_ID AS USER_ID,
    USERS_STREAM_FORMATED_V3.NAME as NAME,
    STUDENT_STREAM_FORMATED_V3.CURRENT_GRADE as CURRENT_GRADE,
    STUDENT_STREAM_FORMATED_V3.GRADE_ID as GRADE_ID,
    USERS_STREAM_FORMATED_V3.CREATED_AT as CREATED_AT,
    USERS_STREAM_FORMATED_V3.UPDATED_AT as UPDATED_AT
FROM USERS_STREAM_FORMATED_V3 JOIN STUDENT_STREAM_FORMATED_V3 WITHIN 2 HOURS  ON USERS_STREAM_FORMATED_V3.USER_ID = STUDENT_STREAM_FORMATED_V3.STUDENT_ID;
-- created manually
-- CREATE SINK CONNECTOR SINK_BASIC_USER_INFO_NEW_V4 WITH (
--       'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
--       'transforms.unwrap.delete.handling.mode'='drop',
--       'tasks.max'='1',
--       'topics'='STUDENT_PUBLIC_INFO_V3',
--       'fields.whitelist'='user_id,name,current_grade,grade_id,created_at,updated_at',
--       'key.converter'='org.apache.kafka.connect.storage.StringConverter',
--       'value.converter'='io.confluent.connect.avro.AvroConverter',
--       'value.converter.schema.registry.url'='http://cp-schema-registry:8081',
--       'delete.enabled'='false',
--       'transforms.unwrap.drop.tombstones'='true',
--       'auto.create'='false',
--       'connection.url'='${file:/config/kafka-connect-config.properties:bob_url}',
--       'insert.mode'='upsert',
--       'table.name.format'='public.user_basic_info',
--       'pk.mode'='record_key',
--       'transforms'='RenameField',
--       'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
--       'transforms.RenameField.renames'='USER_ID:user_id,NAME:name,CURRENT_GRADE:current_grade,GRADE_ID:grade_id,CREATED_AT:created_at,UPDATED_AT:updated_at',
--       'pk.fields'='user_id'
-- );
