SET 'auto.offset.reset' = 'earliest';
CREATE STREAM IF NOT EXISTS USER_GROUP_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.user_group', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USER_GROUP_STREAM_FORMATED_V1
    AS SELECT
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->USER_GROUP_ID AS USER_GROUP_ID,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->USER_GROUP_NAME AS USER_GROUP_NAME,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH AS RESOURCE_PATH,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->ORG_LOCATION_ID AS ORG_LOCATION_ID,
        USER_GROUP_STREAM_ORIGIN_V1.AFTER->IS_SYSTEM AS IS_SYSTEM
    FROM USER_GROUP_STREAM_ORIGIN_V1
    WHERE USER_GROUP_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS USER_GROUP_PUBLIC_INFO_V1 
AS SELECT
    USER_GROUP_STREAM_FORMATED_V1.USER_GROUP_ID AS USER_GROUP_ID,
    USER_GROUP_STREAM_FORMATED_V1.USER_GROUP_NAME AS USER_GROUP_NAME,
    USER_GROUP_STREAM_FORMATED_V1.CREATED_AT AS CREATED_AT,
    USER_GROUP_STREAM_FORMATED_V1.UPDATED_AT AS UPDATED_AT,
    USER_GROUP_STREAM_FORMATED_V1.DELETED_AT AS DELETED_AT,
    USER_GROUP_STREAM_FORMATED_V1.RESOURCE_PATH AS RESOURCE_PATH,
    USER_GROUP_STREAM_FORMATED_V1.ORG_LOCATION_ID AS ORG_LOCATION_ID,
    USER_GROUP_STREAM_FORMATED_V1.IS_SYSTEM AS IS_SYSTEM

FROM USER_GROUP_STREAM_FORMATED_V1
PARTITION BY USER_GROUP_STREAM_FORMATED_V1.USER_GROUP_ID;

CREATE SINK CONNECTOR IF NOT EXISTS USER_GROUP_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_GROUP_PUBLIC_INFO_V1',
      'fields.whitelist'='user_group_id,user_group_name,created_at,updated_at,deleted_at,resource_path,org_location_id,is_system',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_group_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_GROUP_ID:user_group_id,USER_GROUP_NAME:user_group_name,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at,RESOURCE_PATH:resource_path,ORG_LOCATION_ID:org_location_id,IS_SYSTEM:is_system',
      'pk.fields'='user_group_id'
);