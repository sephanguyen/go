CREATE TABLE public.leaving_reason (
    id integer NOT NULL,
    name text NOT NULL,
    leaving_reason_type text NOT NULL,
    remark text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.leaving_reason
    ADD CONSTRAINT leaving_reason_pk PRIMARY KEY (id);

CREATE SEQUENCE public.leaving_reason_id_seq
    AS integer;

ALTER SEQUENCE public.leaving_reason_id_seq OWNED BY public.leaving_reason.id;

ALTER TABLE ONLY public.leaving_reason ALTER COLUMN id SET DEFAULT nextval('public.leaving_reason_id_seq'::regclass);

CREATE POLICY rls_leaving_reason ON "leaving_reason" USING (permission_check(resource_path, 'leaving_reason')) WITH CHECK (permission_check(resource_path, 'leaving_reason'));

ALTER TABLE "leaving_reason" ENABLE ROW LEVEL security;
ALTER TABLE "leaving_reason" FORCE ROW LEVEL security;
