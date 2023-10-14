SET 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS USER_GROUP_MEMBER_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_USER_GROUP_MEMBER_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_GROUP_MEMBER_PUBLIC_INFO_V1',
      'fields.whitelist'='user_id,user_group_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_group_member',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_ID:user_id,USER_GROUP_ID:user_group_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='user_id,user_group_id'
);

DROP CONNECTOR IF EXISTS SCHOOL_HISTORY_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_SCHOOL_HISTORY_PUBLIC_INFO_V1 WITH (
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
      'table.name.format'='bob.school_history',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ID:student_id,SCHOOL_ID:school_id,SCHOOL_COURSE_ID:school_course_id,START_DATE:start_date,END_DATE:end_date,IS_CURRENT:is_current,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='student_id,school_id'
);

DROP CONNECTOR IF EXISTS STUDENT_ENROLLMENT_STATUS_HISTORY_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_STUDENT_ENROLLMENT_STATUS_HISTORY_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_ENROLLMENT_STATUS_HISTORY_PUBLIC_INFO_V1',
      'fields.whitelist'='student_id,location_id,enrollment_status,start_date,end_date,comment,created_at,updated_at,deleted_at,order_id,order_sequence_number',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.student_enrollment_status_history',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ID:student_id,LOCATION_ID:location_id,ENROLLMENT_STATUS:enrollment_status,START_DATE:start_date,END_DATE:end_date,COMMENT:comment,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,ORDER_ID:order_id,ORDER_SEQUENCE_NUMBER:order_sequence_number',
      'pk.fields'='student_id,location_id,enrollment_status,start_date'
);
