-- delete wrong config value 
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

-- add Fk to organizations table
ALTER TABLE public.internal_configuration_value DROP CONSTRAINT IF EXISTS fk_config_value_internal_org_id;
ALTER TABLE public.internal_configuration_value ADD CONSTRAINT fk_config_value_internal_org_id FOREIGN KEY (resource_path) REFERENCES public.organizations(organization_id);

ALTER TABLE public.external_configuration_value DROP CONSTRAINT IF EXISTS external_configuration_value;
ALTER TABLE public.external_configuration_value ADD CONSTRAINT fk_config_value_external_org_id FOREIGN KEY (resource_path) REFERENCES public.organizations(organization_id);

-- replace Rule system by Trigger for org creation
drop rule if exists INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_PARTNER on organizations;
drop rule if exists INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_PARTNER on organizations;

CREATE OR REPLACE FUNCTION generate_config_value_for_new_organization() RETURNS TRIGGER
AS $$
    BEGIN
        INSERT INTO external_configuration_value (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)  
            SELECT uuid_generate_v4() AS uuid_generate_v4,
                ck.config_key,
                ck.value_type,
                now() AS now,
                now() AS now,
                ck.default_value,
                new.resource_path
            FROM configuration_key ck
            WHERE (ck.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL'::text);
        INSERT INTO internal_configuration_value (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)  
            SELECT uuid_generate_v4() AS uuid_generate_v4,
                ck.config_key,
                ck.value_type,
                now() AS now,
                now() AS now,
                ck.default_value,
                new.resource_path
            FROM configuration_key ck
            WHERE (ck.configuration_type = 'CONFIGURATION_TYPE_INTERNAL'::text);
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS init_config_value_for_new_partner on public.organizations;
CREATE TRIGGER init_config_value_for_new_partner AFTER INSERT ON public.organizations FOR EACH ROW EXECUTE PROCEDURE generate_config_value_for_new_organization();

-- replace Rule system by Trigger for config_key creation
drop rule if exists INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_KEY on configuration_key;
drop rule if exists INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_KEY on configuration_key;

CREATE OR REPLACE FUNCTION generate_config_value_for_new_config_key() RETURNS TRIGGER
AS $$
    BEGIN
	INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
		select uuid_generate_v4(), new.config_key, new.value_type, now(), now(), new.default_value, resource_path
		from organizations o 
		where new.configuration_type = 'CONFIGURATION_TYPE_INTERNAL';
	INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
		select uuid_generate_v4(), new.config_key, new.value_type, now(), now(), new.default_value, resource_path
		from organizations o 
		where new.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL';
    RETURN NEW;
END $$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS init_config_value_for_new_key on public.configuration_key;
CREATE TRIGGER init_config_value_for_new_key AFTER INSERT ON public.configuration_key FOR EACH ROW EXECUTE PROCEDURE generate_config_value_for_new_config_key();

-- fix missing config value
INSERT INTO public.internal_configuration_value (configuration_id, created_at, updated_at, config_key, config_value_type, config_value, resource_path) 
select uuid_generate_v4(), now(), now(),
		ck.config_key, ck.value_type, ck.default_value,
		o.resource_path
from organizations o 
	cross join configuration_key ck
where ck.configuration_type = 'CONFIGURATION_TYPE_INTERNAL'
on conflict do nothing;

INSERT INTO public.external_configuration_value (configuration_id, created_at, updated_at, config_key, config_value_type, config_value, resource_path) 
select	uuid_generate_v4(), now(), now(),ck.config_key, ck.value_type, ck.default_value,
		o.resource_path
from organizations o 
	cross join configuration_key ck
where ck.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL'
on conflict do nothing;