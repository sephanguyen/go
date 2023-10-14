CREATE TABLE IF NOT EXISTS public.bank (
	bank_id text NOT NULL,
	bank_code text NOT NULL,
	bank_name text NOT NULL,
	bank_name_phonetic text NOT NULL,
	is_archived bool NOT NULL,
	created_at timestamptz NOT NULL,
	updated_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT bank__bank_code__unique UNIQUE (bank_code, resource_path),
	CONSTRAINT bank__pk PRIMARY KEY (bank_id)
);

CREATE POLICY rls_bank ON "bank"
USING (permission_check(resource_path, 'bank'))
WITH CHECK (permission_check(resource_path, 'bank'));

ALTER TABLE "bank" ENABLE ROW LEVEL security;
ALTER TABLE "bank" FORCE ROW LEVEL security;