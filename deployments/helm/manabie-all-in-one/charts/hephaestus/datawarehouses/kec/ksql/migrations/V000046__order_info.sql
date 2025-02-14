SET 'auto.offset.reset' = 'earliest';

CREATE STREAM IF NOT EXISTS ORDER_STREAM_ORIGIN_V1  WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.fatima.order', value_format='AVRO');

CREATE STREAM IF NOT EXISTS ORDER_STREAM_FORMATTED_V1
    AS SELECT
        ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_ID AS KEY,
        AS_VALUE(ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_ID) AS ORDER_ID,
        ORDER_STREAM_ORIGIN_V1.AFTER->STUDENT_ID AS STUDENT_ID,
        ORDER_STREAM_ORIGIN_V1.AFTER->LOCATION_ID AS LOCATION_ID,
        ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_SEQUENCE_NUMBER AS ORDER_SEQUENCE_NUMBER,
        ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_COMMENT AS ORDER_COMMENT,
        ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_STATUS AS ORDER_STATUS,
        ORDER_STREAM_ORIGIN_V1.AFTER->ORDER_TYPE AS ORDER_TYPE,
        ORDER_STREAM_ORIGIN_V1.AFTER->STUDENT_FULL_NAME AS STUDENT_FULL_NAME,
        ORDER_STREAM_ORIGIN_V1.AFTER->IS_REVIEWED AS IS_REVIEWED,
        ORDER_STREAM_ORIGIN_V1.AFTER->WITHDRAWAL_EFFECTIVE_DATE AS WITHDRAWAL_EFFECTIVE_DATE,
        ORDER_STREAM_ORIGIN_V1.AFTER->BACKGROUND AS BACKGROUND,
        ORDER_STREAM_ORIGIN_V1.AFTER->FUTURE_MEASURES AS FUTURE_MEASURES,
        ORDER_STREAM_ORIGIN_V1.AFTER->LOA_START_DATE AS LOA_START_DATE,
        ORDER_STREAM_ORIGIN_V1.AFTER->LOA_END_DATE AS LOA_END_DATE,
        ORDER_STREAM_ORIGIN_V1.AFTER->CREATED_AT AS ORDER_CREATED_AT,
        ORDER_STREAM_ORIGIN_V1.AFTER->UPDATED_AT AS ORDER_UPDATED_AT,
        CAST(NULL AS VARCHAR) AS ORDER_DELETED_AT
    FROM ORDER_STREAM_ORIGIN_V1
    WHERE ORDER_STREAM_ORIGIN_V1.AFTER->RESOURCE_PATH = '{{ .Values.kecResourcePath }}'
    PARTITION BY AFTER->ORDER_ID
    EMIT CHANGES;


CREATE TABLE IF NOT EXISTS ORDER_TABLE_FORMATTED_V1 (KEY VARCHAR PRIMARY KEY) with (kafka_topic='{{ .Values.topicPrefix }}ORDER_STREAM_FORMATTED_V1', value_format='AVRO');

