INSERT INTO public.schools
(school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge)
VALUES(-2147483646, 'Synersia School', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false)
ON CONFLICT DO NOTHING;
