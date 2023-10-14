CREATE TABLE public.billing_ratio_type (
                                         id integer NOT NULL,
                                         name text NOT NULL,
                                         remark text NULL,
                                         is_archived boolean DEFAULT false NOT NULL,
                                         updated_at timestamp with time zone NOT NULL,
                                         created_at timestamp with time zone NOT NULL,
                                         resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.billing_ratio_type
    ADD CONSTRAINT billing_ratio_type_pk PRIMARY KEY (id);

CREATE SEQUENCE public.billing_ratio_type_id_seq
    AS integer;

ALTER SEQUENCE public.billing_ratio_type_id_seq OWNED BY public.billing_ratio_type.id;

ALTER TABLE ONLY public.billing_ratio_type ALTER COLUMN id SET DEFAULT nextval('public.billing_ratio_type_id_seq'::regclass);

CREATE POLICY rls_billing_ratio_type ON "billing_ratio_type" USING (permission_check(resource_path, 'billing_ratio_type')) WITH CHECK (permission_check(resource_path, 'billing_ratio_type'));

ALTER TABLE "billing_ratio_type" ENABLE ROW LEVEL security;
ALTER TABLE "billing_ratio_type" FORCE ROW LEVEL security;
