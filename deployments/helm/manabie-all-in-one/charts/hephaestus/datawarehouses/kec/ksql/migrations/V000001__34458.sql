SET 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS USERS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.users', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USERS_STREAM_FORMATED_V1
    AS SELECT
         USERS_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
         USERS_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
         USERS_STREAM_ORIGIN_V1.AFTER->COUNTRY AS COUNTRY,
         USERS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
         USERS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
         USERS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
         USERS_STREAM_ORIGIN_V1.AFTER->AVATAR AS AVATAR,
         USERS_STREAM_ORIGIN_V1.AFTER->PHONE_NUMBER AS PHONE_NUMBER,
         USERS_STREAM_ORIGIN_V1.AFTER->EMAIL AS EMAIL,
         USERS_STREAM_ORIGIN_V1.AFTER->DEVICE_TOKEN AS DEVICE_TOKEN,
         USERS_STREAM_ORIGIN_V1.AFTER->ALLOW_NOTIFICATION AS ALLOW_NOTIFICATION,
         USERS_STREAM_ORIGIN_V1.AFTER->USER_GROUP AS USER_GROUP,
         USERS_STREAM_ORIGIN_V1.AFTER->GIVEN_NAME AS GIVEN_NAME,
         USERS_STREAM_ORIGIN_V1.AFTER->LAST_LOGIN_DATE AS LAST_LOGIN_DATE,
         USERS_STREAM_ORIGIN_V1.AFTER->BIRTHDAY AS BIRTHDAY,
         USERS_STREAM_ORIGIN_V1.AFTER->GENDER AS GENDER,
         USERS_STREAM_ORIGIN_V1.AFTER->FIRST_NAME AS FIRST_NAME,
         USERS_STREAM_ORIGIN_V1.AFTER->LAST_NAME AS LAST_NAME,
         USERS_STREAM_ORIGIN_V1.AFTER->FIRST_NAME_PHONETIC AS FIRST_NAME_PHONETIC,
         USERS_STREAM_ORIGIN_V1.AFTER->LAST_NAME_PHONETIC AS LAST_NAME_PHONETIC,
         USERS_STREAM_ORIGIN_V1.AFTER->FULL_NAME_PHONETIC AS FULL_NAME_PHONETIC,
         USERS_STREAM_ORIGIN_V1.AFTER->REMARKS AS REMARKS,
         USERS_STREAM_ORIGIN_V1.AFTER->IS_SYSTEM AS IS_SYSTEM,
         USERS_STREAM_ORIGIN_V1.AFTER->USER_EXTERNAL_ID AS USER_EXTERNAL_ID
    FROM USERS_STREAM_ORIGIN_V1
    WHERE USERS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS STAFF_STREAM_ORIGIN_V1   WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.staff', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STAFF_STREAM_FORMATED_V1
    AS SELECT
         STAFF_STREAM_ORIGIN_V1.AFTER->STAFF_ID AS STAFF_ID,
         STAFF_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
         STAFF_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
         STAFF_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
         STAFF_STREAM_ORIGIN_V1.AFTER->WORKING_STATUS AS WORKING_STATUS,
         STAFF_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
         STAFF_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE
    FROM STAFF_STREAM_ORIGIN_V1
    WHERE STAFF_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS STAFF_PUBLIC_INFO_V1 
