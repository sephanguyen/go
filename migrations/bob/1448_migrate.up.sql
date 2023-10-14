ALTER TABLE public.package_course ALTER COLUMN max_slots_per_course TYPE INTEGER USING(max_slots_per_course::INTEGER);
