CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_FORMATTED
AS SELECT
    AFTER->ID AS KEY,
    AS_VALUE(AFTER->ID) AS ID,
    AFTER->START_DATE AS START_DATE,
    AFTER->END_DATE AS END_DATE,
    AFTER->CREATED_AT AS CREATED_AT,
    AFTER->UPDATED_AT AS UPDATED_AT,
    AFTER->DELETED_AT AS DELETED_AT
FROM TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_ORIGINAL
WHERE TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_ORIGINAL.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY AFTER-> ID
EMIT CHANGES;

CREATE TABLE IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_TABLE (KEY VARCHAR(STRING) PRIMARY KEY)
WITH(kafka_topic='{{ .Values.topicPrefix }}TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_FORMATTED', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIG_FORMATTED
AS SELECT
    AFTER->TIMESHEET_CONFIG_ID AS KEY,
    AS_VALUE(AFTER->TIMESHEET_CONFIG_ID) AS TIMESHEET_CONFIG_ID,
    AFTER->CONFIG_TYPE AS CONFIG_TYPE,
    AFTER->CONFIG_VALUE AS CONFIG_VALUE,
    AFTER->IS_ARCHIVED AS IS_ARCHIVED,
    AFTER->CREATED_AT AS CREATED_AT,
    AFTER->UPDATED_AT AS UPDATED_AT,
    AFTER->DELETED_AT AS DELETED_AT
FROM TIMESHEET_TIMESHEET_CONFIG_ORIGINAL
WHERE TIMESHEET_TIMESHEET_CONFIG_ORIGINAL.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
PARTITION BY AFTER->TIMESHEET_CONFIG_ID
EMIT CHANGES;

CREATE TABLE IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIG_TABLE (KEY VARCHAR(STRING) PRIMARY KEY)
WITH(kafka_topic='{{ .Values.topicPrefix }}TIMESHEET_TIMESHEET_CONFIG_FORMATTED', value_format='AVRO');
