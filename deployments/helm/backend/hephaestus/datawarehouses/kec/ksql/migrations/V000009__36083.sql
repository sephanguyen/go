set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS LESSON_PARTNER_FORM_CONFIGS_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.partner_form_configs', value_format='AVRO');
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
      'table.name.format'='bob.partner_form_configs_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='FORM_CONFIG_ID:form_config_id,PARTNER_ID:partner_id,FEATURE_NAME:feature_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,FORM_CONFIG_DATA:form_config_data',
      'pk.fields'='form_config_id'
);
