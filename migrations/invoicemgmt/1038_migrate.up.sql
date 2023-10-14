CREATE TABLE IF NOT EXISTS public.bulk_payment_request (
    bulk_payment_request_id TEXT NOT NULL,
    payment_method TEXT NOT NULL,
    payment_due_date_from timestamp with time zone NOT NULL,
    payment_due_date_to timestamp with time zone NOT NULL,
    total_file_count INTEGER,
    error_details TEXT,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.bulk_payment_request
    ADD CONSTRAINT bulk_payment_request_pk PRIMARY KEY (bulk_payment_request_id);

CREATE POLICY rls_bulk_payment_request ON "bulk_payment_request" using (permission_check(resource_path, 'bulk_payment_request')) 
    with check (permission_check(resource_path, 'bulk_payment_request'));

CREATE POLICY rls_bulk_payment_request_restrictive ON "bulk_payment_request" 
    AS RESTRICTIVE TO public 
    USING (permission_check(resource_path, 'bulk_payment_request'))
    WITH CHECK (permission_check(resource_path, 'bulk_payment_request'));

ALTER TABLE "bulk_payment_request" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment_request" FORCE ROW LEVEL security;


CREATE TABLE IF NOT EXISTS public.bulk_payment_request_file (
    bulk_payment_request_file_id TEXT NOT NULL,
    bulk_payment_request_id TEXT NOT NULL,
    file_name TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_sequence_number INTEGER NOT NULL,
    is_downloaded BOOLEAN NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bulk_payment_request_file_bulk_payment_request_fk FOREIGN KEY (bulk_payment_request_id) REFERENCES "bulk_payment_request"(bulk_payment_request_id)
);

ALTER TABLE ONLY public.bulk_payment_request_file
    ADD CONSTRAINT bulk_payment_request_file_pk PRIMARY KEY (bulk_payment_request_file_id);

-- Make file_sequence_number unique per bulk_payment_request_id
ALTER TABLE ONLY public.bulk_payment_request_file
    ADD CONSTRAINT bulk_payment_request_file_sequence_bulk_payment_request_unique UNIQUE (file_sequence_number,bulk_payment_request_id);

CREATE POLICY rls_bulk_payment_request_file ON "bulk_payment_request_file" using (permission_check(resource_path, 'bulk_payment_request_file')) 
    with check (permission_check(resource_path, 'bulk_payment_request_file'));

CREATE POLICY rls_bulk_payment_request_file_restrictive ON "bulk_payment_request_file" 
    AS RESTRICTIVE TO public 
    USING (permission_check(resource_path, 'bulk_payment_request_file'))
    WITH CHECK (permission_check(resource_path, 'bulk_payment_request_file'));

ALTER TABLE "bulk_payment_request_file" ENABLE ROW LEVEL security;
ALTER TABLE "bulk_payment_request_file" FORCE ROW LEVEL security;