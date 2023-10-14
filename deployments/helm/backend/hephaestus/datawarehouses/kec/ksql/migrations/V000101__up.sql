SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.course_access_paths', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BOB_COURSE_ACCESS_PATHS_STREAM_FORMATTED_V1
    AS SELECT
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->COURSE_ID + BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID as KEY,
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS COURSE_ACCESS_PATHS_CREATED_AT,
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS COURSE_ACCESS_PATHS_UPDATED_AT,
        BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS COURSE_ACCESS_PATHS_DELETED_AT
    FROM BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1
    WHERE BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->COURSE_ID + BOB_COURSE_ACCESS_PATHS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BOB_COURSE_ACCESS_PATHS_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BOB_COURSE_TYPE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.course_type', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BOB_COURSE_TYPE_STREAM_FORMATTED_V1
    AS SELECT
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->COURSE_TYPE_ID as KEY,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->NAME AS COURSE_TYPE_NAME,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->REMARKS AS COURSE_TYPE_REMARKS,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS COURSE_TYPE_IS_ARCHIVED,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS COURSE_TYPE_CREATED_AT,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS COURSE_TYPE_UPDATED_AT,
        BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS COURSE_TYPE_DELETED_AT
    FROM BOB_COURSE_TYPE_STREAM_ORIGIN_V1
    WHERE BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY BOB_COURSE_TYPE_STREAM_ORIGIN_V1.AFTER->COURSE_TYPE_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BOB_COURSE_TYPE_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BOB_COURSE_TYPE_STREAM_FORMATTED_V1', value_format='AVRO');


CREATE STREAM IF NOT EXISTS BOB_COURSE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.courses', value_format='AVRO');


CREATE STREAM IF NOT EXISTS BOB_COURSE_STREAM_FORMATTED_V1
    AS SELECT
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID as KEY,
        AS_VALUE(BOB_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID) AS COURSE_ID,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->NAME AS COURSES_NAME,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->GRADE AS GRADE,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->SUBJECT AS SUBJECT,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->DISPLAY_ORDER AS DISPLAY_ORDER,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->PRESET_STUDY_PLAN_ID AS PRESET_STUDY_PLAN_ID,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->TEACHING_METHOD AS TEACHING_METHOD,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->ICON AS ICON,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->STATUS AS STATUS,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->REMARKS AS COURSE_REMARKS,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS COURSE_IS_ARCHIVED,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_PARTNER_ID AS COURSE_PARTNER_ID,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_TYPE_ID AS COURSE_TYPE_ID,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS COURSE_CREATED_AT,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS COURSE_UPDATED_AT,
        BOB_COURSE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS COURSE_DELETED_AT
    FROM BOB_COURSE_STREAM_ORIGIN_V1
    WHERE BOB_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY BOB_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS BOB_COURSE_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}BOB_COURSE_STREAM_FORMATTED_V1', value_format='AVRO');


