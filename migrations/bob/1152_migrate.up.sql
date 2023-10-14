DROP INDEX IF EXISTS shuffled_quiz_sets_study_plan_item_idx;
CREATE INDEX shuffled_quiz_sets_study_plan_item_idx ON public.shuffled_quiz_sets(study_plan_item_id);

DROP INDEX IF EXISTS student_event_logs_payload_session_id_idx;
CREATE INDEX student_event_logs_payload_session_id_idx ON student_event_logs((payload->>'session_id'));

DROP INDEX IF EXISTS student_event_logs_event_type_idx;
CREATE INDEX student_event_logs_event_type_idx ON public.student_event_logs(event_type);
