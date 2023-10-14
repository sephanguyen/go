#!/bin/bash

# This script migrate StudyPlanItemIdentity for student_event_logs table.

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
delete from internal_configuration_value 
where (config_key, resource_path) 
	in (
		--Local internal
		('hcm.timesheet_management',	'-2147483636'),
		('user.enrollment.update_status_manual',	'-2147483636'),
		('user.enrollment.update_status_manual',	'16091'),
		('user.enrollment.update_status_manual',	'16093'),
		('payment.order.enable_order_manager',	'-2147483636'),
		('lesson.live_lesson.enable_live_lesson', '-2147483636'),
		-- STG internal
		('lesson.live_lesson.cloud_record', '-2147483636'),
		('user.student_course.allow_input_student_course', '-2147483636'),
		('user.enrollment.update_status_manual', '16091'),
		('lesson.lessonmgmt.allow_write_lesson', '-2147483636'),
		('user.enrollment.update_status_manual', '16093'),
		('payment.order.enable_order_manager', '-2147483636'),
		('lesson.assigned_student_list', '-2147483636'),
		('lesson.lessonmgmt.zoom_selection', '-2147483636'),
		('lesson.lesson_report.enable_lesson_report', '-2147483636'),
		('syllabus.learning_material.content_lo', '-2147483636'),
		('hcm.timesheet_management', '-2147483636'),
		('user.enrollment.update_status_manual', '-2147483636'),
		--UAT internal
		('lesson.live_lesson.enable_live_lesson', '-2147483636'),
		('lesson.live_lesson.cloud_record', '-2147483636'),
		('lesson.lessonmgmt.zoom_selection', '-2147483636'),
		('lesson.lesson_report.enable_lesson_report', '-2147483636'),
		('lesson.lessonmgmt.allow_write_lesson', '-2147483636'),
		('user.student_course.allow_input_student_course', '-2147483636'),
		('syllabus.learning_material.content_lo', '-2147483636'),
		('hcm.timesheet_management', '-2147483636'),
		('lesson.assigned_student_list', '-2147483636'),
		('payment.order.enable_order_manager', '-2147483636'),
		('user.enrollment.update_status_manual', '16091'),
		('user.enrollment.update_status_manual', '16093'),
		('user.enrollment.update_status_manual', '-2147483636'),
		--Prod internal
		('lesson.live_lesson.enable_live_lesson' , '-2147483636'),
		('lesson.live_lesson.cloud_record' , '-2147483636'),
		('lesson.lessonmgmt.zoom_selection' , '-2147483636'),
		('lesson.lesson_report.enable_lesson_report' , '-2147483636'),
		('lesson.lessonmgmt.allow_write_lesson' , '-2147483636'),
		('user.student_course.allow_input_student_course' , '-2147483636'),
		('syllabus.learning_material.content_lo' , '-2147483636'),
		('hcm.timesheet_management' , '-2147483636'),
		('payment.order.enable_order_manager' , '-2147483636'),
		('lesson.assigned_student_list' , '-2147483636'),
		('user.enrollment.update_status_manual' , '16091'),
		('user.enrollment.update_status_manual' , '16093'),
		('user.enrollment.update_status_manual' , '-2147483636')
	);
delete from external_configuration_value ecv 
where (config_key, resource_path) 
	in (
		--STG external
		('user.authentication.ip_address_restriction', '-2147483636'),
		('syllabus.approve_grading', '-2147483636'),
		('general.logo', '-214748364'),
		('user.authentication.allowed_ip_address', '-2147483636'),
		('lesson.zoom.is_enabled', '2147483644'),
		--UAT external
		('user.authentication.ip_address_restriction','-2147483636'),
		('syllabus.approve_grading','-2147483636'),
		('user.authentication.allowed_ip_address','-2147483636'),
		('lesson.zoom.is_enabled','-2147483636'),
		--Prod external
		('user.authentication.ip_address_restriction', '-2147483636'),
		('user.authentication.allowed_ip_address', '-2147483636'),
		('syllabus.approve_grading', '-2147483636'),
		('lesson.zoom.is_enabled', '-2147483636'),
		('lesson.zoom.config', '-2147483636')
	);
EOF
