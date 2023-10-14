DROP INDEX IF EXISTS student_event_logs_payload_study_plan_item_id_idx;

CREATE INDEX student_event_logs_payload_study_plan_item_id_idx
ON student_event_logs((payload->>'study_plan_item_id'))
WHERE payload->>'study_plan_item_id' IS NOT NULL;
