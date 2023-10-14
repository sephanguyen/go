CREATE TABLE IF NOT EXISTS public.lesson_groups (
    lesson_group_id TEXT NOT NULL,
    course_id TEXT NOT NULL,
    media_ids TEXT[],
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__lesson_groups PRIMARY KEY (lesson_group_id, course_id)
);

ALTER TABLE public.lessons
  ADD COLUMN IF NOT EXISTS lesson_group_id TEXT;