SET 'auto.offset.reset' = 'earliest';

-- tags -- (notification)
-- Create a stream from the origin Kafka topic with AVRO value format
CREATE STREAM IF NOT EXISTS TAGS_STREAM_ORIGIN_V1 WITH (
    kafka_topic = '{{ .Values.global.environment }}.kec.datalake.bob.tags',
    value_format = 'AVRO'
);

-- Create a new stream with transformations from the origin stream
CREATE STREAM IF NOT EXISTS TAGS_STREAM_FORMATED_V1 AS
SELECT
    TAGS_STREAM_ORIGIN_V1.AFTER -> TAG_ID AS TAG_ID,
    TAGS_STREAM_ORIGIN_V1.AFTER -> TAG_NAME AS TAG_NAME,
    TAGS_STREAM_ORIGIN_V1.AFTER -> CREATED_AT AS CREATED_AT,
    TAGS_STREAM_ORIGIN_V1.AFTER -> UPDATED_AT AS UPDATED_AT,
    TAGS_STREAM_ORIGIN_V1.AFTER -> DELETED_AT AS DELETED_AT,
    TAGS_STREAM_ORIGIN_V1.AFTER -> IS_ARCHIVED AS IS_ARCHIVED
FROM
    TAGS_STREAM_ORIGIN_V1
WHERE
    TAGS_STREAM_ORIGIN_V1.AFTER -> RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

-- Create a new sink connector for the formatted stream
CREATE SINK CONNECTOR IF NOT EXISTS SINK_TAGS_TABLE_FORMATED_V1 WITH (
    'connector.class' = 'io.confluent.connect.jdbc.JdbcSinkConnector',
    'transforms.unwrap.delete.handling.mode' = 'drop',
    'tasks.max' = '1',
    'topics' = '{{ .Values.topicPrefix }}TAGS_STREAM_FORMATED_V1',
    'fields.whitelist' = 'tag_id,tag_name,created_at,updated_at,deleted_at,is_archived',
    'key.converter' = 'org.apache.kafka.connect.storage.StringConverter',
    'value.converter' = 'io.confluent.connect.avro.AvroConverter',
    'value.converter.schema.registry.url' = '{{ .Values.cpRegistryHost }}',
    'delete.enabled' = 'false',
    'transforms.unwrap.drop.tombstones' = 'true',
    'auto.create' = 'true',
    'connection.url' = '${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
    'insert.mode' = 'upsert',
    'table.name.format' = 'public.tags',
    'pk.mode' = 'record_value',
    'transforms' = 'RenameField',
    'transforms.RenameField.type' = 'org.apache.kafka.connect.transforms.ReplaceField$Value',
    'transforms.RenameField.renames' = 'TAG_ID:tag_id,TAG_NAME:tag_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,IS_ARCHIVED:is_archived',
    'pk.fields' = 'tag_id'
);
