DROP FUNCTION IF EXISTS get_previous_report_of_student;

CREATE OR REPLACE FUNCTION public.get_previous_report_of_student(user_id text, report_course_id text, report_id text) 
returns setof public.lesson_reports
    language sql stable
    as $$
    select lr.* from lesson_reports lr
	join lesson_members lm on lr.lesson_id = lm.lesson_id
	join lessons l on l.lesson_id=lr.lesson_id
where
	CASE WHEN report_id IS NOT NULL 
        THEN l.start_time < (
	            select l1.start_time 
                    from lessons l1 join lesson_reports lr1 on l1.lesson_id=lr1.lesson_id
                    where lr1.lesson_report_id = report_id limit 1)
        ELSE l.start_time <= now()
    END
	and lm.user_id = user_id
	and lm.course_id = report_course_id
order by
	l.start_time desc
limit 1;
$$;

DROP FUNCTION IF EXISTS get_partner_dynamic_form_field_values_by_student;
CREATE OR REPLACE FUNCTION public.get_partner_dynamic_form_field_values_by_student(user_id text, report_id text)
returns setof public.partner_dynamic_form_field_values
    language sql stable as $$
select * from
	partner_dynamic_form_field_values
where
	lesson_report_detail_id = (
	select
		lrd.lesson_report_detail_id
	from
		lesson_report_details lrd
	where
		lrd.lesson_report_id = report_id
		and lrd.student_id = user_id
	limit 1);
$$;
