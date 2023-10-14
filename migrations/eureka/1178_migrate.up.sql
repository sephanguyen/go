CREATE OR REPLACE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn()
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
-- Condition to specify storing master study plan items
-- 1: UPDATE only
-- WHY don't store insert, which always insert null values
    WITH temp_table as (
        SELECT nt.*
        FROM public.study_plans sp
        JOIN new_table nt 
            ON nt.study_plan_id = sp.study_plan_id
        WHERE sp.master_study_plan_id IS NULL
          AND sp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE'
          AND nt.created_at <> nt.updated_at
    )
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
    SELECT
        study_plan_id,
        coalesce(NULLIF(content_structure ->> 'lo_id',''),content_structure->>'assignment_id'),
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
    FROM temp_table

    ON CONFLICT ON CONSTRAINT learning_material_id_study_plan_id_pk DO UPDATE SET
      start_date = EXCLUDED.start_date,
      end_date = EXCLUDED.end_date,
      available_from = EXCLUDED.available_from,
      available_to = EXCLUDED.available_to,
      school_date = EXCLUDED.school_date,
      updated_at = EXCLUDED.updated_at,
      status = EXCLUDED.status,
      deleted_at = EXCLUDED.deleted_at;
    RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_to_master_study_plan ON public.study_plan_items;

DROP TRIGGER IF EXISTS migrate_to_master_study_plan_ins ON public.study_plan_items;
CREATE TRIGGER migrate_to_master_study_plan_ins
    AFTER INSERT
    ON public.study_plan_items
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT EXECUTE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn();

DROP TRIGGER IF EXISTS migrate_to_master_study_plan_udt ON public.study_plan_items;
CREATE TRIGGER migrate_to_master_study_plan_udt
    AFTER UPDATE
    ON public.study_plan_items
    REFERENCING NEW TABLE AS new_table
    FOR EACH STATEMENT EXECUTE FUNCTION public.migrate_study_plan_items_to_master_study_plan_fn();

