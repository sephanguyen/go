ALTER TABLE public.learning_objectives
    ADD COLUMN IF NOT EXISTS instruction TEXT;

ALTER TABLE public.exam_lo
    ADD COLUMN IF NOT EXISTS instruction TEXT;


CREATE OR REPLACE FUNCTION public.migrate_learning_objectives_to_exam_lo_fn()
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    INSERT INTO exam_lo (
        instruction,
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
        NEW.instruction,
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
        instruction = NEW.instruction,
        updated_at = NEW.updated_at,
        deleted_at = NEW.deleted_at;
    RETURN NULL;
END;
$BODY$;
