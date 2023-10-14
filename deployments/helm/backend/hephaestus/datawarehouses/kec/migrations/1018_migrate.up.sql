DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='kec_publication') THEN
      CREATE PUBLICATION kec_publication;
   END IF;
END
$do$;

ALTER PUBLICATION kec_publication SET TABLE 
bob.parents_public_info,
bob.staff_public_info,
bob.students_public_info,
bob.student_parents_public_info,
bob.scheduler_public_info,
bob.lessons_teachers_public_info,
bob.lessons_courses_public_info,
bob.reallocation_public_info,
bob.role_public_info,
bob.permission_public_info,
bob.user_group_public_info,
bob.user_group_member_public_info,
bob.student_enrollment_status_history_public_info,
bob.classroom_public_info,
bob.lesson_reports_public_info,
bob.partner_form_configs_public_info,
bob.school_level_public_info,
bob.school_course_school_info_public_info,
bob.school_history_public_info,
bob.tagged_user_public_info,
bob.user_phone_number_public_info,
bob.user_address_public_info,
bob.partner_dynamic_form_field_values_public_info,
bob.day_info_public_info,
bob.day_type_public_info,
bob.lesson_members_public_info,
invoicemgmt.invoice_public_info,
timesheet.ts_transportation,
timesheet.staff_transportation_expense,
timesheet.ts_lesson,
timesheet.ts_other_working,
timesheet.auto_create_flag_activity_log,
timesheet.auto_create_timesheet_flag,
timesheet.timesheet_confirmation_cut_off_date,
timesheet.timesheet_confirmation_info,
timesheet.timesheet_confirmation_period,
invoicemgmt.invoice_payment_list_public_info,
invoicemgmt.invoice_bill_item_list_public_info,
invoicemgmt.student_payment_detail_history_info
;
