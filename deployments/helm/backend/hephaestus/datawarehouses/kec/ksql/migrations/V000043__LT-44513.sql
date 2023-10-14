SET 'auto.offset.reset' = 'earliest';

/* grade */

CREATE STREAM IF NOT EXISTS GRADE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.mastermgmt.grade', value_format='AVRO');

CREATE STREAM IF NOT EXISTS GRADE_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}GRADE_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID AS KEY,
        AS_VALUE(GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID) AS GRADE_ID,
        GRADE_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        GRADE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        GRADE_STREAM_ORIGIN_V1.AFTER->PARTNER_INTERNAL_ID AS PARTNER_INTERNAL_ID,
        GRADE_STREAM_ORIGIN_V1.AFTER->SEQUENCE AS SEQUENCE,
        GRADE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        GRADE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        GRADE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM GRADE_STREAM_ORIGIN_V1
    WHERE GRADE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_GRADE_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_GRADE_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}GRADE_STREAM_FORMATED_V1',
      'fields.whitelist'='grade_id,name,is_archived,grade_created_at,grade_updated_at,grade_deleted_at,partner_internal_id,sequence',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.grade',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='GRADE_ID:grade_id,NAME:name,IS_ARCHIVED:is_archived,CREATED_AT:grade_created_at,UPDATED_AT:grade_updated_at,DELETED_AT:grade_deleted_at,PARTNER_INTERNAL_ID:partner_internal_id,SEQUENCE:sequence',
      'pk.fields'='grade_id'
);


/* academic_year */

CREATE STREAM IF NOT EXISTS ACADEMIC_YEAR_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.mastermgmt.academic_year', value_format='AVRO');

CREATE STREAM IF NOT EXISTS ACADEMIC_YEAR_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}ACADEMIC_YEAR_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->ACADEMIC_YEAR_ID AS KEY,
        AS_VALUE(ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->ACADEMIC_YEAR_ID) AS ACADEMIC_YEAR_ID,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM ACADEMIC_YEAR_STREAM_ORIGIN_V1
    WHERE ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY ACADEMIC_YEAR_STREAM_ORIGIN_V1.AFTER->ACADEMIC_YEAR_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_ACADEMIC_YEAR_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_ACADEMIC_YEAR_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ACADEMIC_YEAR_STREAM_FORMATED_V1',
      'fields.whitelist'='academic_year_id,name,start_date,end_date,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.academic_year',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ACADEMIC_YEAR_ID:academic_year_id,NAME:name,START_DATE:start_date,END_DATE:end_date,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='academic_year_id'
);


/* location type */
CREATE STREAM IF NOT EXISTS LOCATION_TYPES_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.location_types', value_format='AVRO');

CREATE STREAM IF NOT EXISTS LOCATION_TYPES_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}LOCATION_TYPES_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->LOCATION_TYPE_ID AS KEY,
        AS_VALUE(LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->LOCATION_TYPE_ID) AS LOCATION_TYPE_ID,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->DISPLAY_NAME AS DISPLAY_NAME,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->PARENT_NAME AS PARENT_NAME,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->PARENT_LOCATION_TYPE_ID AS PARENT_LOCATION_TYPE_ID,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM LOCATION_TYPES_STREAM_ORIGIN_V1
    WHERE LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY LOCATION_TYPES_STREAM_ORIGIN_V1.AFTER->LOCATION_TYPE_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_LOCATION_TYPES_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_LOCATION_TYPES_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LOCATION_TYPES_STREAM_FORMATED_V1',
      'fields.whitelist'='location_type_id,name,display_name,parent_name,parent_location_type_id,is_archived,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.location_types',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LOCATION_TYPE_ID:location_type_id,NAME:name,DISPLAY_NAME:display_name,PARENT_NAME:parent_name,PARENT_LOCATION_TYPE_ID:parent_location_type_id,IS_ARCHIVED:is_archived,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='location_type_id'
);


/* locations */
CREATE STREAM IF NOT EXISTS LOCATIONS_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.locations', value_format='AVRO');

CREATE STREAM IF NOT EXISTS LOCATIONS_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}LOCATIONS_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS KEY,
        AS_VALUE(LOCATIONS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID) AS LOCATION_ID,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->LOCATION_TYPE AS LOCATION_TYPE,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->PARENT_LOCATION_ID AS PARENT_LOCATION_ID,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->PARTNER_INTERNAL_ID AS PARTNER_INTERNAL_ID,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->PARTNER_INTERNAL_PARENT_ID AS PARTNER_INTERNAL_PARENT_ID,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->ACCESS_PATH AS ACCESS_PATH,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        LOCATIONS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM LOCATIONS_STREAM_ORIGIN_V1
    WHERE LOCATIONS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY LOCATIONS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_LOCATIONS_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_LOCATIONS_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LOCATIONS_STREAM_FORMATED_V1',
      'fields.whitelist'='location_id,name,location_type,parent_location_id,partner_internal_id,partner_internal_parent_id,is_archived,access_path,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.locations',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LOCATION_ID:location_id,NAME:name,PARENT_LOCATION_ID:parent_location_id,PARTNER_INTERNAL_ID:partner_internal_id,PARTNER_INTERNAL_PARENT_ID:partner_internal_parent_id,IS_ARCHIVED:is_archived,ACCESS_PATH:access_path,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='location_id'
);



/* subject */
CREATE STREAM IF NOT EXISTS SUBJECT_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.subject', value_format='AVRO');

