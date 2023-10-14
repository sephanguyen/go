CREATE TABLE IF NOT EXISTS public.granted_permission (
  user_group_id TEXT NOT NULL,
  user_group_name TEXT NOT NULL,
  role_id TEXT NOT NULL,
  role_name TEXT NOT NULL,
  permission_id TEXT NOT NULL,
  permission_name TEXT NOT NULL,
  location_id TEXT NOT NULL,
  resource_path TEXT NOT NULL,
  CONSTRAINT granted_permission__pk PRIMARY KEY (user_group_id, role_id, permission_id, location_id)
);

CREATE POLICY rls_granted_permission ON "granted_permission"
USING (permission_check(resource_path, 'granted_permission'))
WITH CHECK (permission_check(resource_path, 'granted_permission'));

CREATE POLICY rls_granted_permission_restrictive ON "granted_permission" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'granted_permission'))
WITH CHECK (permission_check(resource_path, 'granted_permission'));

ALTER TABLE IF EXISTS public.granted_permission ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS public.granted_permission FORCE ROW LEVEL security;

CREATE INDEX IF NOT EXISTS granted_permission__user_group_id__idx ON public.granted_permission USING btree (user_group_id);
CREATE INDEX IF NOT EXISTS granted_permission__role_name__idx ON public.granted_permission USING btree (role_name);
CREATE INDEX IF NOT EXISTS granted_permission__permission_name__idx ON public.granted_permission USING btree (permission_name);


DO
$do$
BEGIN
	IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='kafka_connector') THEN
		GRANT DELETE ON granted_permission to kafka_connector;
	END IF;
END
$do$ 