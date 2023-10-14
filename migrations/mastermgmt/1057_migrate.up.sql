INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'syllabus.learning_history_sync.show_paper_count_and_perspective_mapping', 
    'boolean', 
    'false', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Enable for managara-base and managara-hs
UPDATE internal_configuration_value 
SET config_value = 'true'
WHERE config_key = 'syllabus.learning_history_sync.show_paper_count_and_perspective_mapping' and resource_path IN ('-2147483630', '-2147483629');
