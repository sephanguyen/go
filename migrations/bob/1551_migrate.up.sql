ALTER TABLE IF EXISTS public.academic_week ADD COLUMN IF NOT EXISTS "week_order" SMALLINT;

ALTER TABLE IF EXISTS public.academic_week
ADD CONSTRAINT academic_week_order_academic_year_location_id_unique UNIQUE("week_order", academic_year_id, location_id);
