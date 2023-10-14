CREATE TABLE IF NOT EXISTS school_history(
  student_id TEXT NOT NULL,
  school_id TEXT NOT NULL,
  school_course_id TEXT,
  start_date timestamp with time zone,
  end_date timestamp with time zone,

  created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
  deleted_at timestamp with time zone,
  resource_path TEXT DEFAULT autofillresourcepath(),

  CONSTRAINT school_history__pk PRIMARY KEY (student_id, school_id),
  CONSTRAINT school_history__school_id__fk FOREIGN KEY (school_id) REFERENCES public.school_info(school_id),
  CONSTRAINT school_history__student_id__fk FOREIGN KEY (student_id) REFERENCES public.students(student_id)
);

CREATE POLICY rls_school_history ON "school_history"
USING (permission_check(resource_path, 'school_history'))
WITH CHECK (permission_check(resource_path, 'school_history'));

ALTER TABLE "school_history" ENABLE ROW LEVEL security;
ALTER TABLE "school_history" FORCE ROW LEVEL security;
