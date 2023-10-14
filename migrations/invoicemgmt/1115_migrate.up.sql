CREATE TABLE IF NOT EXISTS public.student_payment_detail_action_log (
    student_payment_detail_action_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    user_id text NOT NULL,
    action text NOT NULL,
    action_detail text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT student_payment_detail_action_log__pk PRIMARY KEY (student_payment_detail_action_id),
    CONSTRAINT student_payment_detail_action_log__student_payment_detail__fk FOREIGN KEY (student_payment_detail_id) REFERENCES "student_payment_detail"(student_payment_detail_id),
    CONSTRAINT student_payment_detail_action_log__users__fk FOREIGN KEY (user_id) REFERENCES "users"(user_id)
);

CREATE POLICY rls_student_payment_detail_action_log ON "student_payment_detail_action_log" 
USING (permission_check(resource_path, 'student_payment_detail_action_log')) WITH CHECK (permission_check(resource_path, 'student_payment_detail_action_log'));

CREATE POLICY rls_student_payment_detail_action_log_restrictive ON "student_payment_detail_action_log" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'student_payment_detail_action_log')) WITH CHECK (permission_check(resource_path, 'student_payment_detail_action_log'));


ALTER TABLE "student_payment_detail_action_log" ENABLE ROW LEVEL security;
ALTER TABLE "student_payment_detail_action_log" FORCE ROW LEVEL security;
