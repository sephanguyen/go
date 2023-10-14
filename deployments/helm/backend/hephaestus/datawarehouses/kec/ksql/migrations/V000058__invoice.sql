DROP CONNECTOR IF EXISTS SINK_INVOICE_PUBLIC_INFO;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_key,invoice_id,invoice_sequence_number,type,status,student_id,sub_total,total,outstanding_balance,amount_paid,amount_refunded,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.invoice',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:invoice_key,INVOICE_ID:invoice_id,INVOICE_SEQUENCE_NUMBER:invoice_sequence_number,INVOICE_TYPE:type,INVOICE_STATUS:status,STUDENT_ID:student_id,SUB_TOTAL:sub_total,TOTAL:total,OUTSTANDING_BALANCE:outstanding_balance,AMOUNT_PAID:amount_paid,AMOUNT_REFUNDED:amount_refunded,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='invoice_id'
);