CREATE TABLE IF NOT EXISTS bank_account (
    bank_account_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    is_verified BOOLEAN DEFAULT false,
    bank_branch_id text NOT NULL,
    bank_account_number text NOT NULL,
    bank_account_holder text NOT NULL,
    bank_account_type text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bank_account__pk PRIMARY KEY (bank_account_id),
    CONSTRAINT bank_account_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id),
    CONSTRAINT bank_account_student_payment_detail_fk FOREIGN KEY (student_payment_detail_id) REFERENCES "student_payment_detail"(student_payment_detail_id)
);

CREATE POLICY rls_bank_account ON "bank_account"
USING (permission_check(resource_path, 'bank_account'))
WITH CHECK (permission_check(resource_path, 'bank_account'));

CREATE POLICY rls_bank_account_restrictive ON "bank_account" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bank_account'))
WITH CHECK (permission_check(resource_path, 'bank_account'));

ALTER TABLE "bank_account" ENABLE ROW LEVEL security;
ALTER TABLE "bank_account" FORCE ROW LEVEL security;