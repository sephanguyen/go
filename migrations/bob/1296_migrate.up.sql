CREATE TABLE IF NOT EXISTS public.api_keypair (
  public_key TEXT NOT NULL,
  user_id TEXT NOT NULL,
  private_key TEXT NOT NULL,

  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  deleted_at timestamp with time zone,
  resource_path TEXT DEFAULT autofillresourcepath(),

  CONSTRAINT api_keypair__pk PRIMARY KEY (public_key),
  CONSTRAINT api_keypair_user_id__fk FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);

CREATE POLICY rls_api_keypair ON "api_keypair"
USING (permission_check(resource_path, 'api_keypair'))
WITH CHECK (permission_check(resource_path, 'api_keypair'));

CREATE POLICY rls_api_keypair_restrictive ON "api_keypair" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'api_keypair'))
with check (permission_check(resource_path, 'api_keypair'));


ALTER TABLE IF EXISTS public.api_keypair ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.api_keypair FORCE ROW LEVEL security;
