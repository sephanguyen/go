ALTER TABLE "associated_product" DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_associated_product ON associated_product ;
ALTER TABLE public.product_location DROP CONSTRAINT fk_location_id;

ALTER TABLE public.associated_product RENAME TO student_associated_product;
ALTER TABLE student_associated_product RENAME CONSTRAINT pk_associated_product TO pk_student_associated_product;
ALTER TABLE student_associated_product RENAME CONSTRAINT fk_associated_product_student_product_id TO fk_student_associated_product_student_product_id;
ALTER TABLE student_associated_product RENAME CONSTRAINT fk_associated_product_associated_product_id TO fk_student_associated_product_associated_product_id;

CREATE POLICY rls_student_associated_product ON "student_associated_product" using (permission_check(resource_path, 'student_associated_product')) with check (permission_check(resource_path, 'student_associated_product'));

ALTER TABLE "student_associated_product" ENABLE ROW LEVEL security;
ALTER TABLE "student_associated_product" FORCE ROW LEVEL security;