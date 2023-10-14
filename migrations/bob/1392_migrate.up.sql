CREATE OR REPLACE FUNCTION student_course_assigned_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], use_basic_info boolean)
RETURNS TABLE (
    unique_id text,
    course_id text,
    student_id text,
    resource_path text,
    assigned_slot int
) LANGUAGE SQL
SECURITY INVOKER
AS $func$
select 
CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id) as "unique_id",
sc.course_id, 
sc.student_id,
sc.resource_path,
count(case WHEN attendance_status != 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN 1		 				
  		   WHEN attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' and r.new_lesson_id is null THEN 1
  		   ELSE null END )::int as "assigned_slot"
from student_course sc
inner join lesson_members lm 
on lm.user_id = sc.student_id and lm.course_id = sc.course_id
inner join lessons l
on l.lesson_id = lm.lesson_id and l.center_id = sc.location_id
left join reallocation r on r.original_lesson_id = lm.lesson_id and r.student_id = lm.user_id 
where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
and sc.package_type IN ('PACKAGE_TYPE_ONE_TIME', 'PACKAGE_TYPE_SLOT_BASED')
and l.start_time::date between sc.student_start_date::date and sc.student_end_date::date
and sc.student_end_date > current_date
and case when keyword is not null and use_basic_info = true then exists (select 1 from user_basic_info ubi where ubi.user_id = sc.student_id and ubi.deleted_at is null and lower(ubi.name) like lower(concat('%',keyword,'%'))) else true end
and case when keyword is not null and use_basic_info = false then exists (select 1 from users u where u.user_id = sc.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
and case when student_ids is not null then sc.student_id = ANY(student_ids) else true end
and case when course_ids is not null then sc.course_id = ANY(course_ids) else true end
and case when location_ids is not null then sc.location_id = ANY(location_ids) else true end
and sc.deleted_at is null
and l.deleted_at is null
and lm.deleted_at is null
and r.deleted_at is null
group by CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id), sc.course_id, sc.student_id, sc.resource_path
;
$func$;


CREATE OR REPLACE FUNCTION student_course_assigned_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], use_basic_info boolean)
RETURNS TABLE (
    unique_id text,
    course_id text,
    student_id text,
    week text,
    resource_path text,
    assigned_slot int
) LANGUAGE SQL
SECURITY INVOKER
AS $func$
select 
CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id) as "unique_id",
sc.course_id, 
sc.student_id,
to_char(l.start_time, 'IYYY-IW') as "week",
sc.resource_path,
count(case WHEN attendance_status != 'STUDENT_ATTEND_STATUS_REALLOCATE' THEN 1		 				
  		   WHEN attendance_status = 'STUDENT_ATTEND_STATUS_REALLOCATE' and r.new_lesson_id is null THEN 1
  		   ELSE null END )::int as "assigned_slot"
from student_course sc 
inner join lesson_members lm 
on lm.user_id = sc.student_id and lm.course_id = sc.course_id
inner join lessons l
on l.lesson_id = lm.lesson_id and l.center_id = sc.location_id
left join reallocation r on r.original_lesson_id = lm.lesson_id and r.student_id = lm.user_id 
where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
and sc.package_type IN ('PACKAGE_TYPE_FREQUENCY', 'PACKAGE_TYPE_SCHEDULED')
and l.start_time::date between sc.student_start_date::date and sc.student_end_date::date
and sc.student_end_date > current_date
and case when keyword is not null and use_basic_info = true then exists (select 1 from user_basic_info ubi where ubi.user_id = sc.student_id and ubi.deleted_at is null and lower(ubi.name) like lower(concat('%',keyword,'%'))) else true end
and case when keyword is not null and use_basic_info = false then exists (select 1 from users u where u.user_id = sc.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
and case when student_ids is not null then sc.student_id = ANY(student_ids) else true end
and case when course_ids is not null then sc.course_id = ANY(course_ids) else true end
and case when location_ids is not null then sc.location_id = ANY(location_ids) else true end
and sc.deleted_at is null
and l.deleted_at is null
and lm.deleted_at is null
and r.deleted_at is null
group by CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id), sc.course_id, sc.student_id, to_char(l.start_time, 'IYYY-IW'), sc.resource_path
;
$func$;
