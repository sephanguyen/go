set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS DAY_TYPE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.day_type', value_format='AVRO');
CREATE STREAM IF NOT EXISTS DAY_TYPE_STREAM_FORMATED_V1 AS 
    SELECT
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID AS rowkey,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DISPLAY_NAME AS DISPLAY_NAME,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        AS_VALUE(DAY_TYPE_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID) AS DAY_TYPE_ID
    FROM DAY_TYPE_STREAM_ORIGIN_V1
    WHERE DAY_TYPE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_DAY_TYPE_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}DAY_TYPE_STREAM_FORMATED_V1',
      'fields.whitelist'='day_type_id,display_name,is_archived,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.day_type_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='DAY_TYPE_ID:day_type_id,DISPLAY_NAME:display_name,IS_ARCHIVED:is_archived,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='day_type_id'
);
