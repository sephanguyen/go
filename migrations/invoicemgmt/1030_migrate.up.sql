CREATE TABLE public.invoice_schedule_student (
    invoice_schedule_student_id text NOT NULL,
    invoice_schedule_history_id text NOT NULL,
    student_id text NOT NULL,
    error_details text NOT NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_schedule_student_pk PRIMARY KEY (invoice_schedule_student_id),
    CONSTRAINT invoice_schedule_student_invoice_schedule_history_fk FOREIGN KEY (invoice_schedule_history_id) REFERENCES "invoice_schedule_history"(invoice_schedule_history_id),
    CONSTRAINT invoice_schedule_student_students_fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);

CREATE POLICY rls_invoice_schedule_student ON "invoice_schedule_student" USING (permission_check(resource_path, 'invoice_schedule_student')) WITH CHECK (permission_check(resource_path, 'invoice_schedule_student'));

ALTER TABLE "invoice_schedule_student" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_schedule_student" FORCE ROW LEVEL security; 