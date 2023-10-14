DROP POLICY IF EXISTS rls_student_product on "student_product";
CREATE POLICY rls_student_product_read_all ON "student_product" AS PERMISSIVE FOR select TO PUBLIC
using (1 = 1);
