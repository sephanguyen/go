alter table public.withus_mapping_course_id drop column if exists created_date;
alter table public.withus_mapping_course_id drop column if exists created_by;

alter table public.withus_mapping_exam_lo_id drop column if exists created_date;
alter table public.withus_mapping_exam_lo_id drop column if exists created_by;

alter table public.withus_mapping_question_tag drop column if exists created_date;
alter table public.withus_mapping_question_tag drop column if exists created_by;

alter table public.withus_failed_sync_email_recipient drop column if exists created_date;
alter table public.withus_failed_sync_email_recipient drop column if exists created_by;
