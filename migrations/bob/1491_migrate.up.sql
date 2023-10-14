-- remove old functions
DROP FUNCTION IF EXISTS public.student_course_recurring_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], timezone text, unique_id text, use_basic_info boolean);
DROP FUNCTION IF EXISTS student_course_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date, timezone text, unique_id text, use_basic_info boolean);

-- RECURRING TAB - purchased slot
CREATE OR REPLACE FUNCTION student_course_recurring_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date, timezone text DEFAULT 'UTC', unique_id text DEFAULT NULL, use_basic_info boolean DEFAULT false)
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
    unique_id text
) LANGUAGE SQL
STABLE
PARALLEL SAFE
AS $$
SELECT  scprs.student_id, scprs.course_id, scprs.location_id, scprs.week_start, scprs.week_end, scprs.purchased_slot, 
  coalesce(scars.assigned_slot, 0)::int AS "assigned_slot", (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int AS "slot_gap",
  CASE 
    WHEN (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int < 0 THEN 'Under assigned'
    WHEN (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int = 0 THEN 'Just assigned'
    WHEN (coalesce(scars.assigned_slot, 0)-scprs.purchased_slot)::int > 0 THEN 'Over assigned'
  END AS "status", scprs.unique_id
FROM student_course_purchased_recurring_slot_fn(keyword, student_ids, course_ids, location_ids, timezone, unique_id, use_basic_info) scprs
LEFT JOIN student_course_assigned_recurring_slot_fn(keyword, student_ids, course_ids, location_ids, timezone, unique_id, use_basic_info) scars
  ON scprs.unique_id = scars.unique_id
  AND to_char(scprs.week_start, 'IYYY-IW') = scars.week
WHERE 
  CASE WHEN start_date IS NOT NULL AND end_date IS NOT NULL
		THEN scprs.week_start <= start_date AND end_date <= scprs.week_end
		ELSE
			CASE WHEN start_date IS NOT NULL THEN scprs.week_start >= start_date::Date ELSE true END
			AND CASE WHEN end_date IS NOT NULL THEN scprs.week_end <= end_date::Date ELSE true END 
	END
ORDER BY scprs.week_start ASC, scprs.week_end ASC, scprs.loc_updated_at ASC, scprs.loc_created_at ASC, scprs.course_id ASC, scprs.student_id ASC
; $$;

-- SLOT TAB - purchased AND assigned slots calculation
CREATE OR REPLACE FUNCTION student_course_slot_info_fn(keyword text, student_ids text[], course_ids text[], location_ids text[], start_date date, end_date date, timezone text DEFAULT 'UTC', unique_id text DEFAULT NULL, use_basic_info boolean DEFAULT false)
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
    unique_id text
) LANGUAGE SQL
STABLE
PARALLEL SAFE
AS $$
SELECT scps.student_id, scps.course_id, scps.location_id, scps.student_start_date, scps.student_end_date, scps.purchased_slot, 
  coalesce(scas.assigned_slot, 0)::int AS "assigned_slot", (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int AS "slot_gap",
  CASE 
    WHEN (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int < 0 THEN 'Under assigned'
    WHEN (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int = 0 THEN 'Just assigned'
    WHEN (coalesce(scas.assigned_slot, 0)-scps.purchased_slot)::int > 0 THEN 'Over assigned'
  END AS "status", scps.unique_id
FROM student_course_purchased_slot_fn(keyword, student_ids, course_ids, location_ids, timezone, unique_id, use_basic_info) scps
LEFT JOIN student_course_assigned_slot_fn(keyword, student_ids, course_ids, location_ids, timezone, unique_id, use_basic_info) scas
  ON scps.unique_id = scas.unique_id
WHERE
  CASE WHEN start_date IS NOT NULL AND end_date IS NOT NULL
		THEN scps.student_start_date  <= start_date AND end_date <= scps.student_end_date
		ELSE
			CASE WHEN start_date IS NOT NULL THEN scps.student_start_date >= start_date::Date ELSE true END
      AND CASE WHEN end_date IS NOT NULL THEN scps.student_end_date <= end_date::Date ELSE true END
	END
ORDER BY scps.student_start_date ASC, scps.student_end_date ASC, scps.loc_updated_at ASC, scps.loc_created_at ASC, scps.course_id ASC, scps.student_id ASC
;
$$;
