INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'lesson.lessonmgmt.new_student_tag', 
    'number', 
    '30', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);