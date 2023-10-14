SET 'auto.offset.reset' = 'earliest';

CREATE TABLE IF NOT EXISTS INVOICE_SCHEDULE_PUBLIC_INFO_V1
AS SELECT
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.KEY AS INVOICE_SCHEDULE_ID,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.INVOICE_DATE AS INVOICE_DATE,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.SCHEDULED_DATE AS SCHEDULED_DATE,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.STATUS AS STATUS,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.IS_ARCHIVED AS IS_ARCHIVED,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.REMARKS AS REMARKS,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.USER_ID AS USER_ID,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.INVOICE_SCHEDULE_CREATED_AT AS CREATED_AT,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.INVOICE_SCHEDULE_UPDATED_AT AS UPDATED_AT,
    INVOICE_SCHEDULE_TABLE_FORMATTED_V1.INVOICE_SCHEDULE_DELETED_AT AS DELETED_AT
FROM INVOICE_SCHEDULE_TABLE_FORMATTED_V1;


CREATE SINK CONNECTOR IF NOT EXISTS SINK_INVOICE_SCHEDULE_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}INVOICE_SCHEDULE_PUBLIC_INFO_V1',
      'fields.whitelist'='invoice_schedule_id,invoice_date,scheduled_date,status,is_archived,remarks,user_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.invoice_schedule',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='INVOICE_SCHEDULE_ID:invoice_schedule_id,INVOICE_DATE:invoice_date,SCHEDULED_DATE:scheduled_date,STATUS:status,IS_ARCHIVED:is_archived,REMARKS:remarks,USER_ID:user_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='invoice_schedule_id'
);
