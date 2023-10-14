ALTER TABLE invoicemgmt.payment ADD COLUMN IF NOT EXISTS is_exported BOOLEAN DEFAULT FALSE;
ALTER TABLE invoicemgmt.student_payment_detail ADD COLUMN IF NOT EXISTS payer_phone_number TEXT;

CREATE TABLE IF NOT EXISTS invoicemgmt.bank
(
    bank_id text NOT NULL,
    bank_code text NOT NULL,
    bank_name text NOT NULL,
    bank_name_phonetic text NOT NULL,
    is_archived boolean NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT bank__pk PRIMARY KEY (bank_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.bank_branch
(
    bank_branch_id text NOT NULL,
    bank_branch_code text NOT NULL,
    bank_branch_name text NOT NULL,
    bank_branch_phonetic_name text NOT NULL,
    bank_id text NOT NULL,
    is_archived boolean NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT bank_branch__pk PRIMARY KEY (bank_branch_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.partner_bank
(
    partner_bank_id text,
    bank_number text,
    bank_name text,
    bank_branch_number text,
    bank_branch_name text,
    deposit_items text,
    account_number text,
    remarks text,
    is_archived boolean DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    consignor_code text,
    consignor_name text,
    is_default boolean DEFAULT false,
    record_limit integer DEFAULT 0,
    CONSTRAINT partner_bank__pk PRIMARY KEY (partner_bank_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.bank_mapping
(
    bank_mapping_id text NOT NULL,
    bank_id text NOT NULL,
    partner_bank_id text NOT NULL,
    remarks text,
    is_archived boolean DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT bank_mapping__pk PRIMARY KEY (bank_mapping_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.bank_account
(
    bank_account_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    is_verified boolean DEFAULT false,
    bank_branch_id text NOT NULL,
    bank_account_number text NOT NULL,
    bank_account_holder text NOT NULL,
    bank_account_type text NOT NULL,
    bank_id text,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now(),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT bank_account__pk PRIMARY KEY (bank_account_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.bank;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.bank_branch;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.partner_bank;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.bank_mapping;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.bank_account;
