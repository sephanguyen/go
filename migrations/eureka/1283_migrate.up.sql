
CREATE OR REPLACE FUNCTION public.update_max_score_exam_lo_once_review_option_change()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$
BEGIN
-- max_graded_score_v2 return the highest score for each student matched with current logic 
	WITH mss AS (
        SELECT
            mgs.student_id, 
			mgs.study_plan_id, 
			mgs.learning_material_id, 
			coalesce(mgs.graded_point,0) AS graded_point,
			coalesce(mgs.total_point,0) AS total_point,
			coalesce((mgs.graded_point * 1.0 / mgs.total_point) * 100, 0)::smallint AS max_percentage
        FROM
            max_graded_score_v2() AS mgs
        WHERE
            mgs.learning_material_id = NEW.learning_material_id)

	INSERT
	INTO
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
	resource_path)
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
    updated_at = now()
WHERE (target.max_score != EXCLUDED.max_score OR target.total_score != EXCLUDED.total_score)
AND target.deleted_at IS NULL;

RETURN NULL;
END;
$function$;

DROP TRIGGER IF EXISTS update_max_score_exam_lo_once_review_option_change ON public.exam_lo;

CREATE TRIGGER update_max_score_exam_lo_once_review_option_change AFTER
UPDATE
	ON
	exam_lo FOR EACH ROW
	WHEN (OLD.review_option IS DISTINCT
FROM
	NEW.review_option) EXECUTE FUNCTION update_max_score_exam_lo_once_review_option_change();
