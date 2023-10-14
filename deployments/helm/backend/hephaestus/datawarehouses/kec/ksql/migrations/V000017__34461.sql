SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS USER_TAG_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.user_tag', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USER_TAG_STREAM_FORMATED_V1
    AS SELECT
        USER_TAG_STREAM_ORIGIN_V1.AFTER->USER_TAG_ID AS USER_TAG_ID,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->USER_TAG_NAME AS USER_TAG_NAME,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->USER_TAG_PARTNER_ID AS USER_TAG_PARTNER_ID,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->USER_TAG_TYPE AS USER_TAG_TYPE,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        USER_TAG_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM USER_TAG_STREAM_ORIGIN_V1
    WHERE USER_TAG_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS TAGGED_USER_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.tagged_user', value_format='AVRO');
CREATE STREAM IF NOT EXISTS TAGGED_USER_STREAM_FORMATED_V1
    AS SELECT
        TAGGED_USER_STREAM_ORIGIN_V1.AFTER->TAG_ID AS TAG_ID,
        TAGGED_USER_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        TAGGED_USER_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        TAGGED_USER_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        TAGGED_USER_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM TAGGED_USER_STREAM_ORIGIN_V1
    WHERE TAGGED_USER_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS TAGGED_USER_PUBLIC_INFO_V1 
AS SELECT
    TAGGED_USER_STREAM_FORMATED_V1.TAG_ID AS KEY,
    AS_VALUE(TAGGED_USER_STREAM_FORMATED_V1.TAG_ID) AS TAG_ID,
    TAGGED_USER_STREAM_FORMATED_V1.USER_ID AS USER_ID,
    TAGGED_USER_STREAM_FORMATED_V1.CREATED_AT AS TAGGED_USER_CREATED_AT,
    TAGGED_USER_STREAM_FORMATED_V1.UPDATED_AT AS TAGGED_USER_UPDATED_AT,
    TAGGED_USER_STREAM_FORMATED_V1.DELETED_AT AS TAGGED_USER_DELETED_AT,

    USER_TAG_STREAM_FORMATED_V1.USER_TAG_NAME AS NAME,
    USER_TAG_STREAM_FORMATED_V1.USER_TAG_PARTNER_ID AS USER_TAG_PARTNER_ID,
    USER_TAG_STREAM_FORMATED_V1.USER_TAG_TYPE AS USER_TAG_TYPE,
    USER_TAG_STREAM_FORMATED_V1.IS_ARCHIVED AS IS_ARCHIVED,
    USER_TAG_STREAM_FORMATED_V1.CREATED_AT AS USER_TAG_CREATED_AT,
    USER_TAG_STREAM_FORMATED_V1.UPDATED_AT AS USER_TAG_UPDATED_AT,
    USER_TAG_STREAM_FORMATED_V1.DELETED_AT AS USER_TAG_DELETED_AT

FROM TAGGED_USER_STREAM_FORMATED_V1
JOIN USER_TAG_STREAM_FORMATED_V1 WITHIN 2 HOURS ON TAGGED_USER_STREAM_FORMATED_V1.TAG_ID = USER_TAG_STREAM_FORMATED_V1.USER_TAG_ID;

CREATE SINK CONNECTOR IF NOT EXISTS TAGGED_USER_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}TAGGED_USER_PUBLIC_INFO_V1',
      'fields.whitelist'='tag_id,user_id,tagged_user_created_at,tagged_user_updated_at,tagged_user_deleted_at,user_tag_partner_id,name,is_archived,user_tag_created_at,user_tag_updated_at,user_tag_deleted_at,user_tag_type',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.tagged_user_public_info',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='TAG_ID:tag_id,USER_ID:user_id,TAGGED_USER_CREATED_AT:tagged_user_created_at,TAGGED_USER_UPDATED_AT:tagged_user_updated_at,TAGGED_USER_DELETED_AT:tagged_user_deleted_at,USER_TAG_PARTNER_ID:user_tag_partner_id,NAME:name,IS_ARCHIVED:is_archived,USER_TAG_CREATED_AT:user_tag_created_at,USER_TAG_UPDATED_AT:user_tag_updated_at,USER_TAG_DELETED_AT:user_tag_deleted_at,USER_TAG_TYPE:user_tag_type',
      'pk.fields'='user_id,tag_id'
);

CREATE STREAM IF NOT EXISTS USER_PHONE_NUMBER_STREAM_ORIGIN_V1 WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.bob.user_phone_number', value_format='AVRO');
CREATE STREAM IF NOT EXISTS USER_PHONE_NUMBER_STREAM_FORMATED_V1
    AS SELECT
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->USER_PHONE_NUMBER_ID AS USER_PHONE_NUMBER_ID,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->USER_ID AS USER_ID,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->PHONE_NUMBER AS PHONE_NUMBER,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->TYPE AS TYPE,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT

    FROM USER_PHONE_NUMBER_STREAM_ORIGIN_V1
    WHERE USER_PHONE_NUMBER_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}';

CREATE STREAM IF NOT EXISTS USER_PHONE_NUMBER_PUBLIC_INFO_V1 
AS SELECT
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.USER_PHONE_NUMBER_ID AS USER_PHONE_NUMBER_ID,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.USER_ID AS USER_ID,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.PHONE_NUMBER AS PHONE_NUMBER,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.TYPE AS TYPE,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.CREATED_AT AS CREATED_AT,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.UPDATED_AT AS UPDATED_AT,
    USER_PHONE_NUMBER_STREAM_FORMATED_V1.DELETED_AT AS DELETED_AT

FROM USER_PHONE_NUMBER_STREAM_FORMATED_V1
PARTITION BY USER_PHONE_NUMBER_STREAM_FORMATED_V1.USER_PHONE_NUMBER_ID;

CREATE SINK CONNECTOR IF NOT EXISTS USER_PHONE_NUMBER_STREAM_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}USER_PHONE_NUMBER_PUBLIC_INFO_V1',
      'fields.whitelist'='user_id,user_phone_number_id,phone_number,type,updated_at,created_at,deleted_at',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='bob.user_phone_number_public_info',
      'pk.mode'='record_key',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='USER_ID:user_id,USER_PHONE_NUMBER_ID:user_phone_number_id,PHONE_NUMBER:phone_number,TYPE:type,CREATED_AT:created_at,UPDATED_AT:updated_at,DELETED_AT:deleted_at',
      'pk.fields'='user_phone_number_id'
);