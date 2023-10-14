
DROP POLICY IF EXISTS rls_bank_account_location on "bank_account";
DROP POLICY IF EXISTS rls_billing_address_location on "billing_address";
DROP POLICY IF EXISTS rls_student_payment_detail_location on "student_payment_detail";

DROP POLICY IF EXISTS rls_bank_account on "bank_account";
DROP POLICY IF EXISTS rls_billing_address on "billing_address";
DROP POLICY IF EXISTS rls_student_payment_detail on "student_payment_detail";

CREATE POLICY rls_billing_address ON "billing_address"
USING (permission_check(resource_path, 'billing_address'))
WITH CHECK (permission_check(resource_path, 'billing_address'));

CREATE POLICY rls_bank_account ON "bank_account"
USING (permission_check(resource_path, 'bank_account'))
WITH CHECK (permission_check(resource_path, 'bank_account'));

CREATE POLICY rls_student_payment_detail ON "student_payment_detail"
USING (permission_check(resource_path, 'student_payment_detail'))
WITH CHECK (permission_check(resource_path, 'student_payment_detail'));

ALTER TABLE ONLY public.billing_address
  DROP CONSTRAINT IF EXISTS billing_address_user_basic_info__fk;

ALTER TABLE ONLY public.bank_account
  DROP CONSTRAINT IF EXISTS bank_account_students_fk;

ALTER TABLE ONLY public.student_payment_detail
  DROP CONSTRAINT IF EXISTS student_payment_detail_students_fk,
  ADD CONSTRAINT student_payment_detail_student_id_unique UNIQUE (student_id);
