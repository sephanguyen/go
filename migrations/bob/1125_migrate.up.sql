
update lessons l set resource_path = c.school_id::text
from courses c
where l.course_id = c.course_id
	and(l.resource_path is null or length(l.resource_path)=0);
	
update lesson_members lm set resource_path = l.resource_path
from lessons l
where l.lesson_id = lm.lesson_id
	and(lm.resource_path is null or length(lm.resource_path)=0);
	
update lessons_teachers lt set resource_path = l.resource_path
from lessons l
where l.lesson_id = lt.lesson_id 
	and(lt.resource_path is null or length(lt.resource_path)=0);
	
update lessons_courses lc set resource_path = c.school_id::text
from courses c
where lc.course_id = c.course_id
	and(lc.resource_path is null or length(lc.resource_path)=0);

update lesson_groups lg set resource_path = c.school_id::text
from courses c
where lg.course_id = c.course_id
	and(lg.resource_path is null or length(lg.resource_path)=0);

update lesson_members_states lms set resource_path = l.resource_path::text
from lessons l
where lms.lesson_id = l.lesson_id 
	and(lms.resource_path is null or length(lms.resource_path)=0);

update lesson_polls lp set resource_path = l.resource_path::text
from lessons l
where lp.lesson_id = l.lesson_id 
	and(lp.resource_path is null or length(lp.resource_path)=0);

update lesson_reports lr set resource_path = l.resource_path::text
from lessons l
where lr.lesson_id = l.lesson_id 
	and(lr.resource_path is null or length(lr.resource_path)=0);

update lesson_report_details lrd set resource_path = lr.resource_path::text
from lesson_reports lr
where lrd.lesson_report_id = lr.lesson_report_id 
	and(lrd.resource_path is null or length(lrd.resource_path)=0);

update partner_form_configs set resource_path = partner_id::text;

update partner_dynamic_form_field_values pdffv  set resource_path = lrd.resource_path::text
from lesson_report_details lrd
where pdffv.lesson_report_detail_id = lrd.lesson_report_detail_id 
	and(pdffv.resource_path is null or length(pdffv.resource_path)=0);
