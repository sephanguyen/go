INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'syllabus.resume_video.is_enabled', 
    'boolean', 
    'false', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
