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
    and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
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
        and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
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
    and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
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
and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
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

-- RECURRING TAB - purchased slot
CREATE OR REPLACE FUNCTION student_course_purchased_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[])
RETURNS TABLE (
    student_subscription_id text,
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
base_slot_info.student_subscription_id, 
base_slot_info.student_id,
base_slot_info.course_id, 
lssap.location_id, 
base_slot_info.week_start,
(base_slot_info.week_start + interval '6 days')::date as "week_end",
base_slot_info.purchased_slot, 
loc.created_at as "loc_created_at",
loc.updated_at as "loc_updated_at",
base_slot_info.resource_path
from (select 
    lss.student_subscription_id, 
    lss.student_id,
    lss.course_id, 
    generate_series(to_date(to_char(lss.start_at, 'IYYY-IW'), 'IYYY-IW')
        ,to_date(to_char(lss.end_at, 'IYYY-IW'), 'IYYY-IW')
        ,interval '1 week'
    )::date  as "week_start", 
    lss.course_slot_per_week as "purchased_slot",
    lss.resource_path
    from lesson_student_subscriptions lss
    where lss.end_at > current_date
    and lss.package_type = 'PACKAGE_TYPE_FREQUENCY'
    and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
    and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
    and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
    and lss.deleted_at is null
    UNION
    select 
    student_info.student_subscription_id, 
    student_info.student_id,
    student_info.course_id, 
    student_info.week_start,
    coalesce(class_info.lesson_count, 0)::int as "purchased_slot",
    student_info.resource_path
    from (select 
        lss.student_subscription_id, 
        lss.student_id,
        lss.course_id, 
        generate_series(to_date(to_char(lss.start_at, 'IYYY-IW'), 'IYYY-IW')
            ,to_date(to_char(lss.end_at, 'IYYY-IW'), 'IYYY-IW')
            ,interval '1 week'
        )::date as "week_start",
        lss.resource_path
        from lesson_student_subscriptions lss
        where lss.end_at > current_date
        and lss.package_type = 'PACKAGE_TYPE_SCHEDULED'
        and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
        and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
        and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
        and lss.deleted_at is null) student_info
    left join (select 
        lss.student_subscription_id,
        cl.course_id,
        clm.user_id,
        to_char(l.start_time, 'IYYY-IW') as "week",
        count(l.lesson_id) as "lesson_count"
        from lesson_student_subscriptions lss
        inner join class cl
        on cl.course_id = lss.course_id
        inner join class_member clm
        on cl.class_id = clm.class_id and lss.student_id = clm.user_id
        inner join lessons l
        on l.class_id = cl.class_id and l.course_id = lss.course_id and l.center_id = cl.location_id
        where lss.end_at > current_date
        and lss.package_type = 'PACKAGE_TYPE_SCHEDULED'
        and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
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
        group by lss.student_subscription_id, cl.course_id, clm.user_id, to_char(l.start_time, 'IYYY-IW')
    ) class_info
    on class_info.student_subscription_id = student_info.student_subscription_id 
    and class_info.week = to_char(student_info.week_start, 'IYYY-IW')
    and class_info.course_id = student_info.course_id
    and class_info.user_id = student_info.student_id) base_slot_info
inner join lesson_student_subscription_access_path lssap
on base_slot_info.student_subscription_id = lssap.student_subscription_id
inner join locations loc
on loc.location_id = lssap.location_id
where case when location_ids is not null then lssap.location_id = ANY(location_ids) else true end
and lssap.deleted_at is null
and loc.deleted_at is null
;
$func$;

-- RECURRING TAB - assigned slot
CREATE OR REPLACE FUNCTION student_course_assigned_recurring_slot_fn(keyword text, student_ids text[], course_ids text[], location_ids text[])
RETURNS TABLE (
    student_subscription_id text,
    course_id text,
    student_id text,
    week text,
    resource_path text,
    assigned_slot int
) LANGUAGE SQL
SECURITY INVOKER
AS $func$
select 
lss.student_subscription_id, 
lss.course_id, 
lss.student_id,
to_char(l.start_time, 'IYYY-IW') as "week",
lss.resource_path,
count(lm.lesson_id)::int as "assigned_slot"
from lesson_student_subscriptions lss 
inner join lesson_student_subscription_access_path lssap
on lss.student_subscription_id = lssap.student_subscription_id
inner join lesson_members lm 
on lm.user_id = lss.student_id
inner join lessons l
on l.lesson_id = lm.lesson_id and l.course_id = lss.course_id
where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
and lss.package_type IN ('PACKAGE_TYPE_FREQUENCY', 'PACKAGE_TYPE_SCHEDULED')
and l.start_time::date between lss.start_at::date and lss.end_at::date
and lss.end_at > current_date
and case when keyword is not null then exists (select 1 from users u where u.user_id = lss.student_id and u.deleted_at is null and lower(u.name) like lower(concat('%',keyword,'%'))) else true end
and case when student_ids is not null then lss.student_id = ANY(student_ids) else true end
and case when course_ids is not null then lss.course_id = ANY(course_ids) else true end
and case when location_ids is not null then lssap.location_id = ANY(location_ids) else true end
and lss.deleted_at is null
and l.deleted_at is null
and lm.deleted_at is null
group by lss.student_subscription_id, lss.course_id, lss.student_id, to_char(l.start_time, 'IYYY-IW'), lss.resource_path
;
$func$;