INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'architecture.course.bulk_class_allocation', 
    'string', 
    'false', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
