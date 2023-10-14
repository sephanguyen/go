SET 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS TS_TRANSPORTATION_V1;
CREATE SINK CONNECTOR TS_TRANSPORTATION_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TS_TRANSPORTATION',
      'fields.whitelist'='transportation_expense_id,timesheet_id,transportation_type,transportation_from,transportation_to,cost_amount,round_trip,transportation_expense_remarks,transportation_expense_created_at,transportation_expense_updated_at,transportation_expense_deleted_at,timesheet_status,timesheet_date,timesheet_remark,timesheet_created_at,timesheet_updated_at,timesheet_deleted_at,staff_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.ts_transportation',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='TRANSPORTATION_EXPENSE_ID:transportation_expense_id,TIMESHEET_ID:timesheet_id,TRANSPORTATION_TYPE:transportation_type,TRANSPORTATION_FROM:transportation_from,TRANSPORTATION_TO:transportation_to,COST_AMOUNT:cost_amount,ROUND_TRIP:round_trip,TRANSPORTATION_EXPENSE_REMARKS:transportation_expense_remarks,TRANSPORTATION_EXPENSE_CREATED_AT:transportation_expense_created_at,TRANSPORTATION_EXPENSE_UPDATED_AT:transportation_expense_updated_at,TRANSPORTATION_EXPENSE_DELETED_AT:transportation_expense_deleted_at,TIMESHEET_STATUS:timesheet_status,TIMESHEET_DATE:timesheet_date,TIMESHEET_REMARK:timesheet_remark,TIMESHEET_CREATED_AT:timesheet_created_at,TIMESHEET_UPDATED_AT:timesheet_updated_at,TIMESHEET_DELETED_AT:timesheet_deleted_at,STAFF_ID:staff_id',
      'pk.fields'='transportation_expense_id'
); 


DROP CONNECTOR IF EXISTS STAFF_TRANSPORTATION_EXPENSE_V1;
CREATE SINK CONNECTOR STAFF_TRANSPORTATION_EXPENSE_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STAFF_TRANSPORTATION_EXPENSE',
      'fields.whitelist'='staff_transportation_expense_id,staff_id,location_id,transport_type,transportation_from,transportation_to,cost_amount,round_trip,remarks,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.staff_transportation_expense',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STAFF_TRANSPORTATION_EXPENSE_ID:staff_transportation_expense_id,STAFF_ID:staff_id,LOCATION_ID:location_id,TRANSPORT_TYPE:transport_type,TRANSPORTATION_FROM:transportation_from,TRANSPORTATION_TO:transportation_to,COST_AMOUNT:cost_amount,ROUND_TRIP:round_trip,REMARKS:remarks,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='staff_transportation_expense_id'
);


DROP CONNECTOR IF EXISTS TS_LESSON_V1;
CREATE SINK CONNECTOR TS_LESSON_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TS_LESSON',
      'fields.whitelist'='lesson_id,timesheet_id,flag_on,timesheet_lesson_hour_created_at,timesheet_lesson_hour_updated_at,timesheet_lesson_hour_deleted_at,staff_id,timesheet_status,timesheet_date,timesheet_remark,location_id,timesheet_created_at,timesheet_updated_at,timesheet_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.ts_lesson',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,TIMESHEET_ID:timesheet_id,FLAG_ON:flag_on,TIMESHEET_LESSON_HOUR_CREATED_AT:timesheet_lesson_hour_created_at,TIMESHEET_LESSON_HOUR_UPDATED_AT:timesheet_lesson_hour_updated_at,TIMESHEET_LESSON_HOUR_DELETED_AT:timesheet_lesson_hour_deleted_at,STAFF_ID:staff_id,TIMESHEET_STATUS:timesheet_status,TIMESHEET_DATE:timesheet_date,TIMESHEET_REMARK:timesheet_remark,LOCATION_ID:location_id,TIMESHEET_CREATED_AT:timesheet_created_at,TIMESHEET_UPDATED_AT:timesheet_updated_at,TIMESHEET_DELETED_AT:timesheet_deleted_at',
      'pk.fields'='lesson_id'
);


