SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS STUDENT_COURSE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.student_course', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENT_COURSE_STREAM_FORMATTED_V1
    AS SELECT
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->LOCATION_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_PACKAGE_ID as KEY,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_PACKAGE_ID AS STUDENT_PACKAGE_ID,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->PACKAGE_TYPE AS PACKAGE_TYPE,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT AS COURSE_SLOT,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT_PER_WEEK AS COURSE_SLOT_PER_WEEK,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->WEIGHT AS WEIGHT,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_START_DATE AS STUDENT_START_DATE,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_END_DATE AS STUDENT_END_DATE,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM STUDENT_COURSE_STREAM_ORIGIN_V1
    WHERE STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->LOCATION_ID + STUDENT_COURSE_STREAM_ORIGIN_V1.AFTER->STUDENT_PACKAGE_ID
    EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS STUDENT_COURSE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_COURSE_STREAM_FORMATTED_V1',
      'fields.whitelist'='student_id,course_id,location_id,student_package_id,package_type,course_slot,course_slot_per_week,weight,student_start_date,student_end_date,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.student_course',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ID:student_id,COURSE_ID:course_id,LOCATION_ID:location_id,STUDENT_PACKAGE_ID:student_package_id,PACKAGE_TYPE:package_type,COURSE_SLOT:course_slot,COURSE_SLOT_PER_WEEK:course_slot_per_week,WEIGHT:weight,STUDENT_START_DATE:student_start_date,STUDENT_END_DATE:student_end_date,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='student_id,course_id,location_id,student_package_id'
);
