set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS LESSON_REALLOCATION_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.reallocation', value_format='AVRO');
CREATE STREAM IF NOT EXISTS LESSON_REALLOCATION_STREAM_FORMATED_V1
AS SELECT
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->ORIGINAL_LESSON_ID AS rowkey,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->NEW_LESSON_ID AS NEW_LESSON_ID,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
   LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
   AS_VALUE(LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->ORIGINAL_LESSON_ID) AS ORIGINAL_LESSON_ID
   FROM LESSON_REALLOCATION_STREAM_ORIGIN_V1
   WHERE LESSON_REALLOCATION_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_LESSON_REALLOCATION_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LESSON_REALLOCATION_STREAM_FORMATED_V1',
      'fields.whitelist'='student_id,original_lesson_id,new_lesson_id,created_at,updated_at,deleted_at,course_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.reallocation_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
     'transforms.RenameField.renames'='STUDENT_ID:student_id,ORIGINAL_LESSON_ID:original_lesson_id,NEW_LESSON_ID:new_lesson_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,COURSE_ID:course_id',
      'pk.fields'='student_id,original_lesson_id'
);
