
set 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS SINK_LESSONS_TEACHERS_PUBLIC_INFO;
DROP CONNECTOR IF EXISTS SINK_LESSONS_COURSES_PUBLIC_INFO;
DROP CONNECTOR IF EXISTS SINK_CLASSROOM_PUBLIC_INFO;

DROP STREAM IF EXISTS LESSONS_TEACHERS_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSONS_TEACHERS_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS LESSONS_COURSES_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSONS_COURSES_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS CLASSROOM_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS CLASSROOM_STREAM_ORIGIN_V1;

CREATE STREAM IF NOT EXISTS LESSONS_TEACHERS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.lessons_teachers', value_format='AVRO');
CREATE STREAM IF NOT EXISTS LESSONS_TEACHERS_STREAM_FORMATED_V1 AS 
    SELECT
        LESSONS_TEACHERS_STREAM_ORIGIN_V1.AFTER->LESSON_ID AS LESSON_ID,
        LESSONS_TEACHERS_STREAM_ORIGIN_V1.AFTER->TEACHER_ID AS STAFF_ID,
        LESSONS_TEACHERS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LESSONS_TEACHERS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM LESSONS_TEACHERS_STREAM_ORIGIN_V1
    WHERE LESSONS_TEACHERS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';


CREATE SINK CONNECTOR IF NOT EXISTS SINK_LESSONS_TEACHERS_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LESSONS_TEACHERS_STREAM_FORMATED_V1',
      'fields.whitelist'='lesson_id,staff_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.lessons_teachers',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,STAFF_ID:staff_id,CREATED_AT:created_at,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='lesson_id,staff_id'
);

CREATE STREAM IF NOT EXISTS LESSONS_COURSES_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.lessons_courses', value_format='AVRO');
CREATE STREAM IF NOT EXISTS LESSONS_COURSES_STREAM_FORMATED_V1 AS 
    SELECT
        LESSONS_COURSES_STREAM_ORIGIN_V1.AFTER->LESSON_ID AS LESSON_ID,
        LESSONS_COURSES_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
        LESSONS_COURSES_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LESSONS_COURSES_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM LESSONS_COURSES_STREAM_ORIGIN_V1
    WHERE LESSONS_COURSES_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_LESSONS_COURSES_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LESSONS_COURSES_STREAM_FORMATED_V1',
      'fields.whitelist'='lesson_id,course_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.lessons_courses',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,COURSE_ID:course_id,CREATED_AT:created_at,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='lesson_id,course_id'
);

CREATE STREAM IF NOT EXISTS CLASSROOM_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.classroom', value_format='AVRO');
CREATE STREAM IF NOT EXISTS CLASSROOM_STREAM_FORMATED_V1 AS 
    SELECT
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->CLASSROOM_ID AS rowkey,
        AS_VALUE(CLASSROOM_STREAM_ORIGIN_V1.AFTER->CLASSROOM_ID) AS CLASSROOM_ID,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->REMARKS AS REMARKS,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CLASSROOM_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM CLASSROOM_STREAM_ORIGIN_V1
    WHERE CLASSROOM_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_CLASSROOM_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}CLASSROOM_STREAM_FORMATED_V1',
      'fields.whitelist'='classroom_id,name,location_id,remarks,is_archived,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.classroom',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='CLASSROOM_ID:classroom_id,NAME:name,LOCATION_ID:location_id,REMARKS:remarks,IS_ARCHIVED:is_archived,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='classroom_id'
);
