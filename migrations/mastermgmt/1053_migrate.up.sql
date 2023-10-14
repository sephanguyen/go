-- Turn ON for KEC & KEC-Demo School first, until release date
UPDATE internal_configuration_value 
SET config_value = 'true'
WHERE config_key = 'user.student_management.show_add_import_student_button' and resource_path IN ('-2147483635', '-2147483642');
