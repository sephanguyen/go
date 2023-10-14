CREATE TABLE IF NOT EXISTS "lesson_reports" (
    "lesson_report_id" TEXT NOT NULL,
    "start_time" timestamp with time zone,
    "end_time" timestamp with time zone,
    "lesson_type" TEXT,
    "teacher_ids" TEXT[],
    "student_ids" TEXT[],
    "report_submitting_status" TEXT NOT NULL,
    "school_id" INT,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "resource_path" TEXT,
    CONSTRAINT schools_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id),
    CONSTRAINT lesson_reports_pk PRIMARY KEY (lesson_report_id)
);

CREATE TABLE IF NOT EXISTS "lesson_report_details" (
    "lesson_report_id" TEXT NOT NULL,
    "student_id" TEXT NOT NULL,
    "course_id" TEXT NOT NULL,
    "attendance_status" TEXT,
    "attendance_remark" TEXT,
    "homework_status" TEXT,
    "homework_completion" INT,
    "homework_score" INT,
    "extra_report" JSONB,
    "remark" TEXT,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "resource_path" TEXT,
    CONSTRAINT lesson_reports_fk FOREIGN KEY (lesson_report_id) REFERENCES public.lesson_reports(lesson_report_id),
    CONSTRAINT students_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id),
    CONSTRAINT courses_fk FOREIGN KEY (course_id) REFERENCES public.courses(course_id),
    CONSTRAINT lesson_report_details_pk PRIMARY KEY (lesson_report_id, student_id)
);