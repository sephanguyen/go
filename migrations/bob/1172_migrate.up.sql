CREATE TABLE IF NOT EXISTS public.partner_sync_data_log (
  partner_sync_data_log_id TEXT NOT NULL PRIMARY KEY,
  signature TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  resource_path TEXT DEFAULT autofillresourcepath()
);

CREATE TABLE IF NOT EXISTS public.partner_sync_data_log_split (
  partner_sync_data_log_split_id TEXT NOT NULL PRIMARY KEY,
  partner_sync_data_log_id TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  payload JSONB NOT NULL,
  retry_times INT,
  created_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL,
  resource_path TEXT DEFAULT autofillresourcepath(),
  CONSTRAINT partner_sync_data_log_split__fk FOREIGN KEY (partner_sync_data_log_id) REFERENCES partner_sync_data_log(partner_sync_data_log_id)
);

CREATE POLICY rls_partner_sync_data_log ON "partner_sync_data_log" using (permission_check(resource_path, 'partner_sync_data_log')) with check (permission_check(resource_path, 'partner_sync_data_log'));
ALTER TABLE "partner_sync_data_log" ENABLE ROW LEVEL security;
ALTER TABLE "partner_sync_data_log" FORCE ROW LEVEL security; 

CREATE POLICY rls_partner_sync_data_log_split ON "partner_sync_data_log_split" using (permission_check(resource_path, 'partner_sync_data_log_split')) with check (permission_check(resource_path, 'partner_sync_data_log_split'));
ALTER TABLE "partner_sync_data_log_split" ENABLE ROW LEVEL security;
ALTER TABLE "partner_sync_data_log_split" FORCE ROW LEVEL security; 
