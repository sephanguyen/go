DROP function if exists get_student_study_plans_by_filter_v2();
CREATE or replace FUNCTION public.get_student_study_plans_by_filter_v2() RETURNS table(
    study_plan_id         text,
    master_study_plan_id  text,
    name                  text,
    study_plan_type       text,
    school_id             integer,
    created_at            timestamp with time zone,
    updated_at            timestamp with time zone,
    deleted_at            timestamp with time zone,
    course_id             text,
    resource_path         text,
    book_id               text,
    status                text,
    track_school_progress boolean,
    grades                integer[],
    student_id            text
)
    LANGUAGE sql
    STABLE
AS
$$

select st.*, s_study_plans.student_id
from public.student_study_plans_fn() as s_study_plans
         join study_plans as st
              USING (study_plan_id)
$$;

create view get_student_study_plans_by_filter_view AS
select 
    get_student_study_plans_by_filter_v2.study_plan_id         ,
    get_student_study_plans_by_filter_v2.master_study_plan_id  ,
    get_student_study_plans_by_filter_v2.name                  ,
    get_student_study_plans_by_filter_v2.study_plan_type       ,
    get_student_study_plans_by_filter_v2.school_id             ,
    get_student_study_plans_by_filter_v2.created_at            ,
    get_student_study_plans_by_filter_v2.updated_at            ,
    get_student_study_plans_by_filter_v2.deleted_at            ,
    get_student_study_plans_by_filter_v2.course_id             ,
    get_student_study_plans_by_filter_v2.resource_path         ,
    get_student_study_plans_by_filter_v2.book_id               ,
    get_student_study_plans_by_filter_v2.status                ,
    get_student_study_plans_by_filter_v2.track_school_progress ,
    get_student_study_plans_by_filter_v2.grades                ,
    get_student_study_plans_by_filter_v2.student_id            
FROM get_student_study_plans_by_filter_v2() get_student_study_plans_by_filter_v2(
    study_plan_id         ,
    master_study_plan_id  ,
    name                  ,
    study_plan_type       ,
    school_id             ,
    created_at            ,
    updated_at            ,
    deleted_at            ,
    course_id             ,
    resource_path         ,
    book_id               ,
    status                ,
    track_school_progress ,
    grades                ,
    student_id            
);

