DROP CONNECTOR IF EXISTS USER_ADDRESS_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_USER_ADDRESS_PUBLIC_INFO_V1 WITH (
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
      'table.name.format'='bob.user_address',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_ADDRESS_ID:student_address_id,STUDENT_ID:student_id,ADDRESS_TYPE:address_type,POSTAL_CODE:postal_code,PREFECTURE_ID:prefecture_id,CITY:city,USER_ADDRESS_CREATED_AT:user_address_created_at,USER_ADDRESS_UPDATED_AT:user_address_updated_at,USER_ADDRESS_DELETED_AT:user_address_deleted_at,FIRST_STREET:first_street,SECOND_STREET:second_street',
      'pk.fields'='student_address_id'
);

DROP CONNECTOR IF EXISTS USER_PHONE_NUMBER_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_USER_PHONE_NUMBER_PUBLIC_INFO_V1 WITH (
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
      'table.name.format'='bob.user_phone_number',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_ID:user_id,USER_PHONE_NUMBER_ID:user_phone_number_id,PHONE_NUMBER:phone_number,TYPE:type,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='user_phone_number_id'
);

DROP CONNECTOR IF EXISTS USER_GROUP_STREAM_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_USER_GROUP_PUBLIC_INFO_V1 WITH (
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
      'table.name.format'='bob.user_group',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_GROUP_ID:user_group_id,USER_GROUP_NAME:user_group_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,RESOURCE_PATH:resource_path,ORG_LOCATION_ID:org_location_id,IS_SYSTEM:is_system',
      'pk.fields'='user_group_id'
);