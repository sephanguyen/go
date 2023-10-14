alter table courses
add column if not exists deleted_at timestamp with time zone;

alter table topics
add column if not exists deleted_at timestamp with time zone;

alter table chapters
add column if not exists deleted_at timestamp with time zone;

alter table learning_objectives
add column if not exists deleted_at timestamp with time zone;
