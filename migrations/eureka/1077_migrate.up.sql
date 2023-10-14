CREATE TABLE IF NOT EXISTS public.learning_material (
    learning_material_id text NOT NULL,
    topic_id text NOT NULL,
    name text NOT NULL,
    type text,
    display_order smallint,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path TEXT,
    CONSTRAINT learning_material_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id)
);

/* set RLS */
CREATE POLICY rls_learning_material ON "learning_material" using (
    permission_check(resource_path, 'learning_material')
) with check (
    permission_check(resource_path, 'learning_material')
);

ALTER TABLE
    "learning_material" ENABLE ROW LEVEL security;

ALTER TABLE
    "learning_material" FORCE ROW LEVEL security;