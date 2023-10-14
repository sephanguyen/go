DROP POLICY IF EXISTS rls_order on "order";
CREATE POLICY rls_order_read_all ON "order" AS PERMISSIVE FOR select TO PUBLIC
using (1 = 1);
