ALTER TABLE IF EXISTS public.learning_objectives
    ADD COLUMN IF NOT EXISTS time_limit INTEGER DEFAULT NULL;

ALTER TABLE IF EXISTS public.exam_lo
    ADD COLUMN IF NOT EXISTS time_limit INTEGER DEFAULT NULL;

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
        resource_path,
        instruction,
        grade_to_pass,
        manual_grading,
        time_limit
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
        NEW.resource_path,
        NEW.instruction,
        NEW.grade_to_pass,
        NEW.manual_grading,
        NEW.time_limit
    ) 
    ON CONFLICT ON CONSTRAINT exam_lo_pk DO UPDATE SET 
        topic_id = NEW.topic_id,
        name = NEW.name,
        display_order = NEW.display_order,
        updated_at = NEW.updated_at,
        deleted_at = NEW.deleted_at,
        instruction = NEW.instruction,
        grade_to_pass = NEW.grade_to_pass,
        manual_grading = NEW.manual_grading,
        time_limit = NEW.time_limit;
    RETURN NULL;
END;
$BODY$;