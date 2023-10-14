--Create bank_branch table--
CREATE TABLE IF NOT EXISTS public.bank_branch
(
    bank_branch_id            TEXT                                NOT NULL
        CONSTRAINT bank_branch__pk PRIMARY KEY,
    bank_branch_code          TEXT                                NOT NULL,
    bank_branch_name          TEXT                                NOT NULL,
    bank_branch_phonetic_name TEXT                                NOT NULL,
    bank_id                   TEXT                                NOT NULL
        CONSTRAINT bank_branch__bank_id__fk
            REFERENCES bank,
    is_archived               BOOLEAN                             NOT NULL,
    created_at                TIMESTAMP WITH TIME ZONE            NOT NULL,
    updated_at                TIMESTAMP WITH TIME ZONE            NOT NULL,
    deleted_at                TIMESTAMP WITH TIME ZONE,
    resource_path             TEXT DEFAULT autofillresourcepath() NOT NULL,
    CONSTRAINT bank_branch__bank_branch_code__unique
        UNIQUE (bank_branch_code, bank_id, resource_path)
);

--Create policy for bank_branch table--
CREATE POLICY rls_bank_branch ON "bank_branch"
    USING (permission_check(resource_path, 'bank_branch'))
    WITH CHECK (permission_check(resource_path, 'bank_branch'));
CREATE POLICY rls_bank_branch_restrictive ON "bank_branch" AS RESTRICTIVE TO PUBLIC
    USING (permission_check(resource_path, 'bank_branch'))
    WITH CHECK (permission_check(resource_path, 'bank_branch'));

--Enable rls for bank_branch table--
ALTER TABLE "bank_branch"
    ENABLE ROW LEVEL security;
ALTER TABLE "bank_branch"
    FORCE ROW LEVEL security;