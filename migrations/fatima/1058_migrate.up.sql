CREATE TABLE public.package_course_material (
    package_id integer NOT NULL,
    course_id text NOT NULL,
    material_id integer NOT NULL,
    available_from timestamp with time zone,
    available_until timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.package_course_material
    ADD CONSTRAINT package_course_material_pk PRIMARY KEY (package_id, course_id, material_id);

ALTER TABLE public.package_course_material ADD CONSTRAINT fk_package_id FOREIGN KEY(package_id) REFERENCES public.package(package_id);
ALTER TABLE public.package_course_material ADD CONSTRAINT fk_course_id FOREIGN KEY(course_id) REFERENCES public.courses(course_id);
ALTER TABLE public.package_course_material ADD CONSTRAINT fk_material_id FOREIGN KEY(material_id) REFERENCES public.material(material_id);

CREATE POLICY rls_package_course_material ON "package_course_material" using (permission_check(resource_path, 'package_course_material')) with check (permission_check(resource_path, 'package_course_material'));

ALTER TABLE "package_course_material" ENABLE ROW LEVEL security;
ALTER TABLE "package_course_material" FORCE ROW LEVEL security;

CREATE TABLE public.package_course_fee (
    package_id integer NOT NULL,
    course_id text NOT NULL,
    fee_id integer NOT NULL,
    available_from timestamp with time zone,
    available_until timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.package_course_fee
    ADD CONSTRAINT package_course_fee_pk PRIMARY KEY (package_id, course_id, fee_id);

ALTER TABLE public.package_course_fee ADD CONSTRAINT fk_package_id FOREIGN KEY(package_id) REFERENCES public.package(package_id);
ALTER TABLE public.package_course_fee ADD CONSTRAINT fk_course_id FOREIGN KEY(course_id) REFERENCES public.courses(course_id);
ALTER TABLE public.package_course_fee ADD CONSTRAINT fk_fee_id FOREIGN KEY(fee_id) REFERENCES public.fee(fee_id);

CREATE POLICY rls_package_course_fee ON "package_course_fee" using (permission_check(resource_path, 'package_course_fee')) with check (permission_check(resource_path, 'package_course_fee'));

ALTER TABLE "package_course_fee" ENABLE ROW LEVEL security;
ALTER TABLE "package_course_fee" FORCE ROW LEVEL security;
