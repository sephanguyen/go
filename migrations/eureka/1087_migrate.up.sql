CREATE TABLE IF NOT EXISTS public.assignment (
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
    max_grade integer,
    "status" text,
    instruction text, 
    is_required_grade bool,
    -- setting fields
    allow_resubmission bool,
    require_attachment bool,
    allow_late_submission bool, 
    require_assignment_note bool,
    require_video_submission bool,
    CONSTRAINT assignment_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT assignment_type_check CHECK (type = 'LEARNING_MATERIAL_GENERAL_ASSIGNMENT' :: text),
    CONSTRAINT assignment_topic_id_fk FOREIGN KEY (topic_id) REFERENCES topics(topic_id)
) INHERITS (public.learning_material);

/* set RLS */
CREATE POLICY rls_assignment ON "assignment" using (
    permission_check(resource_path, 'assignment')
) with check (
    permission_check(resource_path, 'assignment')
);

ALTER TABLE "assignment" ENABLE ROW LEVEL security;
ALTER TABLE "assignment" FORCE ROW LEVEL security;