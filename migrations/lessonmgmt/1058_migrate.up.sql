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
