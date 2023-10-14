set 'auto.offset.reset' = 'earliest';

DROP CONNECTOR IF EXISTS SINK_LESSON_PARTNER_FORM_CONFIGS_PUBLIC_INFO_V1;
DROP CONNECTOR IF EXISTS SINK_PARTNER_DYNAMIC_FORM_FIELD_VALUES_PUBLIC_INFO_V1;

DROP STREAM IF EXISTS LESSON_PARTNER_FORM_CONFIGS_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1;
DROP STREAM IF EXISTS PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_FORMATED_V1 DELETE TOPIC;
DROP STREAM IF EXISTS PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1;

CREATE STREAM IF NOT EXISTS LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.partner_form_configs', value_format='AVRO');
CREATE STREAM IF NOT EXISTS LESSON_PARTNER_FORM_CONFIGS_STREAM_FORMATED_V1
AS SELECT
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->FORM_CONFIG_ID AS rowkey,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->PARTNER_ID AS PARTNER_ID,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->FEATURE_NAME AS FEATURE_NAME,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
   LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->FORM_CONFIG_DATA AS FORM_CONFIG_DATA,
   AS_VALUE(LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->FORM_CONFIG_ID) AS FORM_CONFIG_ID
   FROM LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1
   WHERE LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_LESSON_PARTNER_FORM_CONFIGS_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}LESSON_PARTNER_FORM_CONFIGS_STREAM_FORMATED_V1',
      'fields.whitelist'='form_config_id,partner_id,feature_name,created_at,updated_at,deleted_at,form_config_data',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.partner_form_configs',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='FORM_CONFIG_ID:form_config_id,PARTNER_ID:partner_id,FEATURE_NAME:feature_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,FORM_CONFIG_DATA:form_config_data',
      'pk.fields'='form_config_id'
);

CREATE STREAM IF NOT EXISTS PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.lessonmgmt.partner_dynamic_form_field_values', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_FORMATED_V1
AS SELECT
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->DYNAMIC_FORM_FIELD_VALUE_ID AS rowkey,
   AS_VALUE(PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->DYNAMIC_FORM_FIELD_VALUE_ID) AS DYNAMIC_FORM_FIELD_VALUE_ID,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->FIELD_ID AS FIELD_ID,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->LESSON_REPORT_DETAIL_ID AS LESSON_REPORT_DETAIL_ID,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->VALUE_TYPE AS VALUE_TYPE,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->STRING_VALUE AS STRING_VALUE,
   PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->INT_VALUE AS INT_VALUE
   FROM PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1
   WHERE PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_PARTNER_DYNAMIC_FORM_FIELD_VALUES_PUBLIC_INFO_V1 WITH (
    'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
    'transforms.unwrap.delete.handling.mode'='drop',
    'tasks.max'='1',
    'topics'='{{ .Values.topicPrefix }}PARTNER_DYNAMIC_FORM_FIELD_VALUES_STREAM_FORMATED_V1',
    'fields.whitelist'='dynamic_form_field_value_id,field_id,lesson_report_detail_id,created_at,updated_at,deleted_at,value_type,string_value,int_value',
    'key.converter'='org.apache.kafka.connect.storage.StringConverter',
    'value.converter'='io.confluent.connect.avro.AvroConverter',
    'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
    'delete.enabled'='false',
    'transforms.unwrap.drop.tombstones'='true',
    'auto.create'='true',
    'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
    'insert.mode'='upsert',
    'table.name.format'='public.partner_dynamic_form_field_values',
    'pk.mode'='record_value',
    'transforms'='RenameField',
    'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
    'transforms.RenameField.renames'='DYNAMIC_FORM_FIELD_VALUE_ID:dynamic_form_field_value_id,FIELD_ID:field_id,LESSON_REPORT_DETAIL_ID:lesson_report_detail_id,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,VALUE_TYPE:value_type,STRING_VALUE:string_value,INT_VALUE:int_value',
    'pk.fields'='dynamic_form_field_value_id'
);