CREATE TABLE IF NOT EXISTS ORDER_PUBLIC_INFO_V1
AS SELECT
    BILL_ITEM_PUBLIC_INFO_V1.ROW_KEY AS ROW_KEY,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_ITEM_SEQUENCE_NUMBER AS BILL_ITEM_SEQUENCE_NUMBER,
    BILL_ITEM_PUBLIC_INFO_V1.ORDER_ID AS ORDER_ID,
    BILL_ITEM_PUBLIC_INFO_V1.PRODUCT_ID AS PRODUCT_ID,
    BILL_ITEM_PUBLIC_INFO_V1.PRODUCT_DESCRIPTION AS PRODUCT_DESCRIPTION,
    BILL_ITEM_PUBLIC_INFO_V1.PRODUCT_PRICING AS PRODUCT_PRICING,
    BILL_ITEM_PUBLIC_INFO_V1.DISCOUNT_AMOUNT_TYPE AS DISCOUNT_AMOUNT_TYPE,
    BILL_ITEM_PUBLIC_INFO_V1.DISCOUNT_AMOUNT_VALUE AS DISCOUNT_AMOUNT_VALUE,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_TYPE AS BILL_TYPE,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_STATUS AS BILLING_STATUS,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_DATE AS BILLING_DATE,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_FROM AS BILLING_FROM,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_TO AS BILLING_TO,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_SCHEDULE_PERIOD_ID AS BILLING_SCHEDULE_PERIOD_ID,
    BILL_ITEM_PUBLIC_INFO_V1.DISCOUNT_AMOUNT AS DISCOUNT_AMOUNT,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_AMOUNT AS TAX_AMOUNT,
    BILL_ITEM_PUBLIC_INFO_V1.FINAL_PRICE AS FINAL_PRICE,
    BILL_ITEM_PUBLIC_INFO_V1.STUDENT_ID AS STUDENT_ID,
    BILL_ITEM_PUBLIC_INFO_V1.STUDENT_PRODUCT_ID AS STUDENT_PRODUCT_ID,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_APPROVAL_STATUS AS BILLING_APPROVAL_STATUS,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_ITEM_DESCRIPTION AS BILLING_ITEM_DESCRIPTION,
    BILL_ITEM_PUBLIC_INFO_V1.LOCATION_ID AS LOCATION_ID,
    BILL_ITEM_PUBLIC_INFO_V1.DISCOUNT_ID AS DISCOUNT_ID,
    BILL_ITEM_PUBLIC_INFO_V1.PREVIOUS_BILL_ITEM_SEQUENCE_NUMBER AS PREVIOUS_BILL_ITEM_SEQUENCE_NUMBER,
    BILL_ITEM_PUBLIC_INFO_V1.PREVIOUS_BILL_ITEM_STATUS AS PREVIOUS_BILL_ITEM_STATUS,
    BILL_ITEM_PUBLIC_INFO_V1.ADJUSTMENT_PRICE AS ADJUSTMENT_PRICE,
    BILL_ITEM_PUBLIC_INFO_V1.IS_LATEST_BILL_ITEM AS IS_LATEST_BILL_ITEM,
    BILL_ITEM_PUBLIC_INFO_V1.PRICE AS PRICE,
    BILL_ITEM_PUBLIC_INFO_V1.OLD_PRICE AS OLD_PRICE,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_RATIO_NUMERATOR AS BILLING_RATIO_NUMERATOR,
    BILL_ITEM_PUBLIC_INFO_V1.BILLING_RATIO_DENOMINATOR AS BILLING_RATIO_DENOMINATOR,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_ITEM_IS_REVIEWED AS BILL_ITEM_IS_REVIEWED,
    BILL_ITEM_PUBLIC_INFO_V1.RAW_DISCOUNT_AMOUNT AS RAW_DISCOUNT_AMOUNT,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_ITEM_CREATED_AT AS BILL_ITEM_CREATED_AT,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_ITEM_UPDATED_AT AS BILL_ITEM_UPDATED_AT,
    BILL_ITEM_PUBLIC_INFO_V1.BILL_ITEM_DELETED_AT AS BILL_ITEM_DELETED_AT,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_ID AS TAX_ID,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_CATEGORY AS TAX_CATEGORY,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_PERCENTAGE AS TAX_PERCENTAGE,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_NAME AS TAX_NAME,
    BILL_ITEM_PUBLIC_INFO_V1.DEFAULT_FLAG AS DEFAULT_FLAG,
    BILL_ITEM_PUBLIC_INFO_V1.IS_ARCHIVED AS IS_ARCHIVED,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_CREATED_AT AS TAX_CREATED_AT,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_UPDATED_AT AS TAX_UPDATED_AT,
    BILL_ITEM_PUBLIC_INFO_V1.TAX_DELETED_AT AS TAX_DELETED_AT,
    ORDER_TABLE_FORMATTED_V1.ORDER_SEQUENCE_NUMBER AS ORDER_SEQUENCE_NUMBER,
    ORDER_TABLE_FORMATTED_V1.ORDER_COMMENT AS ORDER_COMMENT,
    ORDER_TABLE_FORMATTED_V1.ORDER_STATUS AS ORDER_STATUS,
    ORDER_TABLE_FORMATTED_V1.ORDER_TYPE AS ORDER_TYPE,
    ORDER_TABLE_FORMATTED_V1.STUDENT_FULL_NAME AS STUDENT_FULL_NAME,
    ORDER_TABLE_FORMATTED_V1.IS_REVIEWED AS ORDER_IS_REVIEWED,
    ORDER_TABLE_FORMATTED_V1.WITHDRAWAL_EFFECTIVE_DATE AS WITHDRAWAL_EFFECTIVE_DATE,
    ORDER_TABLE_FORMATTED_V1.BACKGROUND AS BACKGROUND,
    ORDER_TABLE_FORMATTED_V1.FUTURE_MEASURES AS FUTURE_MEASURES,
    ORDER_TABLE_FORMATTED_V1.LOA_START_DATE AS LOA_START_DATE,
    ORDER_TABLE_FORMATTED_V1.LOA_END_DATE AS LOA_END_DATE,
    ORDER_TABLE_FORMATTED_V1.ORDER_CREATED_AT AS ORDER_CREATED_AT,
    ORDER_TABLE_FORMATTED_V1.ORDER_UPDATED_AT AS ORDER_UPDATED_AT,
    ORDER_TABLE_FORMATTED_V1.ORDER_DELETED_AT AS ORDER_DELETED_AT
FROM BILL_ITEM_PUBLIC_INFO_V1
JOIN ORDER_TABLE_FORMATTED_V1
ON BILL_ITEM_PUBLIC_INFO_V1.ORDER_ID = ORDER_TABLE_FORMATTED_V1.KEY;
