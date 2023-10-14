Drop function if exists student_study_plans_fn();
Create or replace  function student_study_plans_fn() returns table(course_id text, student_id text, study_plan_id text, study_plan_type text, deleted_at timestamptz)
language sql stable as
$$
select cs.course_id, cs.student_id, sp.study_plan_id, sp.study_plan_type, cs.deleted_at
from study_plans sp
join course_students cs on cs.course_id = sp.course_id
join student_study_plans ssp
     on cs.student_id = ssp.student_id and sp.study_plan_id = ssp.study_plan_id
where sp.master_study_plan_id is null
union
select cs.course_id, cs.student_id, sp.study_plan_id, sp.study_plan_type, cs.deleted_at
from study_plans sp
join course_students cs on cs.course_id = sp.course_id
where sp.master_study_plan_id is null
$$;
