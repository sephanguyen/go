SET 'auto.offset.reset' = 'earliest';

/* LEAVING_REASON */

CREATE STREAM IF NOT EXISTS LEAVING_REASON_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.leaving_reason', value_format='AVRO');

CREATE STREAM IF NOT EXISTS LEAVING_REASON_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}LEAVING_REASON_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->LEAVING_REASON_ID AS KEY,
        AS_VALUE(LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->LEAVING_REASON_ID) AS LEAVING_REASON_ID,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->REMARK AS REMARK,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->LEAVING_REASON_TYPE AS LEAVING_REASON_TYPE,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM LEAVING_REASON_STREAM_ORIGIN_V1
    WHERE LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY LEAVING_REASON_STREAM_ORIGIN_V1.AFTER->LEAVING_REASON_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_LEAVING_REASON_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_LEAVING_REASON_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LEAVING_REASON_STREAM_FORMATED_V1',
      'fields.whitelist'='leaving_reason_id,leaving_reason_name,leaving_reason_type,leaving_reason_remark,leaving_reason_is_archived,leaving_reason_updated_at,leaving_reason_created_at,leaving_reason_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.leaving_reason',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LEAVING_REASON_ID:leaving_reason_id,NAME:leaving_reason_name,IS_ARCHIVED:leaving_reason_is_archived,CREATED_AT:leaving_reason_created_at,UPDATED_AT:leaving_reason_updated_at,REMARK:leaving_reason_remark,LEAVING_REASON_TYPE:leaving_reason_type,DELETED_AT:leaving_reason_deleted_at',
      'pk.fields'='leaving_reason_id'
);


/* ORDER_ITEM_COURSE */

CREATE STREAM IF NOT EXISTS ORDER_ITEM_COURSE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.order_item_course', value_format='AVRO');

CREATE STREAM IF NOT EXISTS ORDER_ITEM_COURSE_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}ORDER_ITEM_COURSE_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_COURSE_ID AS KEY,
        AS_VALUE(ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_COURSE_ID) AS ORDER_ITEM_COURSE_ID,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->ORDER_ID AS ORDER_ID,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID AS PACKAGE_ID,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_NAME AS COURSE_NAME,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT AS COURSE_SLOT,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT_PER_WEEK AS COURSE_SLOT_PER_WEEK,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM ORDER_ITEM_COURSE_STREAM_ORIGIN_V1
    WHERE ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY ORDER_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->ORDER_ITEM_COURSE_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_ORDER_ITEM_COURSE_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_ORDER_ITEM_COURSE_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ORDER_ITEM_COURSE_STREAM_FORMATED_V1',
      'fields.whitelist'='order_item_course_id_pk,order_id,package_id,course_id,course_name,course_slot,course_slot_per_week,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.order_item_course',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ORDER_ITEM_COURSE_ID:order_item_course_id,ORDER_ID:order_id,PACKAGE_ID:package_id,COURSE_ID:course_id,COURSE_NAME:course_name,COURSE_SLOT:course_slot,COURSE_SLOT_PER_WEEK:course_slot_per_week,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='order_item_course_id'
);

/* PACKAGE_QUANTITY_TYPE_MAPPING */

CREATE STREAM IF NOT EXISTS PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.package_quantity_type_mapping', value_format='AVRO');

CREATE STREAM IF NOT EXISTS PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->PACKAGE_TYPE AS KEY,
        AS_VALUE(PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->PACKAGE_TYPE) AS PACKAGE_TYPE,
        PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->QUANTITY_TYPE AS QUANTITY_TYPE,
        PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
    FROM PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1
    WHERE PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_ORIGIN_V1.AFTER->PACKAGE_TYPE
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_PACKAGE_QUANTITY_TYPE_MAPPING_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_PACKAGE_QUANTITY_TYPE_MAPPING_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PACKAGE_QUANTITY_TYPE_MAPPING_STREAM_FORMATED_V1',
      'fields.whitelist'='package_type,quantity_type,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.package_quantity_type_mapping',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PACKAGE_TYPE:package_type,QUANTITY_TYPE:quantity_type,CREATED_AT:created_at,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='package_type'
);



/* BILL_ITEM_COURSE */

CREATE STREAM IF NOT EXISTS BILL_ITEM_COURSE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.bill_item_course', value_format='AVRO');

CREATE STREAM IF NOT EXISTS BILL_ITEM_COURSE_STREAM_FORMATED_V1
AS SELECT
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->BILL_ITEM_SEQUENCE_NUMBER AS BILL_ITEM_SEQUENCE_NUMBER,
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_NAME AS COURSE_NAME, 
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_WEIGHT AS COURSE_WEIGHT,
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT AS COURSE_SLOT,
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_SLOT_PER_WEEK AS COURSE_SLOT_PER_WEEK,
              BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              CAST(NULL AS VARCHAR) AS DELETED_AT
       FROM BILL_ITEM_COURSE_STREAM_ORIGIN_V1
       WHERE BILL_ITEM_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';