AS SELECT
    STAFF_STREAM_FORMATED_V1.STAFF_ID AS STAFF_ID,
    STAFF_STREAM_FORMATED_V1.CREATED_AT AS STAFF_CREATED_AT,
    STAFF_STREAM_FORMATED_V1.UPDATED_AT AS STAFF_UPDATED_AT,
    STAFF_STREAM_FORMATED_V1.DELETED_AT AS STAFF_DELETED_AT,
    STAFF_STREAM_FORMATED_V1.WORKING_STATUS AS WORKING_STATUS,
    STAFF_STREAM_FORMATED_V1.START_DATE AS START_DATE,
    STAFF_STREAM_FORMATED_V1.END_DATE AS END_DATE,

    USERS_STREAM_FORMATED_V1.COUNTRY AS COUNTRY,
    USERS_STREAM_FORMATED_V1.NAME AS NAME,
    USERS_STREAM_FORMATED_V1.AVATAR AS AVATAR,
    USERS_STREAM_FORMATED_V1.PHONE_NUMBER AS PHONE_NUMBER,
    USERS_STREAM_FORMATED_V1.EMAIL AS LEARNER_IDS,
    USERS_STREAM_FORMATED_V1.DEVICE_TOKEN AS DEVICE_TOKEN,
    USERS_STREAM_FORMATED_V1.ALLOW_NOTIFICATION AS ALLOW_NOTIFICATION,
    USERS_STREAM_FORMATED_V1.USER_GROUP AS USER_GROUP,
    USERS_STREAM_FORMATED_V1.CREATED_AT AS USERS_CREATED_AT,
    USERS_STREAM_FORMATED_V1.UPDATED_AT AS USERS_UPDATED_AT,
    USERS_STREAM_FORMATED_V1.DELETED_AT AS USERS_DELETED_AT,
    USERS_STREAM_FORMATED_V1.GIVEN_NAME AS GIVEN_NAME,
    USERS_STREAM_FORMATED_V1.LAST_LOGIN_DATE AS LAST_LOGIN_DATE,
    USERS_STREAM_FORMATED_V1.BIRTHDAY AS BIRTHDAY,
    USERS_STREAM_FORMATED_V1.GENDER AS GENDER,
    USERS_STREAM_FORMATED_V1.FIRST_NAME AS FIRST_NAME,
    USERS_STREAM_FORMATED_V1.LAST_NAME AS LAST_NAME,
    USERS_STREAM_FORMATED_V1.FIRST_NAME_PHONETIC AS FIRST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.LAST_NAME_PHONETIC AS LAST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.FULL_NAME_PHONETIC AS FULL_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.REMARKS AS REMARKS,
    USERS_STREAM_FORMATED_V1.IS_SYSTEM AS IS_SYSTEM,
    USERS_STREAM_FORMATED_V1.USER_EXTERNAL_ID AS USER_EXTERNAL_ID
FROM USERS_STREAM_FORMATED_V1 JOIN STAFF_STREAM_FORMATED_V1 WITHIN 2 HOURS ON USERS_STREAM_FORMATED_V1.USER_ID = STAFF_STREAM_FORMATED_V1.STAFF_ID;

CREATE SINK CONNECTOR IF NOT EXISTS STAFF_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STAFF_PUBLIC_INFO_V1',
      'fields.whitelist'='country,name,avatar,phone_number,email,device_token,allow_notification,user_group,users_created_at,users_updated_at,users_deleted_at,given_name,last_login_date,birthday,gender,first_name,last_name,first_name_phonetic,last_name_phonetic,full_name_phonetic,remarks,is_system,user_external_id,staff_id,staff_created_at,staff_updated_at,staff_deleted_at,working_status,start_date,end_date',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.staff_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,STAFF_ID:staff_id,STAFF_CREATED_AT:staff_created_at,STAFF_UPDATED_AT:staff_updated_at,STAFF_DELETED_AT:staff_deleted_at,WORKING_STATUS:working_status,START_DATE:start_date,END_DATE:end_date',
      'pk.fields'='staff_id'
);

CREATE STREAM IF NOT EXISTS STUDENT_PARENTS_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.student_parents', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENT_PARENTS_STREAM_FORMATED_V1
AS SELECT
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->PARENT_ID AS PARENT_ID,
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT, 
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
              STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->RELATIONSHIP AS RELATIONSHIP
       FROM STUDENT_PARENTS_STREAM_ORIGIN_V1
       WHERE STUDENT_PARENTS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS STUDENT_PARENTS_PUBLIC_INFO_V1 
AS SELECT
    STUDENT_PARENTS_STREAM_FORMATED_V1.STUDENT_ID AS STUDENT_ID,
    STUDENT_PARENTS_STREAM_FORMATED_V1.PARENT_ID AS PARENT_ID,
    STUDENT_PARENTS_STREAM_FORMATED_V1.CREATED_AT AS CREATED_AT,
    STUDENT_PARENTS_STREAM_FORMATED_V1.UPDATED_AT AS UPDATED_AT,
    STUDENT_PARENTS_STREAM_FORMATED_V1.DELETED_AT AS DELETED_AT,
    STUDENT_PARENTS_STREAM_FORMATED_V1.RELATIONSHIP AS RELATIONSHIP
FROM STUDENT_PARENTS_STREAM_FORMATED_V1;

CREATE SINK CONNECTOR IF NOT EXISTS STUDENT_PARENTS_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_PARENTS_PUBLIC_INFO_V1',
      'fields.whitelist'='student_id,parent_id,created_at,updated_at,deleted_at,relationship',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.student_parents_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ID:student_id,PARENT_ID:parent_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,RELATIONSHIP:relationship',
      'pk.fields'='student_id,parent_id'
);

