INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'arch.master_management.enable_import', 
    'string', 
    '["class","location","locationType","course","courseType","grades","grade","subject","courseAccessPath","academicCalendar","workingHour","timeSlot"]', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
