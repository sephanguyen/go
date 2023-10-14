CREATE TABLE IF NOT EXISTS public.allocate_marker (
    allocate_marker_id text NOT NULL,
    teacher_id text NOT NULL,
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    created_by text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT pk_allocate_marker PRIMARY KEY (student_id,study_plan_id,learning_material_id)
);

CREATE POLICY rls_allocate_marker ON "allocate_marker" using (
    permission_check(resource_path, 'allocate_marker')
) with check (
    permission_check(resource_path, 'allocate_marker')
);