ALTER TABLE "product_course" DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_product_course ON product_course ;
ALTER TABLE public.product_course RENAME CONSTRAINT fk_product_course_id TO fk_package_course_id;
ALTER TABLE public.product_course RENAME CONSTRAINT product_course_pk TO package_course_pk;
ALTER TABLE public.product_course RENAME COLUMN product_id TO package_id;


ALTER TABLE public.product_course RENAME TO package_course;

ALTER TABLE package_course ADD COLUMN max_slot integer NOT NULL DEFAULT 1;

CREATE POLICY rls_package_course ON "package_course" using (permission_check(resource_path, 'package_course')) with check (permission_check(resource_path, 'package_course'));

ALTER TABLE "package_course" ENABLE ROW LEVEL security;
ALTER TABLE "package_course" FORCE ROW LEVEL security;