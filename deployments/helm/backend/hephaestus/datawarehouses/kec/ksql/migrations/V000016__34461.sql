SET 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS SCHOOL_HISTORY_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.school_history', value_format='AVRO');
CREATE STREAM IF NOT EXISTS SCHOOL_HISTORY_STREAM_FORMATED_V1
    AS SELECT
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID AS SCHOOL_ID,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->SCHOOL_COURSE_ID AS SCHOOL_COURSE_ID,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->IS_CURRENT AS IS_CURRENT,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM SCHOOL_HISTORY_STREAM_ORIGIN_V1
    WHERE SCHOOL_HISTORY_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS SCHOOL_HISTORY_PUBLIC_INFO_V1 
AS SELECT
    SCHOOL_HISTORY_STREAM_FORMATED_V1.STUDENT_ID AS STUDENT_ID,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.SCHOOL_ID AS SCHOOL_ID,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.SCHOOL_COURSE_ID AS SCHOOL_COURSE_ID,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.START_DATE AS START_DATE,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.END_DATE AS END_DATE,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.IS_CURRENT AS IS_CURRENT,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.CREATED_AT AS CREATED_AT,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.UPDATED_AT AS UPDATED_AT,
    SCHOOL_HISTORY_STREAM_FORMATED_V1.DELETED_AT AS DELETED_AT

FROM SCHOOL_HISTORY_STREAM_FORMATED_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SCHOOL_HISTORY_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHOOL_HISTORY_PUBLIC_INFO_V1',
      'fields.whitelist'='student_id,school_id,school_course_id,start_date,end_date,is_current,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.school_history_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ID:student_id,SCHOOL_ID:school_id,SCHOOL_COURSE_ID:school_course_id,START_DATE:start_date,END_DATE:end_date,IS_CURRENT:is_current,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='student_id,school_id'
);

CREATE STREAM IF NOT EXISTS USER_ADDRESS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.user_address', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USER_ADDRESS_STREAM_FORMATED_V1
    AS SELECT
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->USER_ADDRESS_ID AS USER_ADDRESS_ID,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->ADDRESS_TYPE AS ADDRESS_TYPE,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->POSTAL_CODE AS POSTAL_CODE,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->PREFECTURE_ID AS PREFECTURE_ID,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->CITY AS CITY,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->FIRST_STREET AS FIRST_STREET,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->SECOND_STREET AS SECOND_STREET,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM USER_ADDRESS_STREAM_ORIGIN_V1
    WHERE USER_ADDRESS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS USER_ADDRESS_PUBLIC_INFO_V1 
AS SELECT
    USER_ADDRESS_STREAM_FORMATED_V1.USER_ADDRESS_ID AS STUDENT_ADDRESS_ID,
    USER_ADDRESS_STREAM_FORMATED_V1.USER_ID AS STUDENT_ID,
    USER_ADDRESS_STREAM_FORMATED_V1.ADDRESS_TYPE AS ADDRESS_TYPE,
    USER_ADDRESS_STREAM_FORMATED_V1.POSTAL_CODE AS POSTAL_CODE,
    USER_ADDRESS_STREAM_FORMATED_V1.PREFECTURE_ID AS PREFECTURE_ID,
    USER_ADDRESS_STREAM_FORMATED_V1.CITY AS CITY,
    USER_ADDRESS_STREAM_FORMATED_V1.FIRST_STREET AS FIRST_STREET,
    USER_ADDRESS_STREAM_FORMATED_V1.SECOND_STREET AS SECOND_STREET,
    USER_ADDRESS_STREAM_FORMATED_V1.CREATED_AT AS USER_ADDRESS_CREATED_AT,
    USER_ADDRESS_STREAM_FORMATED_V1.UPDATED_AT AS USER_ADDRESS_UPDATED_AT,
    USER_ADDRESS_STREAM_FORMATED_V1.DELETED_AT AS USER_ADDRESS_DELETED_AT

FROM USER_ADDRESS_STREAM_FORMATED_V1
PARTITION BY USER_ADDRESS_STREAM_FORMATED_V1.USER_ADDRESS_ID;

CREATE SINK CONNECTOR IF NOT EXISTS USER_ADDRESS_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_ADDRESS_PUBLIC_INFO_V1',
      'fields.whitelist'='student_address_id,student_id,address_type,postal_code,prefecture_id,city,user_address_created_at,user_address_updated_at,user_address_deleted_at,first_street,second_street',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_address_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ADDRESS_ID:student_address_id,STUDENT_ID:student_id,ADDRESS_TYPE:address_type,POSTAL_CODE:postal_code,PREFECTURE_ID:prefecture_id,CITY:city,USER_ADDRESS_CREATED_AT:user_address_created_at,USER_ADDRESS_UPDATED_AT:user_address_updated_at,USER_ADDRESS_DELETED_AT:user_address_deleted_at,FIRST_STREET:first_street,SECOND_STREET:second_street',
      'pk.fields'='student_address_id'
);