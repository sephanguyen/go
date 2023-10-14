CREATE TABLE IF NOT EXISTS public.import_study_plan_task (
  task_id text NOT NULL,
  study_plan_id text NOT NULL,
  status text NOT NULL,
  error_detail text,
  imported_by text NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  created_at timestamp with time zone NOT NULL,
  resource_path text DEFAULT autofillresourcepath(),
  CONSTRAINT import_study_plan_task_pk PRIMARY KEY (task_id)
);

/* set RLS */
CREATE POLICY rls_import_study_plan_task ON "import_study_plan_task" using (
    permission_check(resource_path, 'import_study_plan_task')
) with check (
    permission_check(resource_path, 'import_study_plan_task')
);

ALTER TABLE
    "import_study_plan_task" ENABLE ROW LEVEL security;

ALTER TABLE
    "import_study_plan_task" FORCE ROW LEVEL security;