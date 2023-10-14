SET 'auto.offset.reset' = 'earliest';

/* PRODUCT_GRADE */
CREATE STREAM IF NOT EXISTS PRODUCT_GRADE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.product_grade', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PRODUCT_GRADE_STREAM_FORMATTED_V1
AS SELECT
      PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID + PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID as KEY,
      PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
      PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID AS GRADE_ID,
      PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
      CAST(NULL AS VARCHAR) AS UPDATED_AT,
      CAST(NULL AS VARCHAR) AS DELETED_AT
FROM PRODUCT_GRADE_STREAM_ORIGIN_V1
WHERE PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID + PRODUCT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID
EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_PRODUCT_GRADE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_GRADE_STREAM_FORMATTED_V1',
      'fields.whitelist'='product_id,grade_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_grade',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_ID:product_id,GRADE_ID:grade_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='product_id,grade_id'
);

/* PRODUCT_PRICE */
CREATE STREAM IF NOT EXISTS PRODUCT_PRICE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.product_price', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PRODUCT_PRICE_STREAM_FORMATTED_V1
AS SELECT
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->PRODUCT_PRICE_ID AS KEY,
      AS_VALUE(PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->PRODUCT_PRICE_ID) AS PRODUCT_PRICE_ID,
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->BILLING_SCHEDULE_PERIOD_ID AS BILLING_SCHEDULE_PERIOD_ID,
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->QUANTITY AS QUANTITY,
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->PRICE AS PRICE,
      PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
      CAST(NULL AS VARCHAR) AS UPDATED_AT,
      CAST(NULL AS VARCHAR) AS DELETED_AT
FROM PRODUCT_PRICE_STREAM_ORIGIN_V1
WHERE PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PRODUCT_PRICE_STREAM_ORIGIN_V1.AFTER->PRODUCT_PRICE_ID
EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS SINK_PRODUCT_PRICE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_PRICE_STREAM_FORMATTED_V1',
      'fields.whitelist'='product_price_id,product_id,billing_schedule_period_id,quantity,price,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_price',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_PRICE_ID:product_price_id,PRODUCT_ID:product_id,BILLING_SCHEDULE_PERIOD_ID:billing_schedule_period_id,QUANTITY:quantity,PRICE:price,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='product_price_id'
);


/* PRODUCT_ACCOUNTING_CATEGORY */
CREATE STREAM IF NOT EXISTS PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.product_accounting_category', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PRODUCT_ACCOUNTING_CATEGORY_STREAM_FORMATTED_V1
AS SELECT
      PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID + PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID AS KEY,
      PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
      PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID AS ACCOUNTING_CATEGORY_ID,
      PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
      CAST(NULL AS VARCHAR) AS UPDATED_AT,
      CAST(NULL AS VARCHAR) AS DELETED_AT
FROM PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1
WHERE PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID + PRODUCT_ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID
EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS PRODUCT_ACCOUNTING_CATEGORY_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_ACCOUNTING_CATEGORY_STREAM_FORMATTED_V1',
      'fields.whitelist'='product_id,accounting_category_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_accounting_category',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_ID:product_id,ACCOUNTING_CATEGORY_ID:accounting_category_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='product_id,accounting_category_id'
);
