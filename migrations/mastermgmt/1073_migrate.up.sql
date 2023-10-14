------------ KEC Test -------------
--- Organization ---
INSERT INTO organizations (organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at)
VALUES   ('-2147483623', 'kec-testing-grqf6', 'KEC', '-2147483623', 'kec-test', '', 'COUNTRY_JP', now(), now(), null) ON CONFLICT DO NOTHING;

UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = any('{user.student_course.allow_input_student_course,user.enrollment.update_status_manual}')
AND resource_path = '-2147483623';