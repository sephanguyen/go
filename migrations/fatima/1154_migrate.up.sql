DROP POLICY IF EXISTS rls_bill_item on "bill_item";
CREATE POLICY rls_bill_item_read_all ON "bill_item" AS PERMISSIVE FOR select TO PUBLIC
using (1 = 1);