DROP CONNECTOR IF EXISTS TS_OTHER_WORKING_V1;
CREATE SINK CONNECTOR TS_OTHER_WORKING_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TS_OTHER_WORKING',
      'fields.whitelist'='other_working_hours_id,timesheet_id,timesheet_config_id,start_time,end_time,total_hour,other_working_hour_remarks,other_working_hour_created_at,other_working_hour_updated_at,other_working_hour_deleted_at,config_type,config_value,timesheet_config_created_at,timesheet_config_updated_at,timesheet_config_deleted_at,staff_id,timesheet_status,timesheet_date,timesheet_remark,location_id,timesheet_created_at,timesheet_updated_at,timesheet_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.ts_other_working',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='OTHER_WORKING_HOURS_ID:other_working_hours_id,TIMESHEET_ID:timesheet_id,TIMESHEET_CONFIG_ID:timesheet_config_id,START_TIME:start_time,END_TIME:end_time,TOTAL_HOUR:total_hour,OTHER_WORKING_HOUR_REMARKS:other_working_hour_remarks,OTHER_WORKING_HOUR_CREATED_AT:other_working_hour_created_at,OTHER_WORKING_HOUR_UPDATED_AT:other_working_hour_updated_at,OTHER_WORKING_HOUR_DELETED_AT:other_working_hour_deleted_at,CONFIG_TYPE:config_type,CONFIG_VALUE:config_value,TIMESHEET_CONFIG_CREATED_AT:timesheet_config_created_at,TIMESHEET_CONFIG_UPDATED_AT:timesheet_config_updated_at,TIMESHEET_CONFIG_DELETED_AT:timesheet_config_deleted_at,STAFF_ID:staff_id,TIMESHEET_STATUS:timesheet_status,TIMESHEET_DATE:timesheet_date,TIMESHEET_REMARK:timesheet_remark,LOCATION_ID:location_id,TIMESHEET_CREATED_AT:timesheet_created_at,TIMESHEET_UPDATED_AT:timesheet_updated_at,TIMESHEET_DELETED_AT:timesheet_deleted_at',
      'pk.fields'='other_working_hours_id'
);


DROP CONNECTOR IF EXISTS AUTO_CREATE_FLAG_ACTIVITY_LOG_V1;
CREATE SINK CONNECTOR AUTO_CREATE_FLAG_ACTIVITY_LOG_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}AUTO_CREATE_FLAG_ACTIVITY_LOG',
      'fields.whitelist'='auto_create_flag_activity_log_id,staff_id,change_time,flag_on,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.auto_create_flag_activity_log',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='AUTO_CREATE_FLAG_ACTIVITY_LOG_ID:auto_create_flag_activity_log_id,STAFF_ID:staff_id,CHANGE_TIME:change_time,FLAG_ON:flag_on,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='auto_create_flag_activity_log_id'
);


DROP CONNECTOR IF EXISTS AUTO_CREATE_TIMESHEET_FLAG_V1;
CREATE SINK CONNECTOR AUTO_CREATE_TIMESHEET_FLAG_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}AUTO_CREATE_TIMESHEET_FLAG',
      'fields.whitelist'='staff_id,flag_on,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.auto_create_timesheet_flag',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STAFF_ID:staff_id,FLAG_ON:flag_on,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='staff_id'
);


DROP CONNECTOR IF EXISTS TIMESHEET_CONFIRMATION_CUT_OFF_DATE_V1;
CREATE SINK CONNECTOR TIMESHEET_CONFIRMATION_CUT_OFF_DATE_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TIMESHEET_CONFIRMATION_CUT_OFF_DATE',
      'fields.whitelist'='timesheet_confirmation_cut_off_date_id,cut_off_date,start_date,end_date,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.timesheet_confirmation_cut_off_date',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='TIMESHEET_CONFIRMATION_CUT_OFF_DATE_ID:timesheet_confirmation_cut_off_date_id,CUT_OFF_DATE:cut_off_date,START_DATE:start_date,END_DATE:end_date,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='timesheet_confirmation_cut_off_date_id'
);


DROP CONNECTOR IF EXISTS TIMESHEET_CONFIRMATION_INFO_V1;
CREATE SINK CONNECTOR TIMESHEET_CONFIRMATION_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TIMESHEET_CONFIRMATION_INFO',
      'fields.whitelist'='timesheet_confirmation_info_id,period_id,location_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.timesheet_confirmation_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='TIMESHEET_CONFIRMATION_INFO_ID:timesheet_confirmation_info_id,PERIOD_ID:period_id,LOCATION_ID:location_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='timesheet_confirmation_info_id'
);


DROP CONNECTOR IF EXISTS TIMESHEET_CONFIRMATION_PERIOD_V1;
CREATE SINK CONNECTOR TIMESHEET_CONFIRMATION_PERIOD_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TIMESHEET_CONFIRMATION_PERIOD',
      'fields.whitelist'='timesheet_confirmation_period_id,start_date,end_date,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.timesheet_confirmation_period',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='TIMESHEET_CONFIRMATION_PERIOD_ID:timesheet_confirmation_period_id,START_DATE:start_date,END_DATE:end_date,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='timesheet_confirmation_period_id'
);
