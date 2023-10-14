ALTER TABLE ONLY public.student_packages
    ADD COLUMN IF NOT EXISTS location_ids text[];
