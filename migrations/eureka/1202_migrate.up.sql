-- This function get exam lo scores
create or replace function get_exam_lo_scores()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                submission_id        text,
                graded_point         smallint,
                total_point          smallint,
                status               text,
                result               text,
                created_at           timestamptz
            )
    language sql
    stable
as
$$
select els.student_id,
       els.study_plan_id,
       els.learning_material_id,
       els.submission_id,
    --   when teacher manual grading, we should use with score from teacher
       sum(coalesce(elss.point, elsa.point))::smallint as graded_point,
       els.total_point::smallint                       as total_point,
       els.status,
       els.result,
       els.created_at
from exam_lo_submission els
         join exam_lo_submission_answer elsa using (submission_id)
         left join exam_lo_submission_score elss using (submission_id, quiz_id)
where els.deleted_at is null and
    elsa.deleted_at is null and
    elss.deleted_at is null
group by els.student_id,
         els.study_plan_id,
         els.learning_material_id,
         els.submission_id
$$;