DROP CONNECTOR IF EXISTS SINK_BILL_ITEM_COURSE_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_BILL_ITEM_COURSE_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}BILL_ITEM_COURSE_STREAM_FORMATED_V1',
      'fields.whitelist'='bill_item_sequence_number,course_id,course_name,course_weight,course_slot,course_slot_per_week,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.bill_item_course',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='BILL_ITEM_SEQUENCE_NUMBER:bill_item_sequence_number,COURSE_ID:course_id,COURSE_NAME:course_name,COURSE_WEIGHT:course_weight,COURSE_SLOT:course_slot,COURSE_SLOT_PER_WEEK:course_slot_per_week,CREATED_AT:created_at,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='bill_item_sequence_number,course_id'
);

/* STUDENT_PRODUCT */

CREATE STREAM IF NOT EXISTS STUDENT_PRODUCT_STREAM_ORIGIN_V1
  WITH (KAFKA_TOPIC='{{ .Values.global.environment }}.kec.datalake.fatima.student_product',
        VALUE_FORMAT='AVRO');

CREATE STREAM IF NOT EXISTS STUDENT_PRODUCT_STREAM_FORMATED_V1
  WITH (KAFKA_TOPIC='{{ .Values.topicPrefix }}STUDENT_PRODUCT_STREAM_FORMATED_V1',
        VALUE_FORMAT='AVRO')
  AS SELECT
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID AS KEY,
        AS_VALUE(STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID) AS STUDENT_PRODUCT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->UPCOMING_BILLING_DATE AS UPCOMING_BILLING_DATE,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->START_DATE AS START_DATE,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->END_DATE AS END_DATE,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_STATUS AS PRODUCT_STATUS,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->APPROVAL_STATUS AS APPROVAL_STATUS,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->UPDATED_FROM_STUDENT_PRODUCT_ID AS UPDATED_FROM_STUDENT_PRODUCT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->UPDATED_TO_STUDENT_PRODUCT_ID AS UPDATED_TO_STUDENT_PRODUCT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_LABEL AS STUDENT_PRODUCT_LABEL,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->IS_UNIQUE AS IS_UNIQUE,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->ROOT_STUDENT_PRODUCT_ID AS ROOT_STUDENT_PRODUCT_ID,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->IS_ASSOCIATED AS IS_ASSOCIATED,
        STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->VERSION_NUMBER AS VERSION_NUMBER
  FROM STUDENT_PRODUCT_STREAM_ORIGIN_V1
  WHERE STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY STUDENT_PRODUCT_STREAM_ORIGIN_V1.AFTER->STUDENT_PRODUCT_ID
EMIT CHANGES;

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_STUDENT_PRODUCT_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_STUDENT_PRODUCT_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}STUDENT_PRODUCT_STREAM_FORMATED_V1',
      'fields.whitelist'='student_product_id,student_id,product_id,upcoming_billing_date,start_date,end_date,product_status,approval_status,updated_at,created_at,deleted_at,location_id,updated_from_student_product_id,updated_to_student_product_id,student_product_label,is_unique,root_student_product_id,is_associated,version_number',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.student_product',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='STUDENT_PRODUCT_ID:student_product_id,STUDENT_ID:student_id,PRODUCT_ID:product_id,UPCOMING_BILLING_DATE:upcoming_billing_date,START_DATE:start_date,END_DATE:end_date,PRODUCT_STATUS:product_status,APPROVAL_STATUS:approval_status,UPDATED_AT:updated_at,CREATED_AT:created_at,DELETED_AT:deleted_at,LOCATION_ID:location_id,UPDATED_FROM_STUDENT_PRODUCT_ID:updated_from_student_product_id,UPDATED_TO_STUDENT_PRODUCT_ID:updated_to_student_product_id,STUDENT_PRODUCT_LABEL:student_product_label,IS_UNIQUE:is_unique,ROOT_STUDENT_PRODUCT_ID:root_student_product_id,IS_ASSOCIATED:is_associated,VERSION_NUMBER:version_number',
      'pk.fields'='student_product_id'
);



/* order_action_log */

