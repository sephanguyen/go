update academic_closed_day acd set academic_week_id = aw.academic_week_id 
from academic_week aw
where aw.start_date <= acd."date" 
and aw.end_date >= acd."date" 
and aw.academic_year_id = acd.academic_year_id
and aw.location_id = acd.location_id
and aw.resource_path = acd.resource_path 
-- Fix academic_closed_days are missing academic_week_id
