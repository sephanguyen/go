DROP FUNCTION IF EXISTS get_all_individual_report_of_student;

CREATE OR REPLACE FUNCTION public.get_all_individual_report_of_student(report_user_id text, report_course_id text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lesson_members lm on lr.lesson_id = lm.lesson_id
    join lessons l on l.lesson_id=lr.lesson_id
where lm.user_id = report_user_id
    and lm.course_id = report_course_id
    and l.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL'
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
    and lm.deleted_at is NULL
order by
    l.start_time desc;
$$;