-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS ORDER_ACTION_LOG_STREAM_ORIGIN_V1
  WITH (KAFKA_TOPIC='{{ .Values.global.environment }}.kec.datalake.fatima.order_action_log',
        VALUE_FORMAT='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS ORDER_ACTION_LOG_STREAM_FORMATED_V1
  WITH (KAFKA_TOPIC='{{ .Values.topicPrefix }}ORDER_ACTION_LOG_STREAM_FORMATED_V1',
        VALUE_FORMAT='AVRO')
  AS SELECT
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ORDER_ACTION_LOG_ID AS KEY,
        AS_VALUE(ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ORDER_ACTION_LOG_ID) AS ORDER_ACTION_LOG_ID,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ORDER_ID AS ORDER_ID,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ACTION AS ACTION,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->COMMENT AS COMMENT,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
  FROM ORDER_ACTION_LOG_STREAM_ORIGIN_V1
  WHERE ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY ORDER_ACTION_LOG_STREAM_ORIGIN_V1.AFTER->ORDER_ACTION_LOG_ID
EMIT CHANGES;

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_ORDER_ACTION_LOG_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_ORDER_ACTION_LOG_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ORDER_ACTION_LOG_STREAM_FORMATED_V1',
      'fields.whitelist'='order_action_log_id,staff_id,order_id,action,comment,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.order_action_log',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ORDER_ACTION_LOG_ID:order_action_log_id,USER_ID:staff_id,ORDER_ID:order_id,ACTION:action,COMMENT:comment,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='order_action_log_id'
);


/* product */

-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS PRODUCT_STREAM_ORIGIN_V1
  WITH (KAFKA_TOPIC='{{ .Values.global.environment }}.kec.datalake.fatima.product',
        VALUE_FORMAT='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS PRODUCT_STREAM_FORMATED_V1
  WITH (KAFKA_TOPIC='{{ .Values.topicPrefix }}PRODUCT_STREAM_FORMATED_V1',
        VALUE_FORMAT='AVRO')
  AS SELECT
        PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS KEY,
        AS_VALUE(PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID) AS PRODUCT_ID,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->NAME AS name,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_TYPE AS PRODUCT_TYPE,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->TAX_ID AS TAX_ID,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->AVAILABLE_FROM AS AVAILABLE_FROM,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->AVAILABLE_UNTIL AS AVAILABLE_UNTIL,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->REMARKS AS REMARKS,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->CUSTOM_BILLING_PERIOD AS CUSTOM_BILLING_PERIOD,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->BILLING_SCHEDULE_ID AS BILLING_SCHEDULE_ID,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->DISABLE_PRO_RATING_FLAG AS DISABLE_PRO_RATING_FLAG,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->IS_UNIQUE AS IS_UNIQUE,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_TAG AS PRODUCT_TAG,
        PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_PARTNER_ID AS PRODUCT_PARTNER_ID
  FROM PRODUCT_STREAM_ORIGIN_V1
  WHERE PRODUCT_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PRODUCT_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID
EMIT CHANGES;

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_PRODUCT_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_PRODUCT_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_STREAM_FORMATED_V1',
      'fields.whitelist'='product_id,name,product_type,tax_id,available_from,available_until,remarks,custom_billing_period,billing_schedule_id,disable_pro_rating_flag,is_archived,updated_at,created_at,deleted_at,is_unique,product_tag,product_partner_id',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_ID:product_id,NAME:name,PRODUCT_TYPE:product_type,TAX_ID:tax_id,AVAILABLE_FROM:available_from,AVAILABLE_UNTIL:available_until,REMARKS:remarks,CUSTOM_BILLING_PERIOD:custom_billing_period,BILLING_SCHEDULE_ID:billing_schedule_id,DISABLE_PRO_RATING_FLAG:disable_pro_rating_flag,IS_ARCHIVED:is_archived,UPDATED_AT:updated_at,CREATED_AT:created_at,IS_UNIQUE:is_unique,PRODUCT_TAG:product_tag,PRODUCT_PARTNER_ID:product_partner_id,DELETED_AT:deleted_at',
      'pk.fields'='product_id'
);



-- ACCOUNTING_CATEGORY --
-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1
  WITH (KAFKA_TOPIC='{{ .Values.global.environment }}.kec.datalake.fatima.accounting_category',
        VALUE_FORMAT='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS ACCOUNTING_CATEGORY_STREAM_FORMATED_V1
  WITH (KAFKA_TOPIC='{{ .Values.topicPrefix }}ACCOUNTING_CATEGORY_STREAM_FORMATED_V1',
        VALUE_FORMAT='AVRO')
  AS SELECT
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID AS KEY,
        AS_VALUE(ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID) AS ACCOUNTING_CATEGORY_ID,
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->NAME AS name,
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->REMARKS AS REMARKS,
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT
  FROM ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1
  WHERE ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY ACCOUNTING_CATEGORY_STREAM_ORIGIN_V1.AFTER->ACCOUNTING_CATEGORY_ID
EMIT CHANGES;

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_ACCOUNTING_CATEGORY_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_ACCOUNTING_CATEGORY_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}ACCOUNTING_CATEGORY_STREAM_FORMATED_V1',
      'fields.whitelist'='accounting_category_id,name,remarks,is_archived,updated_at,created_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.accounting_category',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='ACCOUNTING_CATEGORY_ID:accounting_category_id,NAME:name,REMARKS:remarks,IS_ARCHIVED:is_archived,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='accounting_category_id'
);

