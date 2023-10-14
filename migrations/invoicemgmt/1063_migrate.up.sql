CREATE TABLE IF NOT EXISTS student_payment_detail (
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    payer_name text NOT NULL,
    payer_phone_number text,
    payment_method text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT student_payment_detail__pk PRIMARY KEY (student_payment_detail_id),
    CONSTRAINT student_payment_detail_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);

CREATE POLICY rls_student_payment_detail ON "student_payment_detail"
USING (permission_check(resource_path, 'student_payment_detail'))
WITH CHECK (permission_check(resource_path, 'student_payment_detail'));

CREATE POLICY rls_student_payment_detail_restrictive ON "student_payment_detail" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'student_payment_detail'))
WITH CHECK (permission_check(resource_path, 'student_payment_detail'));

ALTER TABLE "student_payment_detail" ENABLE ROW LEVEL security;
ALTER TABLE "student_payment_detail" FORCE ROW LEVEL security;