CREATE STREAM IF NOT EXISTS STUDENTS_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.students', value_format='AVRO');
CREATE STREAM IF NOT EXISTS STUDENTS_STREAM_FORMATED_V1
AS SELECT
              STUDENTS_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT, 
              STUDENTS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->BIRTHDAY AS BIRTHDAY,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->SCHOOL_ID AS SCHOOL_ID,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->STUDENT_NOTE AS STUDENT_NOTE,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->CONTACT_PREFERENCE AS CONTACT_PREFERENCE,
              STUDENTS_STREAM_ORIGIN_V1.AFTER->GRADE_ID AS GRADE_ID
       FROM STUDENTS_STREAM_ORIGIN_V1
       WHERE STUDENTS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';


CREATE STREAM IF NOT EXISTS STUDENTS_PUBLIC_INFO_V1 
AS SELECT
    STUDENTS_STREAM_FORMATED_V1.STUDENT_ID AS STUDENT_ID,
    STUDENTS_STREAM_FORMATED_V1.CREATED_AT AS STUDENTS_CREATED_AT,
    STUDENTS_STREAM_FORMATED_V1.UPDATED_AT AS STUDENTS_UPDATED_AT,
    STUDENTS_STREAM_FORMATED_V1.DELETED_AT AS STUDENTS_DELETED_AT,
    STUDENTS_STREAM_FORMATED_V1.BIRTHDAY AS STUDENTS_BIRTHDAY,
    STUDENTS_STREAM_FORMATED_V1.SCHOOL_ID AS SCHOOL_ID,
    STUDENTS_STREAM_FORMATED_V1.STUDENT_NOTE AS STUDENT_NOTE,
    STUDENTS_STREAM_FORMATED_V1.CONTACT_PREFERENCE AS CONTACT_PREFERENCE,
    STUDENTS_STREAM_FORMATED_V1.GRADE_ID AS GRADE_ID,

    USERS_STREAM_FORMATED_V1.COUNTRY AS COUNTRY,
    USERS_STREAM_FORMATED_V1.NAME AS NAME,
    USERS_STREAM_FORMATED_V1.AVATAR AS AVATAR,
    USERS_STREAM_FORMATED_V1.PHONE_NUMBER AS PHONE_NUMBER,
    USERS_STREAM_FORMATED_V1.EMAIL AS LEARNER_IDS,
    USERS_STREAM_FORMATED_V1.DEVICE_TOKEN AS DEVICE_TOKEN,
    USERS_STREAM_FORMATED_V1.ALLOW_NOTIFICATION AS ALLOW_NOTIFICATION,
    USERS_STREAM_FORMATED_V1.USER_GROUP AS USER_GROUP,
    USERS_STREAM_FORMATED_V1.CREATED_AT AS USERS_CREATED_AT,
    USERS_STREAM_FORMATED_V1.UPDATED_AT AS USERS_UPDATED_AT,
    USERS_STREAM_FORMATED_V1.DELETED_AT AS USERS_DELETED_AT,
    USERS_STREAM_FORMATED_V1.GIVEN_NAME AS GIVEN_NAME,
    USERS_STREAM_FORMATED_V1.LAST_LOGIN_DATE AS LAST_LOGIN_DATE,
    USERS_STREAM_FORMATED_V1.BIRTHDAY AS BIRTHDAY,
    USERS_STREAM_FORMATED_V1.GENDER AS GENDER,
    USERS_STREAM_FORMATED_V1.FIRST_NAME AS FIRST_NAME,
    USERS_STREAM_FORMATED_V1.LAST_NAME AS LAST_NAME,
    USERS_STREAM_FORMATED_V1.FIRST_NAME_PHONETIC AS FIRST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.LAST_NAME_PHONETIC AS LAST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.FULL_NAME_PHONETIC AS FULL_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.REMARKS AS REMARKS,
    USERS_STREAM_FORMATED_V1.IS_SYSTEM AS IS_SYSTEM,
    USERS_STREAM_FORMATED_V1.USER_EXTERNAL_ID AS USER_EXTERNAL_ID
FROM USERS_STREAM_FORMATED_V1 JOIN STUDENTS_STREAM_FORMATED_V1 WITHIN 2 HOURS ON USERS_STREAM_FORMATED_V1.USER_ID = STUDENTS_STREAM_FORMATED_V1.STUDENT_ID;

CREATE SINK CONNECTOR IF NOT EXISTS STUDENTS_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENTS_PUBLIC_INFO_V1',
      'fields.whitelist'='country,name,avatar,phone_number,email,device_token,allow_notification,user_group,users_created_at,users_updated_at,users_deleted_at,given_name,last_login_date,birthday,gender,first_name,last_name,first_name_phonetic,last_name_phonetic,full_name_phonetic,remarks,is_system,user_external_id,student_id,students_created_at,students_updated_at,students_deleted_at,students_birthday,school_id,contact_preference,student_note,grade_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.students_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,STUDENT_ID:student_id,STUDENTS_CREATED_AT:students_created_at,STUDENTS_UPDATED_AT:students_updated_at,STUDENTS_DELETED_AT:students_deleted_at,STUDENTS_BIRTHDAY:students_birthday,SCHOOL_ID:school_id,STUDENT_NOTE:student_note,CONTACT_PREFERENCE:contact_preference,GRADE_ID:grade_id',
      'pk.fields'='student_id'
);