-- product_setting --

-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS PRODUCT_SETTING_STREAM_ORIGIN_V1
  WITH (KAFKA_TOPIC='{{ .Values.global.environment }}.kec.datalake.fatima.product_setting',
        VALUE_FORMAT='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS PRODUCT_SETTING_STREAM_FORMATED_V1
  WITH (KAFKA_TOPIC='{{ .Values.topicPrefix }}PRODUCT_SETTING_STREAM_FORMATED_V1',
        VALUE_FORMAT='AVRO')
  AS SELECT
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS KEY,
        AS_VALUE(PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID) AS PRODUCT_ID,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->IS_ENROLLMENT_REQUIRED AS IS_ENROLLMENT_REQUIRED,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        CAST(NULL AS VARCHAR) AS DELETED_AT,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->IS_PAUSABLE AS IS_PAUSABLE,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->IS_ADDED_TO_ENROLLMENT_BY_DEFAULT AS IS_ADDED_TO_ENROLLMENT_BY_DEFAULT,
        PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->IS_OPERATION_FEE AS IS_OPERATION_FEE
  FROM PRODUCT_SETTING_STREAM_ORIGIN_V1
  WHERE PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY PRODUCT_SETTING_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID
EMIT CHANGES;

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_PRODUCT_SETTING_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_PRODUCT_SETTING_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_SETTING_STREAM_FORMATED_V1',
      'fields.whitelist'='product_id,is_enrollment_required,created_at,updated_at,deleted_at,is_pausable,is_added_to_enrollment_by_default,is_operation_fee',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_setting',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_ID:product_id,IS_ENROLLMENT_REQUIRED:is_enrollment_required,CREATED_AT:created_at,UPDATED_AT:updated_at,IS_PAUSABLE:is_pausable,IS_ADDED_TO_ENROLLMENT_BY_DEFAULT:is_added_to_enrollment_by_default,IS_OPERATION_FEE:is_operation_fee,DELETED_AT:deleted_at',
      'pk.fields'='product_id'
);


-- package_course --


-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_STREAM_ORIGIN_V1
  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.package_course', value_format='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS PACKAGE_COURSE_STREAM_FORMATED_V1
  AS SELECT
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->PACKAGE_ID AS PACKAGE_ID,
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_ID AS COURSE_ID,
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->MANDATORY_FLAG AS MANDATORY_FLAG, 
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->COURSE_WEIGHT AS COURSE_WEIGHT,
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              CAST(NULL AS VARCHAR) AS DELETED_AT,
              PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->MAX_SLOTS_PER_COURSE AS MAX_SLOTS_PER_COURSE
       FROM PACKAGE_COURSE_STREAM_ORIGIN_V1
       WHERE PACKAGE_COURSE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_PACKAGE_COURSE_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_PACKAGE_COURSE_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PACKAGE_COURSE_STREAM_FORMATED_V1',
      'fields.whitelist'='package_id,course_id,mandatory_flag,course_weight,created_at,updated_at,deleted_at,max_slots_per_course',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.package_course',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PACKAGE_ID:package_id,COURSE_ID:course_id,MANDATORY_FLAG:mandatory_flag,COURSE_WEIGHT:course_weight,CREATED_AT:created_at,MAX_SLOTS_PER_COURSE:max_slots_per_course,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='package_id,course_id'
);


-- PRODUCT LOCATION --

-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS PRODUCT_LOCATION_STREAM_ORIGIN_V1
  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.product_location', value_format='AVRO');

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS PRODUCT_LOCATION_STREAM_FORMATED_V1
  AS SELECT
              PRODUCT_LOCATION_STREAM_ORIGIN_V1.AFTER->PRODUCT_ID AS PRODUCT_ID,
              PRODUCT_LOCATION_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
              PRODUCT_LOCATION_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
              CAST(NULL AS VARCHAR) AS DELETED_AT
       FROM PRODUCT_LOCATION_STREAM_ORIGIN_V1
       WHERE PRODUCT_LOCATION_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

-- Drop the existing sink connector if it exists
DROP CONNECTOR IF EXISTS SINK_PRODUCT_LOCATION_TABLE_FORMATED_V1;

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_PRODUCT_LOCATION_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PRODUCT_LOCATION_STREAM_FORMATED_V1',
      'fields.whitelist'='product_id,location_id,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.product_location',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='PRODUCT_ID:product_id,LOCATION_ID:location_id,CREATED_AT:created_at,CREATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='product_id,location_id'
);
