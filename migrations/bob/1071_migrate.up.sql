INSERT INTO public.schools
(school_id, "name", country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge, phone_number, deleted_at)
VALUES(-2147483645, 'Renseikai School', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false, NULL, NULL) ON CONFLICT DO NOTHING
