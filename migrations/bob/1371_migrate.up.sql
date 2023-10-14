-- remove old functions
DROP FUNCTION IF EXISTS public.student_course_purchased_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[]);
DROP FUNCTION IF EXISTS public.student_course_assigned_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[]);
DROP FUNCTION IF EXISTS public.student_course_recurring_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date);

-- RECURRING TAB - purchased slot
CREATE OR REPLACE FUNCTION student_course_purchased_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], use_basic_info boolean)
RETURNS TABLE (
    unique_id text,
    student_id text,
    course_id text,
    location_id text,
    week_start date,
    week_end date,
    purchased_slot int,
    loc_created_at timestamptz,
    loc_updated_at timestamptz,
    resource_path text
)
LANGUAGE SQL
SECURITY INVOKER
AS $func$
select 
base_slot_info.unique_id, 
base_slot_info.student_id,
base_slot_info.course_id, 
base_slot_info.location_id, 
base_slot_info.week_start,
(base_slot_info.week_start + interval '6 days')::date as "week_end",
base_slot_info.purchased_slot, 
loc.created_at as "loc_created_at",
loc.updated_at as "loc_updated_at",
base_slot_info.resource_path
from (select 
    CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id) as "unique_id", 
    sc.student_id,
    sc.course_id, 
    sc.location_id,
    generate_series(to_date(to_char(sc.student_start_date, 'IYYY-IW'), 'IYYY-IW')
        ,to_date(to_char(sc.student_end_date, 'IYYY-IW'), 'IYYY-IW')
        ,interval '1 week'
    )::date  as "week_start", 
    sc.course_slot_per_week as "purchased_slot",
    sc.resource_path
    from student_course sc
    where sc.student_end_date > current_date
    and sc.package_type = 'PACKAGE_TYPE_FREQUENCY'
    and case when keyword is not null and use_basic_info = true then exists (select 1 from user_basic_info ubi where ubi.user_id = sc.student_id and ubi.deleted_at is null and lower(ubi.name) like lower(concat('%',keyword,'%'))) else true end
    and case when keyword is not null and use_basic_info = false then exists (select 1 from users u where u.user_id = sc.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
    and case when student_ids is not null then sc.student_id = ANY(student_ids) else true end
    and case when course_ids is not null then sc.course_id = ANY(course_ids) else true end
    and case when location_ids is not null then sc.location_id = ANY(location_ids) else true end
    and sc.deleted_at is null
    UNION
    select 
    student_info.unique_id, 
    student_info.student_id,
    student_info.course_id, 
    student_info.location_id, 
    student_info.week_start,
    coalesce(class_info.lesson_count, 0)::int as "purchased_slot",
    student_info.resource_path
    from (select 
        CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id) as "unique_id",
        sc.student_id,
        sc.course_id, 
        sc.location_id,
        generate_series(to_date(to_char(sc.student_start_date, 'IYYY-IW'), 'IYYY-IW')
            ,to_date(to_char(sc.student_end_date, 'IYYY-IW'), 'IYYY-IW')
            ,interval '1 week'
        )::date as "week_start",
        sc.resource_path
        from student_course sc
        where sc.student_end_date > current_date
        and sc.package_type = 'PACKAGE_TYPE_SCHEDULED'
        and case when keyword is not null and use_basic_info = true then exists (select 1 from user_basic_info ubi where ubi.user_id = sc.student_id and ubi.deleted_at is null and lower(ubi.name) like lower(concat('%',keyword,'%'))) else true end
        and case when keyword is not null and use_basic_info = false then exists (select 1 from users u where u.user_id = sc.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
        and case when student_ids is not null then sc.student_id = ANY(student_ids) else true end
        and case when course_ids is not null then sc.course_id = ANY(course_ids) else true end
        and case when location_ids is not null then sc.location_id = ANY(location_ids) else true end
        and sc.deleted_at is null) student_info
    left join (select 
        CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id) as "unique_id",
        cl.course_id,
        clm.user_id,
        to_char(l.start_time, 'IYYY-IW') as "week",
        count(l.lesson_id) as "lesson_count"
        from student_course sc
        inner join class cl
        on cl.course_id = sc.course_id
        inner join class_member clm
        on cl.class_id = clm.class_id and sc.student_id = clm.user_id
        inner join lessons l
        on l.class_id = cl.class_id and l.course_id = sc.course_id and l.center_id = cl.location_id
        where sc.student_end_date > current_date
        and sc.package_type = 'PACKAGE_TYPE_SCHEDULED'
        and case when keyword is not null and use_basic_info = true then exists (select 1 from user_basic_info ubi where ubi.user_id = sc.student_id and ubi.deleted_at is null and lower(ubi.name) like lower(concat('%',keyword,'%'))) else true end
        and case when keyword is not null and use_basic_info = false then exists (select 1 from users u where u.user_id = sc.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
        and case when student_ids is not null then sc.student_id = ANY(student_ids) else true end
        and case when course_ids is not null then sc.course_id = ANY(course_ids) else true end
        and case when location_ids is not null then sc.location_id = ANY(location_ids) else true end
        and l.teaching_method = 'LESSON_TEACHING_METHOD_GROUP'
        and l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
        and l.start_time::date between sc.student_start_date::date and sc.student_end_date::date
        and l.class_id is not NULL
        and sc.deleted_at is null
        and cl.deleted_at is null
        and clm.deleted_at is null
        and l.deleted_at is null
        group by CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id), cl.course_id, clm.user_id, to_char(l.start_time, 'IYYY-IW')
    ) class_info
    on class_info.unique_id = student_info.unique_id 
    and class_info.week = to_char(student_info.week_start, 'IYYY-IW')
    and class_info.course_id = student_info.course_id
    and class_info.user_id = student_info.student_id) base_slot_info
