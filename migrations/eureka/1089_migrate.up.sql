CREATE TABLE IF NOT EXISTS public.task_assignment (
    -- Inherited from table public.learning_material: learning_material_id text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: topic_id text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: name text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: type text COLLATE pg_catalog."default",
    -- Inherited from table public.learning_material: display_order smallint,
    -- Inherited from table public.learning_material: created_at timestamp with time zone NOT NULL,
    -- Inherited from table public.learning_material: updated_at timestamp with time zone NOT NULL,
    -- Inherited from table public.learning_material: deleted_at timestamp with time zone,
    -- Inherited from table public.learning_material: resource_path text COLLATE pg_catalog."default",
    attachments text[],
    instruction text, 
    -- setting fields
    require_duration bool,
    require_complete_date bool,
    require_understanding_level bool,
    require_correctness bool,
    require_attachment bool,
    require_assignment_note bool,
    CONSTRAINT task_assignment_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT task_assignment_type_check CHECK (type = 'LEARNING_MATERIAL_TASK_ASSIGNMENT' :: text),
    CONSTRAINT task_assignment_topic_id_fk FOREIGN KEY (topic_id) REFERENCES topics(topic_id)
) INHERITS (public.learning_material);

/* set RLS */
CREATE POLICY rls_task_assignment ON "task_assignment" using (
    permission_check(resource_path, 'task_assignment')
) with check (
    permission_check(resource_path, 'task_assignment')
);

ALTER TABLE "task_assignment" ENABLE ROW LEVEL security;
ALTER TABLE "task_assignment" FORCE ROW LEVEL security;