CREATE TABLE IF NOT EXISTS BOB_COURSE_WITH_LOCATIONS_V1
AS SELECT
    BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.KEY AS ROW_KEY,
    AS_VALUE(BOB_COURSE_TABLE_FORMATTED_V1.KEY) AS COURSE_ID,
    BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.LOCATION_ID AS LOCATION_ID,
    BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.COURSE_ACCESS_PATHS_CREATED_AT AS COURSE_ACCESS_PATHS_CREATED_AT,
    BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.COURSE_ACCESS_PATHS_UPDATED_AT AS COURSE_ACCESS_PATHS_UPDATED_AT,
    BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.COURSE_ACCESS_PATHS_DELETED_AT AS COURSE_ACCESS_PATHS_DELETED_AT,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSES_NAME AS COURSES_NAME,
    BOB_COURSE_TABLE_FORMATTED_V1.GRADE AS GRADE,
    BOB_COURSE_TABLE_FORMATTED_V1.TEACHING_METHOD AS TEACHING_METHOD,
    BOB_COURSE_TABLE_FORMATTED_V1.SUBJECT AS SUBJECT,
    BOB_COURSE_TABLE_FORMATTED_V1.DISPLAY_ORDER AS DISPLAY_ORDER,
    BOB_COURSE_TABLE_FORMATTED_V1.START_DATE AS START_DATE,
    BOB_COURSE_TABLE_FORMATTED_V1.END_DATE AS END_DATE,
    BOB_COURSE_TABLE_FORMATTED_V1.PRESET_STUDY_PLAN_ID AS PRESET_STUDY_PLAN_ID,
    BOB_COURSE_TABLE_FORMATTED_V1.ICON AS ICON,
    BOB_COURSE_TABLE_FORMATTED_V1.STATUS AS STATUS,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_REMARKS AS COURSE_REMARKS,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_IS_ARCHIVED AS COURSE_IS_ARCHIVED,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_PARTNER_ID AS COURSE_PARTNER_ID,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_CREATED_AT AS COURSE_CREATED_AT,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_UPDATED_AT AS COURSE_UPDATED_AT,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_DELETED_AT AS COURSE_DELETED_AT,
    BOB_COURSE_TABLE_FORMATTED_V1.COURSE_TYPE_ID AS COURSE_TYPE_ID
FROM BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1
JOIN BOB_COURSE_TABLE_FORMATTED_V1
ON BOB_COURSE_ACCESS_PATHS_TABLE_FORMATTED_V1.COURSE_ID = BOB_COURSE_TABLE_FORMATTED_V1.KEY;



CREATE TABLE IF NOT EXISTS BOB_COURSE_WITH_LOCATIONS_AND_COURSE_TYPE_V1
AS SELECT
    BOB_COURSE_WITH_LOCATIONS_V1.ROW_KEY AS ROW_KEY,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.KEY AS KEY1,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_ID AS COURSE_ID,
    BOB_COURSE_WITH_LOCATIONS_V1.LOCATION_ID AS LOCATION_ID,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_ACCESS_PATHS_CREATED_AT AS COURSE_ACCESS_PATHS_CREATED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_ACCESS_PATHS_UPDATED_AT AS COURSE_ACCESS_PATHS_UPDATED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_ACCESS_PATHS_DELETED_AT AS COURSE_ACCESS_PATHS_DELETED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSES_NAME AS COURSES_NAME,
    BOB_COURSE_WITH_LOCATIONS_V1.GRADE AS GRADE,
    BOB_COURSE_WITH_LOCATIONS_V1.TEACHING_METHOD AS TEACHING_METHOD,
    BOB_COURSE_WITH_LOCATIONS_V1.SUBJECT AS SUBJECT,
    BOB_COURSE_WITH_LOCATIONS_V1.DISPLAY_ORDER AS DISPLAY_ORDER,
    BOB_COURSE_WITH_LOCATIONS_V1.START_DATE AS START_DATE,
    BOB_COURSE_WITH_LOCATIONS_V1.END_DATE AS END_DATE,
    BOB_COURSE_WITH_LOCATIONS_V1.PRESET_STUDY_PLAN_ID AS PRESET_STUDY_PLAN_ID,
    BOB_COURSE_WITH_LOCATIONS_V1.ICON AS ICON,
    BOB_COURSE_WITH_LOCATIONS_V1.STATUS AS STATUS,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_REMARKS AS COURSE_REMARKS,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_IS_ARCHIVED AS COURSE_IS_ARCHIVED,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_PARTNER_ID AS COURSE_PARTNER_ID,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_CREATED_AT AS COURSE_CREATED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_UPDATED_AT AS COURSE_UPDATED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_DELETED_AT AS COURSE_DELETED_AT,
    BOB_COURSE_WITH_LOCATIONS_V1.COURSE_TYPE_ID AS COURSE_TYPE_ID,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_NAME as COURSE_TYPE_NAME,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_REMARKS as COURSE_TYPE_REMARKS,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_IS_ARCHIVED as COURSE_TYPE_IS_ARCHIVED,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_CREATED_AT as COURSE_TYPE_CREATED_AT,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_UPDATED_AT as COURSE_TYPE_UPDATED_AT,
    BOB_COURSE_TYPE_TABLE_FORMATTED_V1.COURSE_TYPE_DELETED_AT as COURSE_TYPE_DELETED_AT
