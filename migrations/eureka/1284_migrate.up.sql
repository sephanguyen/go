CREATE OR REPLACE FUNCTION assignment_graded_score_v2()
    returns table
            (
                student_id           text,
                study_plan_id        text,
                learning_material_id text,
                student_submission_id text,
                graded_point         smallint,
                total_point          smallint,
                status               text,
                passed               bool,
                created_at timestamptz
            )
    language sql
    stable
as
$$
select 
    ss.student_id,
    ss.study_plan_id,
    ss.learning_material_id,
    ss.student_submission_id,
    ssg.grade::smallint as graded_point,
    a.max_grade::smallint as total_point,
    ss.status,
    ss.understanding_level != 'SUBMISSION_UNDERSTANDING_LEVEL_SAD' as passed,
    ss.created_at 
from student_submissions ss
join student_submission_grades ssg on ss.student_submission_id = ssg.student_submission_id
join assignment a using (learning_material_id)
where ssg.grade != -1 and ssg.status = 'SUBMISSION_STATUS_RETURNED';
$$;

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
    where els.status = 'SUBMISSION_STATUS_RETURNED'
group by els.submission_id;
$$;
