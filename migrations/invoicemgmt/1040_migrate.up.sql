CREATE TABLE IF NOT EXISTS public.new_customer_code_history (
    new_customer_code_history_id text NOT NULL,
    new_customer_code text NOT NULL,
    student_id text NOT NULL,
    account_number integer NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT new_customer_code_history__pk PRIMARY KEY (new_customer_code_history_id),
    CONSTRAINT new_customer_code_history_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);

CREATE POLICY rls_new_customer_code_history ON "new_customer_code_history"
USING (permission_check(resource_path, 'new_customer_code_history'))
WITH CHECK (permission_check(resource_path, 'new_customer_code_history'));

CREATE POLICY rls_new_customer_code_history_restrictive ON "new_customer_code_history" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'new_customer_code_history'))
WITH CHECK (permission_check(resource_path, 'new_customer_code_history'));

ALTER TABLE "new_customer_code_history" ENABLE ROW LEVEL security;
ALTER TABLE "new_customer_code_history" FORCE ROW LEVEL security;