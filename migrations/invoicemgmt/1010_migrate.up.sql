CREATE TABLE IF NOT EXISTS public.invoice_action_log (
    invoice_action_id integer NOT NULL,
    invoice_id integer NOT NULL,
    user_id text NOT NULL,
    action text NOT NULL,
    action_detail text NOT NULL,
    action_comment text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT invoice_action_log_pk PRIMARY KEY (invoice_action_id),
    CONSTRAINT invoice_action_log_invoice_fk FOREIGN KEY (invoice_id) REFERENCES "invoice"(invoice_id),
    CONSTRAINT invoice_action_log_users_fk FOREIGN KEY (user_id) REFERENCES "users"(user_id)
);

CREATE POLICY rls_invoice_action_log ON "invoice_action_log" USING (permission_check(resource_path, 'invoice_action_log')) WITH CHECK (permission_check(resource_path, 'invoice_action_log'));

ALTER TABLE "invoice_action_log" ENABLE ROW LEVEL security;
ALTER TABLE "invoice_action_log" FORCE ROW LEVEL security;
