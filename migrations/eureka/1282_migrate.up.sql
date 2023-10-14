ALTER TABLE public.max_score_submission
    ADD COLUMN IF NOT EXISTS total_score INTEGER,
    ADD COLUMN IF NOT EXISTS max_percentage INTEGER;

CREATE OR REPLACE
FUNCTION public.update_max_score_end_date_change_fnc()
RETURNS TRIGGER 
LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    IF exists (select 1 from exam_lo where learning_material_id = NEW.learning_material_id and review_option = 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE') THEN
        WITH mss AS (
        SELECT mgs.student_id, 
        mgs.study_plan_id, 
        mgs.learning_material_id, 
        coalesce(mgs.graded_point,0), 
        coalesce(mgs.total_point,0), 
        coalesce((mgs.graded_point * 1.0 / mgs.total_point) * 100, 0)::smallint max_percentage
        FROM max_graded_score_v2() mgs 
        WHERE mgs.learning_material_id = NEW.learning_material_id and mgs.study_plan_id = NEW.study_plan_id and mgs.student_id = NEW.student_id)
            INSERT INTO
            max_score_submission AS target (
            student_id,
            study_plan_id,
            learning_material_id,
            max_score,
            total_score,
            max_percentage,
            created_at,
            updated_at,
            deleted_at,
            resource_path )
            SELECT
            mss.student_id,
            mss.study_plan_id,
            mss.learning_material_id,
            mss.graded_point,
            mss.total_point,
            mss.max_percentage,
            now(),
            now(),
            NULL,
            NEW.resource_path
        FROM
            mss
        ON CONFLICT ON CONSTRAINT max_score_submission_study_plan_item_identity_pk DO
        UPDATE
        SET
            max_score = EXCLUDED.max_score,
            total_score = EXCLUDED.total_score,
            max_percentage = EXCLUDED.max_percentage,
            updated_at = now();
    END IF;

    RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS update_max_score_end_date_change ON public.individual_study_plan;

CREATE TRIGGER update_max_score_end_date_change
AFTER UPDATE OF end_date ON public.individual_study_plan
FOR EACH ROW
EXECUTE FUNCTION public.update_max_score_end_date_change_fnc();

