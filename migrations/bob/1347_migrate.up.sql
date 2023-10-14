-- remove old views
DROP VIEW IF EXISTS public.student_course_purchased_slot;
DROP VIEW IF EXISTS public.student_course_assigned_slot;
DROP VIEW IF EXISTS public.student_course_slot_info;

-- remove old functions
DROP FUNCTION IF EXISTS public.student_course_purchased_slot_fn();
DROP FUNCTION IF EXISTS public.student_course_assigned_slot_fn();
DROP FUNCTION IF EXISTS public.student_course_slot_info_fn();

-- SLOT TAB - purchased slot
CREATE OR REPLACE FUNCTION student_course_purchased_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[])
RETURNS TABLE (
    student_subscription_id text,
    student_id text,
    course_id text,
    location_id text,
    student_start_date date,
    student_end_date date,
    purchased_slot int,
    loc_created_at timestamptz,
    loc_updated_at timestamptz,
    resource_path text
)
LANGUAGE SQL
SECURITY INVOKER
AS $func$
select
base_slot_info.student_subscription_id, 
base_slot_info.student_id,
base_slot_info.course_id, 
lssap.location_id, 
base_slot_info.start_at::DATE as "student_start_date",  
base_slot_info.end_at::DATE as "student_end_date", 
base_slot_info.purchased_slot, 
loc.created_at as "loc_created_at",
loc.updated_at as "loc_updated_at",
base_slot_info.resource_path
from (select 
    lss.student_subscription_id, 
    lss.student_id,
    lss.course_id, 
    lss.start_at,  
    lss.end_at, 
    lss.course_slot as "purchased_slot",
    lss.resource_path
    from lesson_student_subscriptions lss
    where lss.end_at > current_date
    and lss.package_type = 'PACKAGE_TYPE_SLOT_BASED'
    and case when keyword is not null then lower(concat(lss.student_last_name, ' ', lss.student_first_name)) like lower(concat('%',keyword,'%')) else true end
    and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
    and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
    and lss.deleted_at is null
    UNION
    select 
    lss.student_subscription_id, 
    lss.student_id,
    lss.course_id, 
    lss.start_at,  
    lss.end_at, 
    coalesce(class_info.lesson_count, 0)::int as "purchased_slot",
    lss.resource_path
    from lesson_student_subscriptions lss
    left join (select 
        lss.student_subscription_id,
        cl.course_id,
        clm.user_id,
        count(l.lesson_id) as "lesson_count"
        from lesson_student_subscriptions lss
        inner join class cl
        on cl.course_id = lss.course_id
        inner join class_member clm
        on cl.class_id = clm.class_id and lss.student_id = clm.user_id
        inner join lessons l
        on l.class_id = cl.class_id and l.course_id = lss.course_id and l.center_id = cl.location_id
        where lss.end_at > current_date
        and lss.package_type = 'PACKAGE_TYPE_ONE_TIME'
        and case when keyword is not null then lower(concat(lss.student_last_name, ' ', lss.student_first_name)) like lower(concat('%',keyword,'%')) else true end
        and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
        and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
        and l.teaching_method = 'LESSON_TEACHING_METHOD_GROUP'
        and l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
        and l.start_time::date between lss.start_at::date and lss.end_at::date
        and l.class_id is not NULL
        and lss.deleted_at is null
        and cl.deleted_at is null
        and clm.deleted_at is null
        and l.deleted_at is null
        group by lss.student_subscription_id, cl.course_id, clm.user_id
    ) class_info
    on class_info.student_subscription_id = lss.student_subscription_id
    and class_info.course_id = lss.course_id
    and class_info.user_id = lss.student_id
    where lss.end_at > current_date
    and lss.package_type = 'PACKAGE_TYPE_ONE_TIME'
    and case when keyword is not null then lower(concat(lss.student_last_name, ' ', lss.student_first_name)) like lower(concat('%',keyword,'%')) else true end
    and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
    and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
    and lss.deleted_at is null) base_slot_info
inner join lesson_student_subscription_access_path lssap
on base_slot_info.student_subscription_id = lssap.student_subscription_id
inner join locations loc
on loc.location_id = lssap.location_id
where case when location_ids is not null then lssap.location_id = ANY(location_ids) else true end
and lssap.deleted_at is null
and loc.deleted_at is null
;
$func$;

-- SLOT TAB - assigned slot
CREATE OR REPLACE FUNCTION student_course_assigned_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[])
RETURNS TABLE (
    student_subscription_id text,
    course_id text,
    student_id text,
    resource_path text,
    assigned_slot int
) LANGUAGE SQL
SECURITY INVOKER
AS $func$
select 
lss.student_subscription_id, 
lss.course_id, 
lss.student_id,
lss.resource_path,
count(lm.lesson_id)::int as "assigned_slot"
from lesson_student_subscriptions lss
inner join lesson_student_subscription_access_path lssap
on lss.student_subscription_id = lssap.student_subscription_id
inner join lesson_members lm 
on lm.user_id = lss.student_id
inner join lessons l
on l.lesson_id = lm.lesson_id and l.course_id = lss.course_id and l.center_id = lssap.location_id
where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
and lss.package_type IN ('PACKAGE_TYPE_ONE_TIME', 'PACKAGE_TYPE_SLOT_BASED')
and l.start_time::date between lss.start_at::date and lss.end_at::date
and lss.end_at > current_date
and case when keyword is not null then lower(concat(lss.student_last_name, ' ', lss.student_first_name)) like lower(concat('%',keyword,'%')) else true end
and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
and case when location_ids is not null then lssap.location_id = ANY(location_ids) else true end
and lss.deleted_at is null
and lssap.deleted_at is null
and l.deleted_at is null
and lm.deleted_at is null
group by lss.student_subscription_id, lss.course_id, lss.student_id, lss.resource_path
;
$func$;

-- SLOT TAB - purchased and assigned slots calculation
CREATE OR REPLACE FUNCTION student_course_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date)
RETURNS TABLE (
    student_id text,
    course_id text,
    location_id text,
    student_start_date date,
    student_end_date date,
    purchased_slot int,
    assigned_slot int,
    slot_gap int,
    status text,
    student_subscription_id text,
    resource_path text
) LANGUAGE SQL
AS $$
select 
scps.student_id,
scps.course_id,
scps.location_id,
scps.student_start_date, 
scps.student_end_date, 
scps.purchased_slot, 
coalesce(scas.assigned_slot, 0)::int as "assigned_slot", 
(coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int as "slot_gap",
case 
	when (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int < 0 then 'Under assigned'
	when (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int = 0 then 'Just assigned'
	when (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int > 0 then 'Over assigned'
end as "status",
scps.student_subscription_id,
scps.resource_path
from student_course_purchased_slot_fn(keyword, student_ids, course_ids, location_ids) scps
left join student_course_assigned_slot_fn(keyword, student_ids, course_ids, location_ids) scas
on scps.student_subscription_id = scas.student_subscription_id
where case when start_date is not null then scps.student_start_date >= start_date::Date else true end
and case when end_date is not null then scps.student_end_date <= end_date::Date else true end
order by scps.student_start_date asc, scps.student_end_date asc, scps.loc_updated_at asc, scps.loc_created_at asc, scps.course_id asc, scps.student_id asc
;
$$;