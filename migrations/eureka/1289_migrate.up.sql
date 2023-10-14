
CREATE OR REPLACE FUNCTION public.exam_lo_graded_score_v2() RETURNS TABLE(student_id text, study_plan_id text, learning_material_id text, submission_id text, graded_point smallint, total_point smallint, status text, result text, created_at timestamp with time zone)
    LANGUAGE sql STABLE
    AS $$
 select els.student_id,
    els.study_plan_id,
    els.learning_material_id,
    els.submission_id,
    sum(coalesce(elss.point, elsa.point))::smallint as graded_point,
    els.total_point::smallint as total_point,
    els.status,
    els.result,
    els.created_at
from exam_lo_submission els
    join exam_lo_submission_answer elsa using (submission_id)
    left join exam_lo_submission_score elss using (submission_id, quiz_id)
    where els.status = 'SUBMISSION_STATUS_RETURNED' AND els.deleted_at IS NULL
group by els.submission_id;
$$;


CREATE OR REPLACE FUNCTION public.update_max_score_exam_lo_once_exam_lo_submission_status_change() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
		WITH mss AS (
	        SELECT
				coalesce(mgs.graded_point,0) AS graded_point,
				coalesce(mgs.total_point,0) AS total_point,
				coalesce((mgs.graded_point * 1.0 / mgs.total_point) * 100, 0)::smallint AS max_percentage
	        FROM
	            max_graded_score_v2() AS mgs
	        WHERE
	            mgs.learning_material_id = NEW.learning_material_id AND 
	            mgs.study_plan_id = NEW.study_plan_id AND 
	            mgs.student_id = NEW.student_id),
	            update_when_mss_empty AS (
               	UPDATE max_score_submission
			    SET updated_at = now(), max_score = NULL, max_percentage = NULL
			    WHERE NOT EXISTS (SELECT 1 FROM mss) AND learning_material_id = NEW.learning_material_id AND study_plan_id = NEW.study_plan_id AND student_id = NEW.student_id)
	           
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
		NEW.student_id,
		NEW.study_plan_id,
		NEW.learning_material_id,
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
RETURN NULL;
END;
$$;


DROP TRIGGER IF EXISTS update_max_score_exam_lo_once_exam_lo_submission_status_change ON public.exam_lo_submission;

CREATE TRIGGER update_max_score_exam_lo_once_exam_lo_submission_status_change AFTER UPDATE OF status, deleted_at
ON public.exam_lo_submission FOR EACH ROW
EXECUTE PROCEDURE public.update_max_score_exam_lo_once_exam_lo_submission_status_change();
