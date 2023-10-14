INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'mastermgmt.academic_calendar.location_type_level', 
    'number', 
    '1', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

ALTER TABLE IF EXISTS public.academic_week ADD COLUMN IF NOT EXISTS "week_order" SMALLINT;

ALTER TABLE IF EXISTS public.academic_week
ADD CONSTRAINT academic_week_order_academic_year_location_id_unique UNIQUE("week_order", academic_year_id, location_id);

ALTER TABLE IF EXISTS public.academic_closed_day ADD CONSTRAINT academic_closed_day_date_academic_year_location_id_unique UNIQUE ("date", academic_year_id, location_id);
