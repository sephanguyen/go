CREATE TABLE IF NOT EXISTS public.exam_lo (
    -- Inherited from table public.learning_material: learning_material_id text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: topic_id text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: name text COLLATE pg_catalog."default" NOT NULL,
    -- Inherited from table public.learning_material: type text COLLATE pg_catalog."default",
    -- Inherited from table public.learning_material: display_order smallint,
    -- Inherited from table public.learning_material: created_at timestamp with time zone NOT NULL,
    -- Inherited from table public.learning_material: updated_at timestamp with time zone NOT NULL,
    -- Inherited from table public.learning_material: deleted_at timestamp with time zone,
    -- Inherited from table public.learning_material: resource_path text COLLATE pg_catalog."default",
    CONSTRAINT exam_lo_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT exam_lo_topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics (topic_id) MATCH SIMPLE ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT exam_lo_type_check CHECK (type = 'LEARNING_MATERIAL_EXAM_LO' :: text)
) INHERITS (public.learning_material);
/* set RLS */
CREATE POLICY rls_exam_lo ON "exam_lo" using (permission_check(resource_path, 'exam_lo')) with check (permission_check(resource_path, 'exam_lo'));
ALTER TABLE "exam_lo" ENABLE ROW LEVEL security;
ALTER TABLE "exam_lo" FORCE ROW LEVEL security;

/* create Trigger: migrate_to_exam_lo */
CREATE OR REPLACE FUNCTION public.migrate_learning_objectives_to_exam_lo_fn()
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    INSERT INTO exam_lo (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
    VALUES (
        NEW.lo_id,
        NEW.topic_id,
        NEW.name,
        'LEARNING_MATERIAL_EXAM_LO',
        NEW.display_order,
        NEW.created_at,
        NEW.updated_at,
        NEW.deleted_at,
        NEW.resource_path
    )
    ON CONFLICT ON CONSTRAINT exam_lo_pk DO UPDATE SET
        topic_id = NEW.topic_id,
        name = NEW.name,
        display_order = NEW.display_order,
        created_at = NEW.created_at,
        updated_at = NEW.updated_at,
        deleted_at = NEW.deleted_at;
    RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_to_exam_lo ON public.learning_objectives;
CREATE TRIGGER migrate_to_exam_lo
    AFTER INSERT OR UPDATE 
    ON public.learning_objectives
    FOR EACH ROW
    WHEN (new.type = 'LEARNING_OBJECTIVE_TYPE_EXAM_LO'::text)
    EXECUTE FUNCTION public.migrate_learning_objectives_to_exam_lo_fn();

/* migrate old data from learning_objectives */
INSERT INTO
    exam_lo (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
SELECT
    lo_id,
    topic_id,
    name,
    'LEARNING_MATERIAL_EXAM_LO',
    display_order,
    created_at,
    updated_at,
    deleted_at,
    resource_path
FROM public.learning_objectives
WHERE type = 'LEARNING_OBJECTIVE_TYPE_EXAM_LO' :: text
ON CONFLICT ON CONSTRAINT exam_lo_pk DO NOTHING;