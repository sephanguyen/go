ALTER TABLE lesson_reports
    ADD COLUMN IF NOT EXISTS "lesson_id" TEXT,
    DROP COLUMN IF EXISTS start_time,
    DROP COLUMN IF EXISTS end_time,
    DROP COLUMN IF EXISTS lesson_type,
    DROP COLUMN IF EXISTS teacher_ids,
    DROP COLUMN IF EXISTS student_ids,
    ADD CONSTRAINT lessons_fk FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id);

ALTER TABLE lesson_report_details
    DROP COLUMN IF EXISTS attendance_status,
    DROP COLUMN IF EXISTS attendance_remark,
    DROP COLUMN IF EXISTS homework_status,
    DROP COLUMN IF EXISTS homework_completion,
    DROP COLUMN IF EXISTS homework_score,
    DROP COLUMN IF EXISTS extra_report,
    DROP COLUMN IF EXISTS remark;

ALTER TABLE lesson_members
    ADD COLUMN IF NOT EXISTS attendance_status TEXT,
    ADD COLUMN IF NOT EXISTS attendance_remark TEXT,
    ADD COLUMN IF NOT EXISTS course_id TEXT,
    ADD CONSTRAINT courses_fk FOREIGN KEY (course_id) REFERENCES public.courses(course_id);

ALTER TABLE lessons
    ADD COLUMN IF NOT EXISTS teaching_model TEXT;