CREATE STREAM IF NOT EXISTS SUBJECT_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}SUBJECT_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        SUBJECT_STREAM_ORIGIN_V1.AFTER->SUBJECT_ID AS KEY,
        AS_VALUE(SUBJECT_STREAM_ORIGIN_V1.AFTER->SUBJECT_ID) AS SUBJECT_ID,
        SUBJECT_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        SUBJECT_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        SUBJECT_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        SUBJECT_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM SUBJECT_STREAM_ORIGIN_V1
    WHERE SUBJECT_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY SUBJECT_STREAM_ORIGIN_V1.AFTER->SUBJECT_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_SUBJECT_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_SUBJECT_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SUBJECT_STREAM_FORMATED_V1',
      'fields.whitelist'='subject_id,name,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.subject',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SUBJECT_ID:subject_id,NAME:name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='subject_id'
);


/* class_class_members */
CREATE STREAM IF NOT EXISTS CLASS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.class', value_format='AVRO');
CREATE STREAM IF NOT EXISTS CLASS_MEMBERS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.class_member', value_format='AVRO');

CREATE STREAM IF NOT EXISTS CLASS_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}CLASS_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        CLASS_STREAM_ORIGIN_V1.AFTER->CLASS_ID AS KEY,
        AS_VALUE(CLASS_STREAM_ORIGIN_V1.AFTER->CLASS_ID) AS CLASS_ID,
        CLASS_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        CLASS_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
        CLASS_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID AS SCHOOL_ID,
        CLASS_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        CLASS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CLASS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CLASS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM CLASS_STREAM_ORIGIN_V1
    WHERE CLASS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY CLASS_STREAM_ORIGIN_V1.AFTER->CLASS_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS CLASS_TABLE (KEY VARCHAR(STRING) PRIMARY KEY)
WITH(kafka_topic='{{ .Values.topicPrefix }}CLASS_STREAM_FORMATED_V1', value_format='AVRO');


CREATE STREAM IF NOT EXISTS CLASS_MEMBERS_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}CLASS_MEMBERS_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->CLASS_MEMBER_ID AS KEY,
        AS_VALUE(CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->CLASS_MEMBER_ID) AS CLASS_MEMBER_ID,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->CLASS_ID AS CLASS_ID,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM CLASS_MEMBERS_STREAM_ORIGIN_V1
    WHERE CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY CLASS_MEMBERS_STREAM_ORIGIN_V1.AFTER->CLASS_MEMBER_ID
    EMIT CHANGES;

CREATE TABLE IF NOT EXISTS CLASS_MEMBERS_TABLE (KEY VARCHAR(STRING) PRIMARY KEY)
WITH(kafka_topic='{{ .Values.topicPrefix }}CLASS_MEMBERS_STREAM_FORMATED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS CLASS_CLASS_MEMBERS_JOIN
AS SELECT
    CLASS_TABLE.KEY AS KEY,
    CLASS_MEMBERS_TABLE.KEY AS KEY1,
    CLASS_MEMBERS_TABLE.CLASS_MEMBER_ID AS ROWKEY,
    CLASS_TABLE.CLASS_ID AS CLASS_ID,
    CLASS_TABLE.NAME AS NAME,
    CLASS_TABLE.COURSE_ID AS COURSE_ID,
    CLASS_TABLE.SCHOOL_ID AS SCHOOL_ID,
    CLASS_TABLE.LOCATION_ID AS LOCATION_ID,
    CLASS_TABLE.CREATED_AT AS CLASS_CREATED_AT,
    CLASS_TABLE.UPDATED_AT AS CLASS_UPDATED_AT,
    CLASS_TABLE.DELETED_AT AS CLASS_DELETED_AT,
    CLASS_MEMBERS_TABLE.USER_ID AS USER_ID,
    CLASS_MEMBERS_TABLE.START_DATE AS START_DATE,
    CLASS_MEMBERS_TABLE.END_DATE AS END_DATE,
    CLASS_MEMBERS_TABLE.CREATED_AT AS CLASS_MEMBER_CREATED_AT,
    CLASS_MEMBERS_TABLE.UPDATED_AT AS CLASS_MEMBER_UPDATED_AT,
    CLASS_MEMBERS_TABLE.DELETED_AT AS CLASS_MEMBER_DELETED_AT
FROM CLASS_MEMBERS_TABLE 
JOIN CLASS_TABLE
ON  CLASS_TABLE.KEY = CLASS_MEMBERS_TABLE.CLASS_ID;

DROP CONNECTOR IF EXISTS SINK_CLASS_CLASS_MEMBER_JOIN_V1;
CREATE SINK CONNECTOR SINK_CLASS_CLASS_MEMBER_JOIN_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}CLASS_CLASS_MEMBERS_JOIN',
      'fields.whitelist'='class_member_id,class_id,student_id,class_member_updated_at,class_member_created_at,class_member_deleted_at,start_date,end_date,name,course_id,location_id,school_id,class_updated_at,class_created_at,class_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'= '{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.class_class_members',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='CLASS_MEMBER_ID:class_member_id,CLASS_ID:class_id,USER_ID:student_id,CLASS_MEMBER_CREATED_AT:class_member_created_at,CLASS_MEMBER_UPDATED_AT:class_member_updated_at,CLASS_MEMBER_DELETED_AT:class_member_deleted_at,START_DATE:start_date,END_DATE:end_date,NAME:name,COURSE_ID:course_id,LOCATION_ID:location_id,CLASS_CREATED_AT:class_created_at,CLASS_UPDATED_AT:class_updated_at,CLASS_DELETED_AT:class_deleted_at',
      'pk.fields'='class_member_id'
);
