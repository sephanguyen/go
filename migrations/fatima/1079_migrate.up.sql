ALTER TABLE ONLY public.student_package_by_order
    ADD CONSTRAINT student_id_package_id_unique UNIQUE (student_id,package_id);