FROM BOB_COURSE_WITH_LOCATIONS_V1
JOIN BOB_COURSE_TYPE_TABLE_FORMATTED_V1
ON BOB_COURSE_WITH_LOCATIONS_V1.COURSE_TYPE_ID = BOB_COURSE_TYPE_TABLE_FORMATTED_V1.KEY;

DROP CONNECTOR IF EXISTS  SINK_COURSE_PUBLIC_INFO_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_BOB_COURSE_WITH_LOCATIONS_AND_COURSE_TYPE_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}BOB_COURSE_WITH_LOCATIONS_AND_COURSE_TYPE_V1',
      'fields.whitelist'='course_id,location_id,course_access_paths_created_at,course_access_paths_updated_at,course_access_paths_deleted_at,courses_name,grade,teaching_method,subject,display_order,start_date,end_date,preset_study_plan_id,icon,status,courses_remarks,course_is_archived,course_partner_id,course_created_at,course_updated_at,course_deleted_at,course_type_id,course_type_name,course_type_remarks,course_type_is_archived,course_type_created_at,course_type_updated_at,course_type_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.courses',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COURSE_ID:course_id,LOCATION_ID:location_id,COURSE_ACCESS_PATHS_CREATED_AT:course_access_paths_created_at,COURSE_ACCESS_PATHS_UPDATED_AT:course_access_paths_updated_at,COURSE_ACCESS_PATHS_DELETED_AT:course_access_paths_deleted_at,COURSES_NAME:courses_name,GRADE:grade,TEACHING_METHOD:teaching_method,SUBJECT:subject,DISPLAY_ORDER:display_order,START_DATE:start_date,END_DATE:end_date,PRESET_STUDY_PLAN_ID:preset_study_plan_id,ICON:icon,STATUS:status,COURSE_REMARKS:courses_remarks,COURSE_IS_ARCHIVED:course_is_archived,COURSE_PARTNER_ID:course_partner_id,COURSE_CREATED_AT:course_created_at,COURSE_UPDATED_AT:course_updated_at,COURSE_DELETED_AT:course_deleted_at,COURSE_TYPE_ID:course_type_id,COURSE_TYPE_NAME:course_type_name,COURSE_TYPE_REMARKS:course_type_remarks,COURSE_TYPE_IS_ARCHIVED:course_type_is_archived,COURSE_TYPE_CREATED_AT:course_type_created_at,COURSE_TYPE_UPDATED_AT:course_type_updated_at,COURSE_TYPE_DELETED_AT:course_type_deleted_at',
      'pk.fields'='course_id,location_id'
);


-- course_academic_year --

-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1
  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.mastermgmt.course_academic_year', value_format='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS COURSE_ACADEMIC_YEAR_STREAM_FORMATED_V1
  AS SELECT
              COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
              COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->ACADEMIC_YEAR_ID AS ACADEMIC_YEAR_ID,
              COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
              COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
       FROM COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1
       WHERE COURSE_ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_COURSE_ACADEMIC_YEAR_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}COURSE_ACADEMIC_YEAR_STREAM_FORMATED_V1',
      'fields.whitelist'='course_id,academic_year_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.course_academic_year',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COURSE_ID:course_id,ACADEMIC_YEAR_ID:academic_year_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='course_id,academic_year_id'
);


-- course_class --

CREATE SINK CONNECTOR IF NOT EXISTS SINK_COURSE_CLASS_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}CLASS_STREAM_FORMATED_V1',
      'fields.whitelist'='course_id,class_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.course_class',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COURSE_ID:course_id,CLASS_ID:class_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='course_id,class_id'
);
