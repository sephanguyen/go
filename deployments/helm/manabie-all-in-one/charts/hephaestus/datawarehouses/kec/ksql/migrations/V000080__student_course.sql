SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS FILE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.file', value_format='AVRO');
CREATE STREAM IF NOT EXISTS FILE_STREAM_FORMATTED_V1
    AS SELECT
        FILE_STREAM_ORIGIN_V1.AFTER->FILE_ID as KEY,
        AS_VALUE(FILE_STREAM_ORIGIN_V1.AFTER->FILE_ID) AS FILE_ID,
        FILE_STREAM_ORIGIN_V1.AFTER->FILE_NAME AS FILE_NAME,
        FILE_STREAM_ORIGIN_V1.AFTER->FILE_TYPE AS FILE_TYPE,
        FILE_STREAM_ORIGIN_V1.AFTER->DOWNLOAD_LINK AS DOWNLOAD_LINK,
        FILE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        FILE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM FILE_STREAM_ORIGIN_V1
    WHERE FILE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY FILE_STREAM_ORIGIN_V1.AFTER->FILE_ID
    EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS FILE_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}FILE_STREAM_FORMATTED_V1',
      'fields.whitelist'='file_id,file_name,file_type,download_link,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.file',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='FILE_ID:file_id,FILE_NAME:file_name,FILE_TYPE:file_type,DOWNLOAD_LINK:download_link,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='file_id'
);


CREATE STREAM IF NOT EXISTS ORDER_ITEM_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.order_item', value_format='AVRO');
CREATE STREAM IF NOT EXISTS ORDER_ITEM_STREAM_FORMATTED_V1
    AS SELECT
       ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_ID AS KEY,
        AS_VALUE(ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_ID) AS ORDER_ITEM_ID,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->ORDER_ID AS ORDER_ID,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->DISCOUNT_ID AS DISCOUNT_ID,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID AS STUDENT_PRODUCT_ID,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->PRODUCT_NAME AS PRODUCT_NAME,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->EFFECTIVE_DATE AS EFFECTIVE_DATE,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->CANCELLATION_DATE AS CANCELLATION_DATE,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS ORDER_ITEM_CREATED_AT,
        ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS ORDER_ITEM_UPDATED_AT,
        CAST(NULL AS VARCHAR) AS ORDER_ITEM_DELETED_AT
    FROM ORDER_ITEM_STREAM_ORIGIN_V1
    WHERE ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY ORDER_ITEM_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_ID
    EMIT CHANGES;

CREATE SINK CONNECTOR IF NOT EXISTS ORDER_ITEM_PUBLIC_INFO WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ORDER_ITEM_STREAM_FORMATTED_V1',
      'fields.whitelist'='order_item_id,order_id,product_id,discount_id,start_date,student_product_id,product_name,effective_date,cancellation_date,end_date,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.order_item',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ORDER_ITEM_ID:order_item_id,ORDER_ID:order_id,PRODUCT_ID:product_id,DISCOUNT_ID:discount_id,START_DATE:start_date,STUDENT_PRODUCT_ID:student_product_id,PRODUCT_NAME:product_name,EFFECTIVE_DATE:effective_date,CANCELLATION_DATE:cancellation_date,END_DATE:end_date,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='order_item_id'
);
