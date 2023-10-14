DROP CONNECTOR IF EXISTS STAFF_STREAM_FORMATED_V1;
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
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,STAFF_ID:staff_id,STAFF_CREATED_AT:staff_created_at,STAFF_UPDATED_AT:staff_updated_at,STAFF_DELETED_AT:staff_deleted_at,WORKING_STATUS:working_status,START_DATE:start_date,END_DATE:end_date',
      'pk.fields'='staff_id'
);

DROP CONNECTOR IF EXISTS STUDENTS_STREAM_FORMATED_V1;
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
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,STUDENT_ID:student_id,STUDENTS_CREATED_AT:students_created_at,STUDENTS_UPDATED_AT:students_updated_at,STUDENTS_DELETED_AT:students_deleted_at,STUDENTS_BIRTHDAY:students_birthday,SCHOOL_ID:school_id,STUDENT_NOTE:student_note,CONTACT_PREFERENCE:contact_preference,GRADE_ID:grade_id',
      'pk.fields'='student_id'
);

DROP CONNECTOR IF EXISTS PARENTS_STREAM_FORMATED_V1;
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
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='COUNTRY:country,NAME:name,AVATAR:avatar,PHONE_NUMBER:phone_number,EMAIL:email,DEVICE_TOKEN:device_token,ALLOW_NOTIFICATION:allow_notification,USER_GROUP:user_group,USERS_CREATED_AT:users_created_at,USERS_UPDATED_AT:users_updated_at,USERS_DELETED_AT:users_deleted_at,GIVEN_NAME:given_name,LAST_LOGIN_DATE:last_login_date,BIRTHDAY:birthdat,GENDER:gender,FIRST_NAME:first_name,LAST_NAME:last_name,FIRST_NAME_PHONETIC:first_name_phonetic,LAST_NAME_PHONETIC:last_name_phonetic,FULL_NAME_PHONETIC:full_name_phonetic,REMARKS:remarks,IS_SYSTEM:is_system,USER_EXTERNAL_ID:user_external_id,PARENT_ID:parent_id,PARENTS_CREATED_AT:parents_created_at,PARENTS_UPDATED_AT:parents_updated_at,PARENTS_DELETED_AT:parents_deleted_at',
      'pk.fields'='parent_id'
);

DROP CONNECTOR IF EXISTS ROLE_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS ROLE_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ROLE_PUBLIC_INFO_V1',
      'fields.whitelist'='role_id,role_name,role_created_at,role_updated_at,role_deleted_at,granted_role_id,user_group_id,granted_role_created_at,granted_role_updated_at,granted_role_deleted_at,location_id,granted_role_access_path_created_at,granted_role_access_path_updated_at,granted_role_access_path_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.role_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ROLE_ID:role_id,ROLE_NAME:role_name,ROLE_CREATED_AT:role_created_at,ROLE_UPDATED_AT:role_updated_at,ROLE_DELETED_AT:role_deleted_at,GRANTED_ROLE_ID:granted_role_id,USER_GROUP_ID:user_group_id,GRANTED_ROLE_CREATED_AT:granted_role_created_at,GRANTED_ROLE_UPDATED_AT:granted_role_updated_at,GRANTED_ROLE_DELETED_AT:granted_role_deleted_at,LOCATION_ID:location_id,GRANTED_ROLE_ACCESS_PATH_CREATED_AT:granted_role_access_path_created_at,GRANTED_ROLE_ACCESS_PATH_UPDATED_AT:granted_role_access_path_updated_at,GRANTED_ROLE_ACCESS_PATH_DELETED_AT:granted_role_access_path_deleted_at',
      'pk.fields'='role_id'
);

DROP CONNECTOR IF EXISTS SCHOOL_LEVEL_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SCHOOL_LEVEL_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHOOL_LEVEL_PUBLIC_INFO_V1',
      'fields.whitelist'='level_id,level_name,school_level_sequence,school_level_created_at,school_level_updated_at,school_level_deleted_at,school_level_is_archived,grade_id,name,partner_internal_id,grade_sequence,grade_created_at,grade_updated_at,grade_deleted_at,grade_is_archived,school_level_grade_created_at,school_level_grade_updated_at,school_level_grade_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.school_level_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LEVEL_ID:level_id,LEVEL_NAME:level_name,SCHOOL_LEVEL_SEQUENCE:school_level_sequence,SCHOOL_LEVEL_CREATED_AT:school_level_created_at,SCHOOL_LEVEL_UPDATED_AT:school_level_updated_at,SCHOOL_LEVEL_DELETED_AT:school_level_deleted_at,SCHOOL_LEVEL_IS_ARCHIVED:school_level_is_archived,GRADE_ID:grade_id,NAME:name,PARTNER_INTERNAL_ID:partner_internal_id,GRADE_SEQUENCE:grade_sequence,GRADE_CREATED_AT:grade_created_at,GRADE_UPDATED_AT:grade_updated_at,GRADE_DELETED_AT:grade_deleted_at,GRADE_IS_ARCHIVED:grade_is_archived,SCHOOL_LEVEL_GRADE_CREATED_AT:school_level_grade_created_at,SCHOOL_LEVEL_GRADE_UPDATED_AT:school_level_grade_updated_at,SCHOOL_LEVEL_GRADE_DELETED_AT:school_level_grade_deleted_at',
      'pk.fields'='level_id'
);

