INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'calendar.weekly_view.integrated_academic_calendar.is_enabled', 
    'string', 
    'off', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
