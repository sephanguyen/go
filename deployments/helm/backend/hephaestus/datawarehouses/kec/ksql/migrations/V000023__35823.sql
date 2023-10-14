set 'auto.offset.reset' = 'earliest';

-- ORIGINAL

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_OTHER_WORKING_HOURS_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.other_working_hours', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TRANSPORTATION_EXPENSE_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.transportation_expense', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_STAFF_TRANSPORTATION_EXPENSE_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.staff_transportation_expense', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_LESSONS_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.lessons', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_LESSON_HOURS_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet_lesson_hours', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_AUTO_CREATE_FLAG_ACTIVITY_LOG_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.auto_create_flag_activity_log', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_AUTO_CREATE_TIMESHEET_FLAG_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.auto_create_timesheet_flag', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIRMATION_CUT_OFF_DATE_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet_confirmation_cut_off_date', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIRMATION_INFO_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet_confirmation_info', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIRMATION_PERIOD_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet_confirmation_period', value_format='AVRO');

CREATE STREAM IF NOT EXISTS TIMESHEET_TIMESHEET_CONFIG_ORIGINAL
WITH (kafka_topic='{{ .Values.global.environment }}.kec.datalake.timesheet.timesheet_config', value_format='AVRO');
