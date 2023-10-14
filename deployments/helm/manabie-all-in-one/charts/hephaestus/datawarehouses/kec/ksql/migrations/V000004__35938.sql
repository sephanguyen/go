set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS LESSONS_COURSES_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.lessons_courses', value_format='AVRO');
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
      'fields.whitelist'='lesson_id,course_id,created_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.lessons_courses_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,COURSE_ID:course_id,CREATED_AT:created_at,DELETED_AT:deleted_at',
      'pk.fields'='lesson_id,course_id'
);
