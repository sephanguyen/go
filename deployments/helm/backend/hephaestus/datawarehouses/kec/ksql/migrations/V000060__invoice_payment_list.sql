DROP CONNECTOR IF EXISTS SINK_INVOICE_PAYMENT_LIST_PUBLIC_INFO;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_PAYMENT_LIST_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_PAYMENT_LIST_PUBLIC_INFO_V1',
      'fields.whitelist'='payment_id,invoice_id,invoice_sequence_number,student_id,payment_sequence_number,payment_status,payment_method,payment_due_date,payment_expiry_date,payment_date,amount,result_code,payment_created_at,payment_updated_at,invoice_created_at,invoice_updated_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.invoice_payment_list',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PAYMENT_ID:payment_id,INVOICE_ID:invoice_id,INVOICE_SEQUENCE_NUMBER:invoice_sequence_number,STUDENT_ID:student_id,PAYMENT_SEQUENCE_NUMBER:payment_sequence_number,PAYMENT_STATUS:payment_status,PAYMENT_METHOD:payment_method,PAYMENT_DUE_DATE:payment_due_date,PAYMENT_EXPIRY_DATE:payment_expiry_date,PAYMENT_DATE:payment_date,AMOUNT:amount,RESULT_CODE:result_code,PAYMENT_CREATED_AT:payment_created_at,PAYMENT_UPDATED_AT:payment_updated_at,INVOICE_CREATED_AT:invoice_created_at,INVOICE_UPDATED_AT:invoice_updated_at',
      'pk.fields'='payment_id'
);