DROP CONNECTOR IF EXISTS SCHOOL_COURSE_SCHOOL_INFO_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SCHOOL_COURSE_SCHOOL_INFO_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}SCHOOL_COURSE_SCHOOL_INFO_PUBLIC_INFO_V1',
      'fields.whitelist'='school_course_id,school_course_name,school_course_name_phonetic,school_id,school_course_is_archived,school_course_partner_id,school_course_created_at,school_course_updated_at,school_course_deleted_at,school_name,school_name_phonetic,school_level_id,school_info_is_archived,school_info_created_at,school_info_updated_at,school_info_deleted_at,school_partner_id,address',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.school_course_school_info_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='SCHOOL_COURSE_ID:school_course_id,SCHOOL_COURSE_NAME:school_course_name,SCHOOL_COURSE_NAME_PHONETIC:school_course_name_phonetic,SCHOOL_ID:school_id,SCHOOL_COURSE_IS_ARCHIVED:school_course_is_archived,SCHOOL_COURSE_PARTNER_ID:school_course_partner_id,SCHOOL_COURSE_CREATED_AT:school_course_created_at,SCHOOL_COURSE_UPDATED_AT:school_course_updated_at,SCHOOL_COURSE_DELETED_AT:school_course_deleted_at,SCHOOL_NAME:school_name,SCHOOL_NAME_PHONETIC:school_name_phonetic,SCHOOL_LEVEL_ID:school_level_id,SCHOOL_INFO_IS_ARCHIVED:school_info_is_archived,SCHOOL_INFO_CREATED_AT:school_info_created_at,SCHOOL_INFO_UPDATED_AT:school_info_updated_at,SCHOOL_INFO_DELETED_AT:school_info_deleted_at,SCHOOL_PARTNER_ID:school_partner_id,ADDRESS:address',
      'pk.fields'='school_id'
);

DROP CONNECTOR IF EXISTS USER_ADDRESS_STREAM_FORMATED_V1;
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
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ADDRESS_ID:student_address_id,STUDENT_ID:student_id,ADDRESS_TYPE:address_type,POSTAL_CODE:postal_code,PREFECTURE_ID:prefecture_id,CITY:city,USER_ADDRESS_CREATED_AT:user_address_created_at,USER_ADDRESS_UPDATED_AT:user_address_updated_at,USER_ADDRESS_DELETED_AT:user_address_deleted_at,FIRST_STREET:first_street,SECOND_STREET:second_street',
      'pk.fields'='student_address_id'
);

DROP CONNECTOR IF EXISTS USER_PHONE_NUMBER_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS USER_PHONE_NUMBER_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_PHONE_NUMBER_PUBLIC_INFO_V1',
      'fields.whitelist'='user_id,user_phone_number_id,phone_number,type,updated_at,created_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_phone_number_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_ID:user_id,USER_PHONE_NUMBER_ID:user_phone_number_id,PHONE_NUMBER:phone_number,TYPE:type,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='user_phone_number_id'
);

DROP CONNECTOR IF EXISTS PERMISSION_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS PERMISSION_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PERMISSION_PUBLIC_INFO_V1',
      'fields.whitelist'='user_group_id,granted_permission_role_id,location_id,user_group_name,role_name,granted_permission_permission_name,permission_id,permission_created_at,permission_updated_at,permission_deleted_at,permission_permission_name,permission_role_role_id,permission_role_created_at,permission_role_updated_at,permission_role_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.permission_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_GROUP_ID:user_group_id,GRANTED_PERMISSION_ROLE_ID:granted_permission_role_id,LOCATION_ID:location_id,USER_GROUP_NAME:user_group_name,ROLE_NAME:role_name,GRANTED_PERMISSION_PERMISSION_NAME:granted_permission_permission_name,PERMISSION_ID:permission_id,PERMISSION_CREATED_AT:permission_created_at,PERMISSION_UPDATED_AT:permission_updated_at,PERMISSION_DELETED_AT:permission_deleted_at,PERMISSION_PERMISSION_NAME:permission_permission_name,PERMISSION_ROLE_ROLE_ID:permission_role_role_id,PERMISSION_ROLE_CREATED_AT:permission_role_created_at,PERMISSION_ROLE_UPDATED_AT:permission_role_updated_at,PERMISSION_ROLE_DELETED_AT:permission_role_deleted_at',
      'pk.fields'='permission_id'
);

DROP CONNECTOR IF EXISTS USER_GROUP_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS USER_GROUP_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_GROUP_PUBLIC_INFO_V1',
      'fields.whitelist'='user_group_id,user_group_name,created_at,updated_at,deleted_at,resource_path,org_location_id,is_system',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_group_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_GROUP_ID:user_group_id,USER_GROUP_NAME:user_group_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,RESOURCE_PATH:resource_path,ORG_LOCATION_ID:org_location_id,IS_SYSTEM:is_system',
      'pk.fields'='user_group_id'
);