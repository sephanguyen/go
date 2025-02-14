SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS PERMISSION_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.permission', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PERMISSION_STREAM_FORMATED_V1
    AS SELECT
        PERMISSION_STREAM_ORIGIN_V1.AFTER->PERMISSION_ID AS PERMISSION_ID,
        PERMISSION_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        PERMISSION_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        PERMISSION_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT,
        PERMISSION_STREAM_ORIGIN_V1.AFTER->PERMISSION_NAME AS PERMISSION_NAME

    FROM PERMISSION_STREAM_ORIGIN_V1
    WHERE PERMISSION_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS PERMISSION_ROLE_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.permission_role', value_format='AVRO');
CREATE STREAM IF NOT EXISTS PERMISSION_ROLE_STREAM_FORMATED_V1
    AS SELECT
        PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->PERMISSION_ID AS PERMISSION_ID,
        PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->ROLE_ID AS ROLE_ID,
        PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM PERMISSION_ROLE_STREAM_ORIGIN_V1
    WHERE PERMISSION_ROLE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS GRANTED_PERMISSION_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.granted_permission', value_format='AVRO');
CREATE STREAM IF NOT EXISTS GRANTED_PERMISSION_STREAM_FORMATED_V1
    AS SELECT
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->USER_GROUP_ID AS USER_GROUP_ID,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->ROLE_ID AS ROLE_ID,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->PERMISSION_ID AS PERMISSION_ID,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->USER_GROUP_NAME AS USER_GROUP_NAME,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->ROLE_NAME AS ROLE_NAME,
        GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->PERMISSION_NAME AS PERMISSION_NAME

    FROM GRANTED_PERMISSION_STREAM_ORIGIN_V1
    WHERE GRANTED_PERMISSION_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS PERMISSION_PUBLIC_INFO_V1 
AS SELECT
    GRANTED_PERMISSION_STREAM_FORMATED_V1.USER_GROUP_ID AS USER_GROUP_ID,
    GRANTED_PERMISSION_STREAM_FORMATED_V1.ROLE_ID AS GRANTED_PERMISSION_ROLE_ID,
    GRANTED_PERMISSION_STREAM_FORMATED_V1.LOCATION_ID AS LOCATION_ID,
    GRANTED_PERMISSION_STREAM_FORMATED_V1.USER_GROUP_NAME AS USER_GROUP_NAME,
    GRANTED_PERMISSION_STREAM_FORMATED_V1.ROLE_NAME AS ROLE_NAME,
    GRANTED_PERMISSION_STREAM_FORMATED_V1.PERMISSION_NAME AS GRANTED_PERMISSION_PERMISSION_NAME,

    PERMISSION_STREAM_FORMATED_V1.PERMISSION_ID AS PERMISSION_ID,
    PERMISSION_STREAM_FORMATED_V1.CREATED_AT AS PERMISSION_CREATED_AT,
    PERMISSION_STREAM_FORMATED_V1.UPDATED_AT AS PERMISSION_UPDATED_AT,
    PERMISSION_STREAM_FORMATED_V1.DELETED_AT AS PERMISSION_DELETED_AT,
    PERMISSION_STREAM_FORMATED_V1.PERMISSION_NAME AS PERMISSION_PERMISSION_NAME,

    PERMISSION_ROLE_STREAM_FORMATED_V1.ROLE_ID AS PERMISSION_ROLE_ROLE_ID,
    PERMISSION_ROLE_STREAM_FORMATED_V1.CREATED_AT AS PERMISSION_ROLE_CREATED_AT,
    PERMISSION_ROLE_STREAM_FORMATED_V1.UPDATED_AT AS PERMISSION_ROLE_UPDATED_AT,
    PERMISSION_ROLE_STREAM_FORMATED_V1.DELETED_AT AS PERMISSION_ROLE_DELETED_AT

FROM PERMISSION_STREAM_FORMATED_V1 
JOIN PERMISSION_ROLE_STREAM_FORMATED_V1 WITHIN 2 HOURS ON PERMISSION_STREAM_FORMATED_V1.PERMISSION_ID = PERMISSION_ROLE_STREAM_FORMATED_V1.PERMISSION_ID
JOIN GRANTED_PERMISSION_STREAM_FORMATED_V1 WITHIN 2 HOURS ON PERMISSION_STREAM_FORMATED_V1.PERMISSION_ID = GRANTED_PERMISSION_STREAM_FORMATED_V1.PERMISSION_ID;

CREATE SINK CONNECTOR IF NOT EXISTS PERMISSION_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}PERMISSION_PUBLIC_INFO_V1',
      'fields.whitelist'='user_group_id,granted_permission_role_id,location_id,user_group_name,role_name,granted_permission_permission_name,permission_id,permission_created_at,permission_updated_at,permission_deleted_at,permission_permission_name,permission_role_role_id,permission_role_created_at,permission_role_updated_at,permission_role_deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.permission_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_GROUP_ID:user_group_id,GRANTED_PERMISSION_ROLE_ID:granted_permission_role_id,LOCATION_ID:location_id,USER_GROUP_NAME:user_group_name,ROLE_NAME:role_name,GRANTED_PERMISSION_PERMISSION_NAME:granted_permission_permission_name,PERMISSION_ID:permission_id,PERMISSION_CREATED_AT:permission_created_at,PERMISSION_UPDATED_AT:permission_updated_at,PERMISSION_DELETED_AT:permission_deleted_at,PERMISSION_PERMISSION_NAME:permission_permission_name,PERMISSION_ROLE_ROLE_ID:permission_role_role_id,PERMISSION_ROLE_CREATED_AT:permission_role_created_at,PERMISSION_ROLE_UPDATED_AT:permission_role_updated_at,PERMISSION_ROLE_DELETED_AT:permission_role_deleted_at',
      'pk.fields'='permission_id'
);