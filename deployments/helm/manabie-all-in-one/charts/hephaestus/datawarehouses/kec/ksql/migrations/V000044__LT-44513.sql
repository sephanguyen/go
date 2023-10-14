SET 'auto.offset.reset' = 'earliest';

/* grade */

CREATE STREAM IF NOT EXISTS MASTERMGMT_GRADE_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.mastermgmt.grade', value_format='AVRO');

CREATE STREAM IF NOT EXISTS MASTERMGMT_GRADE_STREAM_FORMATED_V1 with (kafka_topic='{{ .Values.topicPrefix }}MASTERMGMT_GRADE_STREAM_FORMATED_V1', value_format='AVRO')
    AS SELECT
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID AS KEY,
        AS_VALUE(MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID) AS GRADE_ID,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->NAME AS NAME,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->IS_ARCHIVED AS IS_ARCHIVED,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->PARTNER_INTERNAL_ID AS PARTNER_INTERNAL_ID,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->SEQUENCE AS SEQUENCE,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS CREATED_AT,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS UPDATED_AT,
        MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->DELETED_AT AS DELETED_AT
    FROM MASTERMGMT_GRADE_STREAM_ORIGIN_V1
    WHERE MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY MASTERMGMT_GRADE_STREAM_ORIGIN_V1.AFTER->GRADE_ID
    EMIT CHANGES;

DROP CONNECTOR IF EXISTS SINK_GRADE_TABLE_FORMATED_V1;
CREATE SINK CONNECTOR IF NOT EXISTS SINK_GRADE_TABLE_FORMATED_V1 WITH (
      'connector.class'='io.confluent.connect.jdbc.JdbcSinkConnector',
      'transforms.unwrap.delete.handling.mode'='drop',
      'tasks.max'='1',
      'topics'='{{ .Values.topicPrefix }}MASTERMGMT_GRADE_STREAM_FORMATED_V1',
      'fields.whitelist'='grade_id,name,is_archived,grade_created_at,grade_updated_at,grade_deleted_at,partner_internal_id,sequence',
      'key.converter'='org.apache.kafka.connect.storage.StringConverter',
      'value.converter'='io.confluent.connect.avro.AvroConverter',
      'value.converter.schema.registry.url'='{{ .Values.cpRegistryHost }}',
      'delete.enabled'='false',                     
      'transforms.unwrap.drop.tombstones'='true',
      'auto.create'='true',                                                                                                                             
      'connection.url'='${file:/decrypted/kafka-connect.secrets.properties:kec_url}',
      'insert.mode'='upsert',
      'table.name.format'='public.grade',
      'pk.mode'='record_value',
      'transforms'='RenameField',
      'transforms.RenameField.type'= 'org.apache.kafka.connect.transforms.ReplaceField$Value',
      'transforms.RenameField.renames'='GRADE_ID:grade_id,NAME:name,IS_ARCHIVED:is_archived,CREATED_AT:grade_created_at,UPDATED_AT:grade_updated_at,DELETED_AT:grade_deleted_at,PARTNER_INTERNAL_ID:partner_internal_id,SEQUENCE:sequence',
      'pk.fields'='grade_id'
);

