DROP CONNECTOR IF EXISTS SINK_INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_BILL_ITEM_LIST_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_bill_item_id,invoice_id,invoice_sequence_number,student_id,bill_item_sequence_number,invoice_bill_item_created_at,invoice_bill_item_updated_at,invoice_created_at,invoice_updated_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.invoice_bill_item_list',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='KEY:invoice_bill_item_id,INVOICE_ID:invoice_id,INVOICE_SEQUENCE_NUMBER:invoice_sequence_number,STUDENT_ID:student_id,BILL_ITEM_SEQUENCE_NUMBER:bill_item_sequence_number,INVOICE_BILL_ITEM_CREATED_AT:invoice_bill_item_created_at,INVOICE_BILL_ITEM_UPDATED_AT:invoice_bill_item_updated_at,INVOICE_CREATED_AT:invoice_created_at,INVOICE_UPDATED_AT:invoice_updated_at',
      'pk.fields'='invoice_bill_item_id'
);