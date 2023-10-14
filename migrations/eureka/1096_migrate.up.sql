CREATE OR REPLACE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn()
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
-- Condition to specify storing master study plan items 
-- 1: UPDATE only
-- WHY don't store insert, which always insert null values
    IF EXISTS (
        SELECT * 
        FROM public.study_plans sp
        WHERE sp.study_plan_id = NEW.study_plan_id 
            AND sp.master_study_plan_id IS NULL
            AND sp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE'
            AND TG_OP = 'UPDATE'
    )
    THEN 
    INSERT INTO master_study_plan (
        study_plan_id,
        learning_material_id,
        status,
        start_date,
        end_date,
        available_from,
        available_to,
        created_at,
        updated_at,
        deleted_at,
        school_date,
        resource_path
    )
    VALUES (
        NEW.study_plan_id,
        coalesce(NULLIF(new.content_structure ->> 'lo_id',''),new.content_structure->>'assignment_id'),
        NEW.status,
        NEW.start_date,
        NEW.end_date,
        NEW.available_from,
        NEW.available_to,
        NEW.created_at,
        NEW.updated_at,
        NEW.deleted_at,
        NEW.school_date,
        NEW.resource_path
    )
    ON CONFLICT ON CONSTRAINT learning_material_id_study_plan_id_pk DO UPDATE SET
    status = NEW.status,
    start_date = NEW.start_date,
    end_date = NEW.end_date,
    available_from = NEW.available_from,
    available_to = NEW.available_to,
    school_date = NEW.school_date,
    updated_at = NEW.updated_at,
    deleted_at = NEW.deleted_at;
    END IF; 
    RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_to_master_study_plan ON public.study_plan_items;
CREATE TRIGGER migrate_to_master_study_plan
    AFTER INSERT OR UPDATE
    ON public.study_plan_items
    FOR EACH ROW
    EXECUTE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn();
