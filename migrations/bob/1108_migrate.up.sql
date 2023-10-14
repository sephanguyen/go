CREATE TABLE IF NOT EXISTS "lesson_report_approval_records" (
    "record_id" TEXT NOT NULL PRIMARY KEY,
    "lesson_report_id" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "approved_by" TEXT NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "deleted_at" timestamp with time zone,
    "resource_path" TEXT,
    CONSTRAINT lesson_reports_fk FOREIGN KEY (lesson_report_id) REFERENCES public.lesson_reports(lesson_report_id),
    CONSTRAINT users_fk FOREIGN KEY (approved_by) REFERENCES public.users(user_id)
);