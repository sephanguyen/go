ALTER TABLE IF EXISTS public.lessons
    ADD COLUMN IF NOT EXISTS "name" text,
    ADD COLUMN IF NOT EXISTS "start_time" timestamp with time zone,
    ADD COLUMN IF NOT EXISTS "end_time" timestamp with time zone;

CREATE TABLE IF NOT EXISTS "lessons_courses" (
    "lesson_id" text NOT NULL,
    "course_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    CONSTRAINT lessons_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT courses_fk FOREIGN KEY (course_id) REFERENCES public.courses(course_id),
    CONSTRAINT lessons_courses_pk PRIMARY KEY (lesson_id, course_id)
);

CREATE TABLE IF NOT EXISTS "lessons_teachers" (
    "lesson_id" text NOT NULL,
    "teacher_id" text NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    CONSTRAINT lessons_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT teachers_fk FOREIGN KEY (teacher_id) REFERENCES public.teachers(teacher_id),
    CONSTRAINT lessons_teachers_pk PRIMARY KEY (lesson_id, teacher_id)
);
