-- Update get_all_individual_report_of_student
DROP FUNCTION IF EXISTS get_all_individual_report_of_student_v2;

CREATE OR REPLACE FUNCTION public.get_all_individual_report_of_student_v2(report_user_id text, report_course_id text, report_lesson_id text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    with lessonset as (
        select lesson_id from lesson_members lm
        where lm.course_id = report_course_id
        and lm.user_id = report_user_id
        and lm.deleted_at is null
    ),
   previous_lessons as (
   		select l.lesson_id from lessons l
        where l.lesson_id in (select lesson_id from lessonset)
   			and start_time <= (select start_time from lessons where lesson_id = report_lesson_id) 
   			and l.deleted_at is null
   			and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
            and l.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL'
   			order by start_time desc
   			limit 101
   ),
   next_lessons as (
   		select l.lesson_id from lessons l
        where l.lesson_id in (select lesson_id from lessonset)
   			and start_time > (select start_time from lessons where lesson_id = report_lesson_id) 
   			and l.deleted_at is null
   			and l.scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
            and l.teaching_method = 'LESSON_TEACHING_METHOD_INDIVIDUAL'
   			order by start_time asc
   			limit 100
   )
    select * from lesson_reports lr
    where lesson_id in (
   		select lesson_id from previous_lessons
   		union all
   		select lesson_id from next_lessons
   		order by lesson_id desc
   )
   and lr.deleted_at is NULL
$$;

-- Update get_all_group_report_by_lesson_id
DROP FUNCTION IF EXISTS get_all_group_report_by_lesson_id;

CREATE OR REPLACE FUNCTION public.get_all_group_report_by_lesson_id(lesson_id_query text) 
returns setof public.lesson_reports
    language sql stable
    as $$
   with base_lesson as (select * from lessons where lesson_id = lesson_id_query and deleted_at is null),
   previous_lessons as (
   		select * from lessons where class_id = (select bl.class_id from base_lesson bl) 
   			and start_time < (select bl.start_time from base_lesson bl) 
   			and deleted_at is null
   			and scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
   			order by start_time desc
   			limit 100
   ),
   next_lessons as (
   		select * from lessons where class_id = (select bl.class_id from base_lesson bl) 
   			and start_time > (select bl.start_time from base_lesson bl) 
   			and deleted_at is null
   			and scheduling_status <> 'LESSON_SCHEDULING_STATUS_CANCELED'
   			order by start_time asc
   			limit 100
   )
   
   select * from lesson_reports lr
   where lesson_id in (
   		select lesson_id from previous_lessons
   		union all
   		select lesson_id from base_lesson
   		union all
   		select lesson_id from next_lessons
   		order by lesson_id desc
   )
   and lr.deleted_at is NULL
$$;
