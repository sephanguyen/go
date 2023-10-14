CREATE OR REPLACE VIEW public.student_course_slot_info AS
select 
slot_info.student_id,
slot_info.course_id,
slot_info.location_id,
slot_info.student_start_date, 
slot_info.student_end_date, 
slot_info.purchased_slot, 
coalesce(assigned_slot_info.assigned_slot, 0) as "assigned_slot", 
coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot as "slot_gap",
case 
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot < 0 then 'Under assigned'
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot = 0 then 'Just assigned'
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot > 0 then 'Over assigned'
end as "status",
slot_info.student_subscription_id
from (select 
	lss.student_subscription_id, 
	lss.student_id,
	lss.course_id, 
	lssap.location_id, 
	lss.start_at::DATE as "student_start_date",  
	lss.end_at::DATE as "student_end_date", 
	lss.course_slot as "purchased_slot", 
	loc.created_at as "loc_created_at",
	loc.updated_at as "loc_updated_at"
	from lesson_student_subscriptions lss
	inner join lesson_student_subscription_access_path lssap
	on lss.student_subscription_id = lssap.student_subscription_id
	inner join locations loc
	on loc.location_id = lssap.location_id
	where lss.course_slot is not null
	and lss.end_at > current_date
	and lss.deleted_at is null
	and lssap.deleted_at is null
	and loc.deleted_at is null) slot_info
	left join (select 
		lss.student_subscription_id, 
		lss.course_id, 
		lss.student_id, 
		count(lm.lesson_id) as "assigned_slot"
		from lesson_student_subscriptions lss 
		inner join lesson_members lm 
		on lm.user_id = lss.student_id
		inner join lessons l
		on l.lesson_id = lm.lesson_id and l.course_id = lss.course_id
		where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
		and lss.course_slot is not null
		and lss.end_at > current_date
		and lss.deleted_at is null
		and l.deleted_at is null
		and lm.deleted_at is null
		group by lss.student_subscription_id, lss.course_id, lss.student_id) assigned_slot_info
on slot_info.student_subscription_id = assigned_slot_info.student_subscription_id
order by slot_info.student_start_date asc, slot_info.student_end_date asc, slot_info.loc_updated_at asc, slot_info.loc_created_at asc, slot_info.course_id asc, slot_info.student_id asc
;

CREATE OR REPLACE VIEW public.student_course_recurring_slot_info AS
select 
slot_info.student_id,
slot_info.course_id,
slot_info.location_id,
slot_info.week_start,
(slot_info.week_start + interval '6 days')::date as "week_end",
slot_info.purchased_slot, 
coalesce(assigned_slot_info.assigned_slot, 0) as "assigned_slot", 
coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot as "slot_gap",
case 
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot < 0 then 'Under assigned'
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot = 0 then 'Just assigned'
	when coalesce(assigned_slot_info.assigned_slot, 0)-slot_info.purchased_slot > 0 then 'Over assigned'
end as "status",
slot_info.student_subscription_id
from (select 
	lss.student_subscription_id, 
	lss.student_id,
	lss.course_id, 
	lssap.location_id, 
	generate_series(
		to_date(to_char(lss.start_at, 'IYYY-IW'), 'IYYY-IW')
		,to_date(to_char(lss.end_at, 'IYYY-IW'), 'IYYY-IW')
		,interval '1 week'
	)::date  as "week_start", 
	lss.course_slot_per_week as "purchased_slot", 
	loc.created_at as "loc_created_at",
	loc.updated_at as "loc_updated_at"
	from lesson_student_subscriptions lss
	inner join lesson_student_subscription_access_path lssap
	on lss.student_subscription_id = lssap.student_subscription_id
	inner join locations loc
	on loc.location_id = lssap.location_id
	where lss.course_slot_per_week is not null
	and lss.end_at > current_date
	and lss.deleted_at is null
    and lssap.deleted_at is null
	and loc.deleted_at is null) slot_info
	left join (select 
		lss.student_subscription_id, 
		lss.course_id, 
		lss.student_id,
		to_char(l.start_time, 'IYYY-IW') as "week",
		count(lm.lesson_id) as "assigned_slot"
		from lesson_student_subscriptions lss 
		inner join lesson_members lm 
		on lm.user_id = lss.student_id
		inner join lessons l
		on l.lesson_id = lm.lesson_id and l.course_id = lss.course_id
		where l.scheduling_status != 'LESSON_SCHEDULING_STATUS_CANCELED'
		and lss.course_slot_per_week is not null
		and lss.end_at > current_date
		and lss.deleted_at is null
		and l.deleted_at is null
		and lm.deleted_at is null
		group by lss.student_subscription_id, lss.course_id, lss.student_id, to_char(l.start_time, 'IYYY-IW')) assigned_slot_info
on slot_info.student_subscription_id = assigned_slot_info.student_subscription_id
and to_char(slot_info.week_start, 'IYYY-IW') = assigned_slot_info.week
order by slot_info.week_start asc, "week_end" asc, slot_info.loc_updated_at asc, slot_info.loc_created_at asc, slot_info.course_id asc, slot_info.student_id asc
;
