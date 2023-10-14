INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'user.student_management.show_add_import_student_button', 
    'boolean', 
    'true', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Hide for KEC & KEC-Demo School 
UPDATE internal_configuration_value 
SET config_value = 'false'
WHERE config_key = 'user.student_management.show_add_import_student_button' and resource_path IN ('-2147483635', '-2147483642');
