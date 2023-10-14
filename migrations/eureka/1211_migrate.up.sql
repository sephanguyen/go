-- Trigger study_plan_items to individual_study_plan after insert
CREATE OR REPLACE FUNCTION trigger_study_plan_items_to_individual_study_plan()
    RETURNS TRIGGER
    LANGUAGE 'plpgsql'
AS $FUNCTION$
BEGIN
    -- Two conditions to specify storing individual study plan items
-- 1 - This individual item should differ master item when insert
-- 2 - This individual item is task assigment (without  master) when insert
    WITH temp_table as (
        SELECT nt.*, spi.student_id, COALESCE(spi.master_study_plan_id, spi.study_plan_id) as r_study_plan_id
        FROM study_plans sp
         JOIN new_table nt
              ON nt.study_plan_id = sp.study_plan_id
         JOIN student_study_plans spi ON nt.study_plan_id = spi.study_plan_id
         LEFT JOIN study_plan_items master_spi
              ON master_spi.study_plan_item_id = nt.copy_study_plan_item_id
        WHERE   
                -- 1 - This individual item should differ master item
                (
                    sp.master_study_plan_id is not null
                    -- this is should be differ master
                    AND nt.created_at <> nt.updated_at
                )
                -- 2 - This individual item is task assigment (without master)
                OR (
                        sp.study_plan_type = 'STUDY_PLAN_TYPE_INDIVIDUAL'
                    )
    )
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
    SELECT
        r_study_plan_id,
        COALESCE(NULLIF(content_structure->>'lo_id', ''), content_structure->>'assignment_id'),
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
    FROM temp_table

    ON CONFLICT ON CONSTRAINT learning_material_id_student_id_study_plan_id_pk DO UPDATE SET
     status = EXCLUDED.status,
     start_date = EXCLUDED.start_date,
     end_date = EXCLUDED.end_date,
     available_from = EXCLUDED.available_from,
     available_to = EXCLUDED.available_to,
     school_date = EXCLUDED.school_date,
     updated_at = EXCLUDED.updated_at,
     deleted_at = EXCLUDED.deleted_at;
    RETURN NULL;
END;
$FUNCTION$;
