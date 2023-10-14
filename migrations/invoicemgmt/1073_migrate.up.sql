CREATE TABLE IF NOT EXISTS public.order (
	order_id text NOT NULL,
	student_id text NOT NULL,
	location_id text NOT NULL,
	order_sequence_number integer NOT NULL,
	order_comment text NULL,
	order_status text NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	order_type text NULL,
	student_full_name text NOT NULL,
	is_reviewed bool NULL DEFAULT false,
	withdrawal_effective_date timestamptz NULL
);

ALTER TABLE ONLY public.order
    ADD CONSTRAINT order_pk PRIMARY KEY (order_id);

ALTER TABLE ONLY public.order
	ADD CONSTRAINT order_sequence_number_resource_path_unique UNIQUE (order_sequence_number, resource_path);


CREATE POLICY rls_order ON "order" 
    using (permission_check(resource_path, 'order')) 
    with check (permission_check(resource_path, 'order'));

CREATE POLICY rls_order_restrictive ON "order" 
    AS RESTRICTIVE TO public 
    USING (permission_check(resource_path, 'order'))
    WITH CHECK (permission_check(resource_path, 'order'));


ALTER TABLE "order" ENABLE ROW LEVEL security;
ALTER TABLE "order" FORCE ROW LEVEL security;