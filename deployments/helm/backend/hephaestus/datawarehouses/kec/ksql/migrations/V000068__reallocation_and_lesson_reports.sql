set 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS SINK_LESSON_REALLOCATION_PUBLIC_INFO_V1;
DROP CONNECTOR IF EXISTS SINK_LESSON_REPORTS_PUBLIC_INFO;

DROP STREAM IF EXISTS LESSON_REALLOCATION_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSON_REALLOCATION_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS LESSON_REPORTS_PUBLIC_INFO DELETE TOPIC;
DROP STREAM IF EXISTS LESSON_REPORTS_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSON_REPORT_DETAILS_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSON_REPORTS_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1;

CREATE STREAM IF NOT EXISTS LESSON_REALLOCATION_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.reallocation', value_format='AVRO');
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
      'table.name.format'='public.reallocation',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
     'transforms.RenameField.renames'='STUDENT_ID:student_id,ORIGINAL_LESSON_ID:original_lesson_id,NEW_LESSON_ID:new_lesson_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,COURSE_ID:course_id',
      'pk.fields'='student_id,original_lesson_id'
);

CREATE STREAM IF NOT EXISTS LESSON_REPORTS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.lesson_reports', value_format='AVRO');
CREATE STREAM IF NOT EXISTS LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.lesson_report_details', value_format='AVRO');

CREATE STREAM IF NOT EXISTS LESSON_REPORTS_STREAM_FORMATED_V1 AS
    SELECT
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_ID AS KEY,
        AS_VALUE(LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_ID) AS LESSON_REPORT_ID,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->REPORT_SUBMITTING_STATUS AS REPORT_SUBMITTING_STATUS,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->FORM_CONFIG_ID AS FORM_CONFIG_ID,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->LESSON_ID AS LESSON_ID
    FROM LESSON_REPORTS_STREAM_ORIGIN_V1
    WHERE LESSON_REPORTS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->LESSON_REPORT_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS LESSON_REPORTS_TABLE_FORMATED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}LESSON_REPORTS_STREAM_FORMATED_V1', value_format='AVRO');

CREATE STREAM IF NOT EXISTS LESSON_REPORT_DETAILS_STREAM_FORMATED_V1 AS
    SELECT
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_DETAIL_ID AS KEY,
        AS_VALUE(LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_DETAIL_ID) AS LESSON_REPORT_DETAIL_ID,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_ID AS LESSON_REPORT_ID,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->REPORT_VERSIONS AS REPORT_VERSIONS,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1
    WHERE LESSON_REPORT_DETAILS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->LESSON_REPORT_DETAIL_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS LESSON_REPORT_DETAILS_TABLE_FORMATED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}LESSON_REPORT_DETAILS_STREAM_FORMATED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS LESSON_REPORTS_PUBLIC_INFO AS
    SELECT
        LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.KEY AS LESSON_REPORT_DETAIL_ID,
        LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.STUDENT_ID AS STUDENT_ID,
        LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.CREATED_AT AS LESSON_REPORT_DETAILS_CREATED_AT,
        LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.UPDATED_AT AS LESSON_REPORT_DETAILS_UPDATED_AT,
        LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.DELETED_AT AS LESSON_REPORT_DETAILS_DELETED_AT,

        AS_VALUE(LESSON_REPORTS_TABLE_FORMATED_V1.LESSON_REPORT_ID) AS LESSON_REPORT_ID,
        LESSON_REPORTS_TABLE_FORMATED_V1.CREATED_AT AS LESSON_REPORTS_CREATED_AT,
        LESSON_REPORTS_TABLE_FORMATED_V1.UPDATED_AT AS LESSON_REPORTS_UPDATED_AT,
        LESSON_REPORTS_TABLE_FORMATED_V1.DELETED_AT AS LESSON_REPORTS_DELETED_AT,
        LESSON_REPORTS_TABLE_FORMATED_V1.REPORT_SUBMITTING_STATUS AS REPORT_SUBMITTING_STATUS,
        LESSON_REPORTS_TABLE_FORMATED_V1.FORM_CONFIG_ID AS FORM_CONFIG_ID,
        LESSON_REPORTS_TABLE_FORMATED_V1.LESSON_ID AS LESSON_ID
FROM LESSON_REPORT_DETAILS_TABLE_FORMATED_V1 JOIN LESSON_REPORTS_TABLE_FORMATED_V1 
ON LESSON_REPORT_DETAILS_TABLE_FORMATED_V1.LESSON_REPORT_ID = LESSON_REPORTS_TABLE_FORMATED_V1.KEY;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_LESSON_REPORTS_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LESSON_REPORTS_PUBLIC_INFO',
      'fields.whitelist'='lesson_report_detail_id,lesson_report_id,student_id,lesson_report_details_created_at,lesson_report_details_updated_at,lesson_report_details_deleted_at,lesson_reports_created_at,lesson_reports_updated_at,lesson_reports_deleted_at,report_submitting_status,form_config_id,lesson_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.lesson_reports',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_REPORT_DETAIL_ID:lesson_report_detail_id,LESSON_REPORT_ID:lesson_report_id,STUDENT_ID:student_id,LESSON_REPORT_DETAILS_CREATED_AT:lesson_report_details_created_at,LESSON_REPORT_DETAILS_UPDATED_AT:lesson_report_details_updated_at,LESSON_REPORT_DETAILS_DELETED_AT:lesson_report_details_deleted_at,LESSON_REPORTS_CREATED_AT:lesson_reports_created_at,LESSON_REPORTS_UPDATED_AT:lesson_reports_updated_at,LESSON_REPORTS_DELETED_AT:lesson_reports_deleted_at,REPORT_SUBMITTING_STATUS:report_submitting_status,FORM_CONFIG_ID:form_config_id,LESSON_ID:lesson_id',
      'pk.fields'='lesson_report_detail_id'
);
