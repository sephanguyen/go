CREATE OR REPLACE FUNCTION graded_score()
RETURNS TABLE(
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                submission_id        text,
                graded_points        smallint,
                total_points         smallint,
                status               text
    )
    LANGUAGE sql
    STABLE
AS
$$
	select student_id, study_plan_id, learning_material_id, submission_id, graded_points, total_points, status from lo_graded_score()
	union all
	select student_id, study_plan_id, learning_material_id, submission_id, graded_points, total_points, status from fc_graded_score()
	union all 
	select  student_id, study_plan_id, learning_material_id, student_submission_id, graded_point, total_point, status from get_assignment_scores()
	union all
	select  student_id, study_plan_id, learning_material_id, student_submission_id, graded_point, total_point, status from get_task_assignment_scores()
	union all
	select  student_id, study_plan_id, learning_material_id, submission_id, graded_point, total_point, status from get_exam_lo_scores()
$$;

create or replace function public.max_graded_score()
    RETURNS TABLE
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                graded_points        smallint,
                total_points         smallint
            )
    LANGUAGE sql
    STABLE
AS
$$
select distinct on (student_id, 
    study_plan_id ,
    learning_material_id) student_id,
                          study_plan_id,
                          learning_material_id,
                          graded_points,
                          total_points
from
    graded_score()
where total_points > 0
order by student_id, study_plan_id, learning_material_id, graded_points * 1.0 / total_points desc
$$;
