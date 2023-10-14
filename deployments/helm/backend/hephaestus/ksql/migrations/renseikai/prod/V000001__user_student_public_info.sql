set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS USERS_STREAM_ORIGIN   WITH (kafka_topic='prod.renseikai.bob.public.users', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USERS_STREAM_FORMATED
    AS SELECT
         USERS_STREAM_ORIGIN.AFTER->USER_ID as USER_ID,
         USERS_STREAM_ORIGIN.AFTER->NAME as NAME,
         USERS_STREAM_ORIGIN.AFTER->COUNTRY as COUNTRY,
         USERS_STREAM_ORIGIN.AFTER->CREATED_AT as CREATED_AT,
         USERS_STREAM_ORIGIN.AFTER->UPDATED_AT as UPDATED_AT
    FROM USERS_STREAM_ORIGIN;

CREATE  STREAM IF NOT EXISTS STUDENT_STREAM_ORIGIN  WITH (kafka_topic='prod.renseikai.bob.public.students', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENT_STREAM_FORMATED
    AS SELECT
              STUDENT_STREAM_ORIGIN.AFTER-> STUDENT_ID as STUDENT_ID,
              STUDENT_STREAM_ORIGIN.AFTER->CURRENT_GRADE as CURRENT_GRADE,
              STUDENT_STREAM_ORIGIN.AFTER->GRADE_ID as GRADE_ID
       FROM STUDENT_STREAM_ORIGIN;

CREATE STREAM IF NOT EXISTS STUDENT_PUBLIC_INFO AS
SELECT
    USERS_STREAM_FORMATED.USER_ID AS USER_ID,
    USERS_STREAM_FORMATED.NAME as NAME,
    STUDENT_STREAM_FORMATED.CURRENT_GRADE as CURRENT_GRADE,
    STUDENT_STREAM_FORMATED.GRADE_ID as GRADE_ID,
    USERS_STREAM_FORMATED.CREATED_AT as CREATED_AT,
    USERS_STREAM_FORMATED.UPDATED_AT as UPDATED_AT
FROM USERS_STREAM_FORMATED JOIN STUDENT_STREAM_FORMATED WITHIN 2 HOURS  ON USERS_STREAM_FORMATED.USER_ID = STUDENT_STREAM_FORMATED.STUDENT_ID;
-- created manually
CREATE SINK CONNECTOR IF NOT EXISTS SINK_BASIC_USER_INFO_NEW WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='STUDENT_PUBLIC_INFO',
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
