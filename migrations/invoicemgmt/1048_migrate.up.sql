CREATE TABLE IF NOT EXISTS bank_mapping (
    bank_mapping_id text NOT NULL,
    bank_id text NOT NULL,
    partner_bank_id text NOT NULL,
    remarks TEXT,
    is_archived BOOLEAN DEFAULT false,

    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NULL,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT bank_mapping__pk PRIMARY KEY (bank_mapping_id),
    CONSTRAINT bank_mapping_bank_fk FOREIGN KEY (bank_id) REFERENCES "bank"(bank_id),
    CONSTRAINT bank_mapping_partner_bank_fk FOREIGN KEY (partner_bank_id) REFERENCES "partner_bank"(partner_bank_id)
);

CREATE POLICY rls_bank_mapping ON "bank_mapping"
USING (permission_check(resource_path, 'bank_mapping'))
WITH CHECK (permission_check(resource_path, 'bank_mapping'));

CREATE POLICY rls_bank_mapping_restrictive ON "bank_mapping" 
AS RESTRICTIVE TO public 
USING (permission_check(resource_path, 'bank_mapping'))
WITH CHECK (permission_check(resource_path, 'bank_mapping'));

ALTER TABLE "bank_mapping" ENABLE ROW LEVEL security;
ALTER TABLE "bank_mapping" FORCE ROW LEVEL security;