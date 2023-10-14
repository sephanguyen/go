SET 'auto.offset.reset' = 'earliest';

/* PACKAGE_COURSE_FEE */
CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.package_course_fee', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_FEE_STREAM_FORMATTED_V1
AS SELECT
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID + PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->COURSE_ID + PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->FEE_ID  as KEY,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->FEE_ID AS FEE_ID,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID AS PACKAGE_ID,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->QUANTITY AS QUANTITY,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->AVAILABLE_FROM AS AVAILABLE_FROM,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->AVAILABLE_UNTIL AS AVAILABLE_UNTIL,
      PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
      CAST(NULL AS VARCHAR) AS UPDATED_AT,
      CAST(NULL AS VARCHAR) AS DELETED_AT
FROM PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1
WHERE PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID + PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->COURSE_ID + PACKAGE_COURSE_FEE_STREAM_ORIGIN_V1.AFTER->FEE_ID
EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS PACKAGE_COURSE_FEE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PACKAGE_COURSE_FEE_STREAM_FORMATTED_V1',
      'fields.whitelist'='package_id,course_id,fee_id,quantity,available_from,available_until,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.package_course_fee',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PACKAGE_ID:package_id,COURSE_ID:course_id,FEE_ID:fee_id,QUANTITY:quantity,AVAILABLE_FROM:available_from,AVAILABLE_UNTIL:available_until,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='package_id,course_id,fee_id'
);


/* PACKAGE_COURSE_MATERIAL */
CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.package_course_material', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_MATERIAL_STREAM_FORMATTED_V1
AS SELECT
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID + PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->COURSE_ID + PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->MATERIAL_ID as KEY,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID AS PACKAGE_ID,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->MATERIAL_ID AS MATERIAL_ID,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->AVAILABLE_FROM AS AVAILABLE_FROM,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->AVAILABLE_UNTIL AS AVAILABLE_UNTIL,
      PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
      CAST(NULL AS VARCHAR) AS UPDATED_AT,
      CAST(NULL AS VARCHAR) AS DELETED_AT
FROM PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1
WHERE PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID + PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->COURSE_ID + PACKAGE_COURSE_MATERIAL_STREAM_ORIGIN_V1.AFTER->MATERIAL_ID
EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS PACKAGE_COURSE_MATERIAL_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PACKAGE_COURSE_MATERIAL_STREAM_FORMATTED_V1',
      'fields.whitelist'='package_id,course_id,material_id,available_from,available_until,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.package_course_material',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PACKAGE_ID:package_id,COURSE_ID:course_id,MATERIAL_ID:material_id,AVAILABLE_FROM:available_from,AVAILABLE_UNTIL:available_until,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='package_id,course_id,material_id'
);

