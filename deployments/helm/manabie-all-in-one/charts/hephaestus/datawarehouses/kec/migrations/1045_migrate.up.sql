DROP TABLE IF EXISTS public.invoice_bill_item_list;

CREATE TABLE IF NOT EXISTS public.invoice_bill_item_list (
    invoice_bill_item_id text NOT NULL,
    invoice_id text NOT NULL,
    invoice_sequence_number INTEGER,
    past_billing_status TEXT NOT NULL,
    student_id text NOT NULL,
    bill_item_sequence_number integer NOT NULL,
    invoice_bill_item_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_bill_item_updated_at timestamp with time zone DEFAULT (now() at time zone 'utc'),
    invoice_bill_item_deleted_at timestamp with time zone,
    invoice_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_deleted_at timestamp with time zone,
    CONSTRAINT pk__invoice_bill_item_list_public_info PRIMARY KEY (invoice_bill_item_id)
);

DROP TABLE IF EXISTS public.invoice_payment_list;

CREATE TABLE IF NOT EXISTS public.payment (
    payment_id text NOT NULL,
    invoice_id text NOT NULL,
    invoice_sequence_number INTEGER,
    student_id text NOT NULL,
    payment_sequence_number INTEGER,
    payment_status text NOT NULL,
    payment_method text NOT NULL,
    payment_due_date timestamp with time zone NOT NULL,
    payment_expiry_date timestamp with time zone NOT NULL,
    payment_date timestamp with time zone,
    amount numeric(12,2),
    result_code text,
    is_exported BOOLEAN DEFAULT FALSE,
    payment_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    payment_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    payment_deleted_at timestamp with time zone,
    invoice_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_deleted_at timestamp with time zone,
    CONSTRAINT pk__payment PRIMARY KEY (payment_id)
);

DROP TABLE IF EXISTS public.student_payment_detail;
CREATE TABLE IF NOT EXISTS public.student_payment_detail (
    student_payment_detail_action_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    payer_name text NOT NULL,
    payer_phone_number text,
    payment_method text NOT NULL,
    staff_id text NOT NULL,
    action text NOT NULL,
    student_payment_detail_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_deleted_at timestamp with time zone,
    student_payment_detail_action_log_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_action_log_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_action_log_deleted_at timestamp with time zone,
    CONSTRAINT pk__student_payment_detail_history_info PRIMARY KEY (student_payment_detail_action_id)
);

CREATE TABLE IF NOT EXISTS public.bank_branch
(
    bank_branch_id text NOT NULL,
    bank_branch_code text NOT NULL,
    bank_branch_name text NOT NULL,
    bank_branch_phonetic_name text NOT NULL,
    bank_id text NOT NULL,
    bank_branch_is_archived boolean NOT NULL,
    bank_branch_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_branch_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_branch_deleted_at timestamp with time zone,
    bank_code text NOT NULL,
    bank_name text NOT NULL,
    bank_name_phonetic text NOT NULL,
    bank_is_archived boolean NOT NULL,
    bank_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_deleted_at timestamp with time zone,

    CONSTRAINT billing_address__pk PRIMARY KEY (bank_branch_id)
);

CREATE TABLE IF NOT EXISTS public.bank_mapping
(
    bank_mapping_id text NOT NULL,
    bank_id text NOT NULL,
    partner_bank_id text NOT NULL,
    bank_mapping_remarks text,
    bank_mapping_is_archived boolean DEFAULT false,
    bank_mapping_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_mapping_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_mapping_deleted_at timestamp with time zone,
    bank_bank_code text NOT NULL,
    bank_bank_name text NOT NULL,
    bank_bank_name_phonetic text NOT NULL,
    bank_is_archived boolean NOT NULL,
    bank_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_deleted_at timestamp with time zone,
    partner_bank_bank_number text,
    partner_bank_bank_name text,
    partner_bank_bank_branch_number text,
    partner_bank_bank_branch_name text,
    deposit_items text,
    account_number text,
    consignor_code text,
    consignor_name text,
    is_default boolean DEFAULT false,
    record_limit integer DEFAULT 0,
    partner_bank_remarks text,
    partner_bank_is_archived boolean DEFAULT false,
    partner_bank_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    partner_bank_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    partner_bank_deleted_at timestamp with time zone,

    CONSTRAINT bank_mapping__pk PRIMARY KEY (bank_mapping_id)
);

CREATE TABLE IF NOT EXISTS public.bank_account
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
    bank_account_created_at timestamp with time zone NOT NULL DEFAULT now(),
    bank_account_updated_at timestamp with time zone NOT NULL DEFAULT now(),
    bank_account_deleted_at timestamp with time zone,
    payer_name text NOT NULL,
    payer_phone_number text,
    payment_method text NOT NULL,
    student_payment_detail_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    student_payment_detail_deleted_at timestamp with time zone,
    bank_code text NOT NULL,
    bank_name text NOT NULL,
    bank_name_phonetic text NOT NULL,
    bank_is_archived boolean NOT NULL,
    bank_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    bank_deleted_at timestamp with time zone,

    CONSTRAINT bank_account_id__pk PRIMARY KEY (bank_account_id)
);

ALTER PUBLICATION kec_publication ADD TABLE public.bank_branch;
ALTER PUBLICATION kec_publication ADD TABLE public.bank_mapping;
ALTER PUBLICATION kec_publication ADD TABLE public.payment;
ALTER PUBLICATION kec_publication ADD TABLE public.bank_account;

