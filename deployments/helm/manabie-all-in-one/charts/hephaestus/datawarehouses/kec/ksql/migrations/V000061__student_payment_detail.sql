DROP CONNECTOR IF EXISTS SINK_STUDENT_PAYMENT_DETAIL_HISTORY_PUBLIC_INFO_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_STUDENT_PAYMENT_DETAIL_HISTORY_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_PAYMENT_DETAIL_HISTORY_PUBLIC_INFO_V1',
      'fields.whitelist'='student_payment_detail_action_id,student_payment_detail_id,student_id,payment_method,staff_id,action_type,created_at,updated_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.student_payment_detail',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:student_payment_detail_action_id,STUDENT_PAYMENT_DETAIL_ID:student_payment_detail_id,STUDENT_ID:student_id,PAYMENT_METHOD:payment_method,STAFF_ID:staff_id,ACTION_TYPE:action_type,CREATED_AT:created_at, UPDATED_AT:updated_at',
      'pk.fields'='student_payment_detail_action_id'
);