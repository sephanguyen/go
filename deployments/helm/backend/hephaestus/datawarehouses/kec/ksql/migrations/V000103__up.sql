DROP CONNECTOR IF EXISTS TS_LESSON_V2;
CREATE SINK CONNECTOR TS_LESSON_V3 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TS_LESSON_V1',
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
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LESSON_ID:lesson_id,TIMESHEET_ID:timesheet_id,FLAG_ON:flag_on,TIMESHEET_LESSON_HOUR_CREATED_AT:timesheet_lesson_hour_created_at,TIMESHEET_LESSON_HOUR_UPDATED_AT:timesheet_lesson_hour_updated_at,TIMESHEET_LESSON_HOUR_DELETED_AT:timesheet_lesson_hour_deleted_at,STAFF_ID:staff_id,TIMESHEET_STATUS:timesheet_status,TIMESHEET_DATE:timesheet_date,TIMESHEET_REMARK:timesheet_remark,LOCATION_ID:location_id,TIMESHEET_CREATED_AT:timesheet_created_at,TIMESHEET_UPDATED_AT:timesheet_updated_at,TIMESHEET_DELETED_AT:timesheet_deleted_at',
      'pk.fields'='lesson_id'
);
