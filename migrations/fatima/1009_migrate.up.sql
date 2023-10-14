CREATE TABLE public.billing_schedule (
    id integer NOT NULL,
    name text NOT NULL,
    remarks text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.billing_schedule
    ADD CONSTRAINT billing_schedule_pk PRIMARY KEY (id);

CREATE SEQUENCE public.billing_schedule_id_seq
    AS integer;
    
ALTER SEQUENCE public.billing_schedule_id_seq OWNED BY public.billing_schedule.id;

ALTER TABLE ONLY public.billing_schedule ALTER COLUMN id SET DEFAULT nextval('public.billing_schedule_id_seq'::regclass);

CREATE POLICY rls_billing_schedule ON "billing_schedule" USING (permission_check(resource_path, 'billing_schedule')) WITH CHECK (permission_check(resource_path, 'billing_schedule'));

ALTER TABLE "billing_schedule" ENABLE ROW LEVEL security;
ALTER TABLE "billing_schedule" FORCE ROW LEVEL security;