inner join locations loc
on loc.location_id = base_slot_info.location_id
and loc.deleted_at is null
;
$func$;

-- RECURRING TAB - assigned slot
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
count(lm.lesson_id)::int as "assigned_slot"
from student_course sc 
inner join lesson_members lm 
on lm.user_id = sc.student_id
inner join lessons l
on l.lesson_id = lm.lesson_id and l.course_id = sc.course_id and l.center_id = sc.location_id
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
group by CONCAT(sc.student_id, sc.course_id, sc.location_id, sc.student_package_id), sc.course_id, sc.student_id, to_char(l.start_time, 'IYYY-IW'), sc.resource_path
;
$func$;

-- RECURRING TAB - purchased and assigned slots calculation
CREATE OR REPLACE FUNCTION student_course_recurring_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date, use_basic_info boolean)
RETURNS TABLE (
    student_id text,
    course_id text,
    location_id text,
    week_start date,
    week_end date,
    purchased_slot int,
    assigned_slot int,
    slot_gap int,
    status text,
    unique_id text,
    resource_path text
) LANGUAGE SQL
AS $$
select 
scprs.student_id,
scprs.course_id,
scprs.location_id,
scprs.week_start,
scprs.week_end,
scprs.purchased_slot, 
coalesce(scars.assigned_slot, 0)::int as "assigned_slot", 
(coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int as "slot_gap",
case 
	when (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int < 0 then 'Under assigned'
	when (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int = 0 then 'Just assigned'
	when (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int > 0 then 'Over assigned'
end as "status",
scprs.unique_id,
scprs.resource_path
from student_course_purchased_recurring_slot_fn(keyword, student_ids, course_ids, location_ids, use_basic_info) scprs
left join student_course_assigned_recurring_slot_fn(keyword, student_ids, course_ids, location_ids, use_basic_info) scars
on scprs.unique_id = scars.unique_id
and to_char(scprs.week_start, 'IYYY-IW') = scars.week
where case when start_date is not null then scprs.week_start >= start_date::Date else true end
and case when end_date is not null then scprs.week_end <= end_date::Date else true end
order by scprs.week_start asc, scprs.week_end asc, scprs.loc_updated_at asc, scprs.loc_created_at asc, scprs.course_id asc, scprs.student_id asc
;
$$;