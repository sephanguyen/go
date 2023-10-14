-- Update get_all_individual_report_of_student
DROP FUNCTION IF EXISTS get_all_individual_report_of_student;

CREATE OR REPLACE FUNCTION public.get_all_individual_report_of_student(report_user_id text, report_course_id text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lesson_members lm on lr.lesson_id = lm.lesson_id
    join lessons l on l.lesson_id=lr.lesson_id and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
where lm.user_id = report_user_id
    and lm.course_id = report_course_id
    and l.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL'
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
    and lm.deleted_at is NULL
order by
    l.start_time desc;
$$;

-- Update get_all_group_report_by_lesson_id
DROP FUNCTION IF EXISTS get_all_group_report_by_lesson_id;

CREATE OR REPLACE FUNCTION public.get_all_group_report_by_lesson_id(lesson_id_query text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lessons l on l.lesson_id = lr.lesson_id and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
where l.class_id = (
    select distinct class_id
      from lessons l1 where l1.lesson_id = lesson_id_query 
      and l1.deleted_at is NULL
      limit 1)
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
order by
    l.start_time desc
$$;

-- Create get_previous_lesson_report_group_v2
DROP FUNCTION IF EXISTS get_previous_lesson_report_group_v2;

CREATE OR REPLACE FUNCTION public.get_previous_lesson_report_group_v2(lesson_id_query text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lessons l on l.lesson_id = lr.lesson_id and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
where l.start_time < (
    select l1.start_time 
      from lessons l1 where l1.lesson_id = lesson_id_query 
      and l1.class_id = l.class_id
      and l1.deleted_at is NULL limit 1)
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
order by
    l.start_time desc
limit 1;
$$;

-- Create get_previous_report_of_student_v5
DROP FUNCTION IF EXISTS get_previous_report_of_student_v5;

CREATE OR REPLACE FUNCTION public.get_previous_report_of_student_v5(report_user_id text, report_course_id text, report_id text, report_lesson_id text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
    join lesson_members lm on lr.lesson_id = lm.lesson_id
    join lessons l on l.lesson_id=lr.lesson_id and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
where
    CASE WHEN report_id IS NOT NULL 
        THEN l.start_time < (
                select l1.start_time 
                    from lessons l1 join lesson_reports lr1 on l1.lesson_id=lr1.lesson_id
                    where lr1.lesson_report_id = report_id and l1.deleted_at is NULL and lr1.deleted_at is NULL 
                    and l1.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL' limit 1)
        ELSE l.start_time < (
                CASE WHEN report_lesson_id IS NOT NULL 
                    THEN
                        (select l2.start_time 
                            from lessons l2 where l2.lesson_id = report_lesson_id and l2.deleted_at is NULL
                            and l2.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL' limit 1)
                    ELSE now()
                END
            )
    END
    and lm.user_id = report_user_id
    and lm.course_id = report_course_id
    and l.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL'
    and lr.deleted_at is NULL
    and l.deleted_at is NULL
    and lm.deleted_at is NULL
order by
    l.start_time desc
limit 1;
$$;
