set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS LESSONS_TEACHERS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.lessons_teachers', value_format='AVRO');
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
      'fields.whitelist'='lesson_id,staff_id,created_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.lessons_teachers_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,STAFF_ID:staff_id,CREATED_AT:created_at,DELETED_AT:deleted_at',
      'pk.fields'='lesson_id,staff_id'
);
