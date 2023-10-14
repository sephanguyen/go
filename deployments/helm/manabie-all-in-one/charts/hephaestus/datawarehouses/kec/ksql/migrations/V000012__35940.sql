set 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS DAY_INFO_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.day_info', value_format='AVRO');
CREATE STREAM IF NOT EXISTS DAY_INFO_STREAM_FORMATED_V1 AS 
    SELECT
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DATE AS DATE,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DAY_TYPE_ID AS DAY_TYPE_ID,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->OPENING_TIME AS OPENING_TIME,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->STATUS AS STATUS,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->TIME_ZONE AS TIME_ZONE,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        DAY_INFO_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM DAY_INFO_STREAM_ORIGIN_V1
    WHERE DAY_INFO_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE SINK CONNECTOR IF NOT EXISTS SINK_DAY_INFO_PUBLIC_INFO_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}DAY_INFO_STREAM_FORMATED_V1',
      'fields.whitelist'='location_id,date,day_type_id,opening_time,status,time_zone,created_at,updated_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.day_info_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='LOCATION_ID:location_id,DATE:date,DAY_TYPE_ID:day_type_id,OPENING_TIME:opening_time,STATUS:status,TIME_ZONE:time_zone,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='location_id,date'
);
