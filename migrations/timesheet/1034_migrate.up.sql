CREATE OR REPLACE VIEW "public"."location_timesheets_non_confirmed_count" AS 
 SELECT loc.location_id,
    ( SELECT count(t2.timesheet_id) AS count
           FROM timesheet t2
          WHERE ((t2.location_id = loc.location_id) AND t2.deleted_at IS NULL AND (t2.timesheet_status = ANY (ARRAY['TIMESHEET_STATUS_DRAFT'::text, 'TIMESHEET_STATUS_SUBMITTED'::text, 'TIMESHEET_STATUS_APPROVED'::text])))) AS count,
    loc.deleted_at,
    loc.resource_path
   FROM (timesheet t
     JOIN locations loc ON ((t.location_id = loc.location_id)))
  GROUP BY loc.location_id;