CREATE STREAM IF NOT EXISTS PARENTS_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.parents', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PARENTS_STREAM_FORMATED_V1
AS SELECT
              PARENTS_STREAM_ORIGIN_V1.AFTER->PARENT_ID AS PARENT_ID,
              PARENTS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT, 
              PARENTS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              PARENTS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
       FROM PARENTS_STREAM_ORIGIN_V1
       WHERE PARENTS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS PARENTS_PUBLIC_INFO_V1 
AS SELECT
    PARENTS_STREAM_FORMATED_V1.PARENT_ID AS PARENT_ID,
    PARENTS_STREAM_FORMATED_V1.CREATED_AT AS PARENTS_CREATED_AT,
    PARENTS_STREAM_FORMATED_V1.UPDATED_AT AS PARENTS_UPDATED_AT,
    PARENTS_STREAM_FORMATED_V1.DELETED_AT AS PARENTS_DELETED_AT,

    USERS_STREAM_FORMATED_V1.COUNTRY AS COUNTRY,
    USERS_STREAM_FORMATED_V1.NAME AS NAME,
    USERS_STREAM_FORMATED_V1.AVATAR AS AVATAR,
    USERS_STREAM_FORMATED_V1.PHONE_NUMBER AS PHONE_NUMBER,
    USERS_STREAM_FORMATED_V1.EMAIL AS LEARNER_IDS,
    USERS_STREAM_FORMATED_V1.DEVICE_TOKEN AS DEVICE_TOKEN,
    USERS_STREAM_FORMATED_V1.ALLOW_NOTIFICATION AS ALLOW_NOTIFICATION,
    USERS_STREAM_FORMATED_V1.USER_GROUP AS USER_GROUP,
    USERS_STREAM_FORMATED_V1.CREATED_AT AS USERS_CREATED_AT,
    USERS_STREAM_FORMATED_V1.UPDATED_AT AS USERS_UPDATED_AT,
    USERS_STREAM_FORMATED_V1.DELETED_AT AS USERS_DELETED_AT,
    USERS_STREAM_FORMATED_V1.GIVEN_NAME AS GIVEN_NAME,
    USERS_STREAM_FORMATED_V1.LAST_LOGIN_DATE AS LAST_LOGIN_DATE,
    USERS_STREAM_FORMATED_V1.BIRTHDAY AS BIRTHDAY,
    USERS_STREAM_FORMATED_V1.GENDER AS GENDER,
    USERS_STREAM_FORMATED_V1.FIRST_NAME AS FIRST_NAME,
    USERS_STREAM_FORMATED_V1.LAST_NAME AS LAST_NAME,
    USERS_STREAM_FORMATED_V1.FIRST_NAME_PHONETIC AS FIRST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.LAST_NAME_PHONETIC AS LAST_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.FULL_NAME_PHONETIC AS FULL_NAME_PHONETIC,
    USERS_STREAM_FORMATED_V1.REMARKS AS REMARKS,
    USERS_STREAM_FORMATED_V1.IS_SYSTEM AS IS_SYSTEM,
    USERS_STREAM_FORMATED_V1.USER_EXTERNAL_ID AS USER_EXTERNAL_ID
FROM USERS_STREAM_FORMATED_V1 JOIN PARENTS_STREAM_FORMATED_V1 WITHIN 2 HOURS  ON USERS_STREAM_FORMATED_V1.USER_ID = PARENTS_STREAM_FORMATED_V1.PARENT_ID;

CREATE SINK CONNECTOR IF NOT EXISTS PARENTS_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PARENTS_PUBLIC_INFO_V1',
      'fields.whitelist'='country,name,avatar,phone_number,email,device_token,allow_notification,user_group,users_created_at,users_updated_at,users_deleted_at,given_name,last_login_date,birthday,gender,first_name,last_name,first_name_phonetic,last_name_phonetic,full_name_phonetic,remarks,is_system,user_external_id,parent_id,parents_created_at,parents_updated_at,parents_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.parents_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,PARENT_ID:parent_id,PARENTS_CREATED_AT:parents_created_at,PARENTS_UPDATED_AT:parents_updated_at,PARENTS_DELETED_AT:parents_deleted_at',
      'pk.fields'='parent_id'
);