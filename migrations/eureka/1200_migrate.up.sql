-- this function get exam lo scores with latest score mode
create or replace function get_exam_lo_scores_latest_score()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                graded_point         smallint,
                total_point          smallint,
                status               text,
                passed               bool,
                total_attempts       smallint

            )
    language sql
    stable
as
$$
select distinct on (student_id, study_plan_id, learning_material_id) student_id,
                                                                     study_plan_id,
                                                                     learning_material_id,
                                                                     graded_point,
                                                                     total_point,
                                                                     status,
                                                                    --  true when a submission is passed
                                                                     bool_or(result = 'EXAM_LO_SUBMISSION_PASSED')
                                                                     over ( partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id )          as passed,
                                                                     count(*) over (partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id)::smallint as total_attempts
from get_exam_lo_scores()
-- get the latest scores by created_at
order by student_id, study_plan_id, learning_material_id, created_at desc;
$$;

create or replace function get_exam_lo_scores_grade_to_pass()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                graded_point         smallint,
                total_point          smallint,
                status               text,
                passed               bool,
                total_attempts       smallint
            )
    language sql
    stable
as
$$
select distinct on (student_id, study_plan_id, learning_material_id) student_id,
                                                                     study_plan_id,
                                                                     learning_material_id,
                                                                     -- graded point is calculated
                                                                     -- if all submission are fails choose latest score
                                                                     -- if a submission is passed choose grade_to_pass from exam_lo setting
                                                                     -- if a submission is passed from 2nd
                                                                     -- ex : s1 failed, s2 pass, s3 pass
                                                                     -- we will calculate s2 = pass * count(s1,s2,s3) > 1
                                                                     coalesce(NULLIF(
                                                                             (e.grade_to_pass *
                                                                              (result = 'EXAM_LO_SUBMISSION_PASSED')::integer *
                                                                              (count(*) over (
                                                                                  partition by student_id,
                                                                                      study_plan_id,
                                                                                      learning_material_id
                                                                                  ) > 1)::integer)::smallint,0)
                                                                         , s.graded_point)                   as graded_point,
                                                                     total_point,
                                                                     status,
                                                                     --  true if submission differ failed
                                                                     (result = 'EXAM_LO_SUBMISSION_PASSED') as passed,
                                                                     count(*) over (partition by student_id,
                                                                         study_plan_id,
                                                                         learning_material_id)::smallint     as total_attempts
from get_exam_lo_scores() s
         join exam_lo e using (learning_material_id)
-- order by the submissions are passed -> then latest
order by student_id, study_plan_id, learning_material_id, (result = 'EXAM_LO_SUBMISSION_PASSED') desc, s.created_at desc
$$;
