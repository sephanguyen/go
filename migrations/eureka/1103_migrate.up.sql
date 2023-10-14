DROP TRIGGER IF EXISTS trigger_study_plan_items_to_individual_study_plan on study_plan_items;

TRUNCATE public.individual_study_plan;

-- Trigger study_plan_items to individual_study_plan
CREATE OR REPLACE FUNCTION trigger_study_plan_items_to_individual_study_plan()
    RETURNS TRIGGER
    LANGUAGE 'plpgsql'
AS $FUNCTION$
BEGIN
-- Two conditions to specify storing individual study plan items 
-- 1 - This item should be individual item when update
-- 2 - This individual item is task assigment (without  master) when insert
    IF EXISTS (
        SELECT 1
        FROM study_plans sp
        LEFT JOIN study_plan_items master_spi
            ON master_spi.study_plan_item_id = NEW.copy_study_plan_item_id
        WHERE sp.study_plan_id = NEW.study_plan_id
            AND (
                -- 1 - This item should be individual item when update
                (
                    sp.master_study_plan_id is not null 
                    AND NEW.created_at <> NEW.updated_at
                )
                -- 2 - This individual item is task assigment (without master)
                OR sp.study_plan_type = 'STUDY_PLAN_TYPE_INDIVIDUAL'
            )
    )
    THEN
        INSERT INTO public.individual_study_plan (
            study_plan_id,
            learning_material_id,
            student_id,
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
            (SELECT COALESCE(master_study_plan_id, study_plan_id) FROM public.student_study_plans ssp WHERE ssp.study_plan_id = NEW.study_plan_id),
            COALESCE(NULLIF(NEW.content_structure->>'lo_id', ''), NEW.content_structure->>'assignment_id'),
            (SELECT student_id FROM public.student_study_plans spi WHERE spi.study_plan_id = NEW.study_plan_id),
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
        ON CONFLICT ON CONSTRAINT learning_material_id_student_id_study_plan_id_pk DO UPDATE SET
            status = NEW.status,
            start_date = NEW.start_date,
            end_date = NEW.end_date,
            available_from = NEW.available_from,
            available_to = NEW.available_to,
            school_date = NEW.school_date,
            updated_at = NEW.updated_at,
            deleted_at = NEW.deleted_at;
    END IF;

    -- Delete invididual study plan if master and individual are same
    IF EXISTS (
        SELECT 1
        FROM study_plans sp
        LEFT JOIN study_plan_items master_spi
            ON master_spi.study_plan_item_id = NEW.copy_study_plan_item_id
        WHERE sp.study_plan_id = NEW.study_plan_id
            AND sp.master_study_plan_id is not null 
            AND ( NEW.start_date = master_spi.start_date or coalesce(NEW.start_date, master_spi.start_date) is null)
            AND ( NEW.available_from = master_spi.available_from or coalesce(NEW.available_from, master_spi.available_from) is null)
            AND ( NEW.available_to = master_spi.available_to or coalesce(NEW.available_to, master_spi.available_to) is null)
            AND ( NEW.end_date = master_spi.end_date or coalesce(NEW.end_date, master_spi.end_date) is null)
            AND ( NEW.school_date = master_spi.school_date or coalesce(NEW.school_date, master_spi.school_date) is null)
            AND  NEW.status = master_spi.status
    )
    THEN
        DELETE FROM individual_study_plan 
        WHERE 
            study_plan_id = (
                SELECT master_study_plan_id 
                FROM public.student_study_plans ssp 
                WHERE ssp.study_plan_id = NEW.study_plan_id
            )
            AND learning_material_id = COALESCE(NULLIF(NEW.content_structure->>'lo_id', ''), NEW.content_structure->>'assignment_id')
            AND student_id = (SELECT student_id FROM public.student_study_plans spi WHERE spi.study_plan_id = NEW.study_plan_id);
    END IF;

RETURN NULL;
END;
$FUNCTION$;

CREATE TRIGGER trigger_study_plan_items_to_individual_study_plan
AFTER INSERT OR UPDATE ON public.study_plan_items
FOR EACH ROW
EXECUTE FUNCTION public.trigger_study_plan_items_to_individual_study_plan();