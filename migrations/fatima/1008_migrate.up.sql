CREATE TABLE public.accounting_category (
    id integer NOT NULL,
    name text NOT NULL,
    remarks text NULL,
    is_archived boolean DEFAULT false NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.accounting_category
    ADD CONSTRAINT accounting_category_pk PRIMARY KEY (id);

CREATE SEQUENCE public.accounting_category_id_seq
    AS integer;
    
ALTER SEQUENCE public.accounting_category_id_seq OWNED BY public.accounting_category.id;

ALTER TABLE ONLY public.accounting_category ALTER COLUMN id SET DEFAULT nextval('public.accounting_category_id_seq'::regclass);

CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;

do $$
declare selectrow record;
begin
for selectrow in
    select 'CREATE POLICY rls_' || T.mytable || ' ON ' || T.mytable || ' using (permission_check(resource_path,' || T.mytable || '::text)) with check (permission_check(resource_path,'|| T.mytable || '::text));' as script
    from (select tablename as mytable from pg_tables where schemaname = 'public') t
loop
    begin
        execute selectrow.script;
    end;
end loop;
end;
$$;

ALTER TABLE "accounting_category" ENABLE ROW LEVEL security;
ALTER TABLE "accounting_category" FORCE ROW LEVEL security;
