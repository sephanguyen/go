CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_action_log (
    invoice_action_id TEXT NOT NULL,
    payment_sequence_number INTEGER,
    invoice_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    action TEXT NOT NULL,
    action_detail TEXT NOT NULL,
    action_comment TEXT NOT NULL,
    resource_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__invoice_action_log PRIMARY KEY (invoice_action_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.new_customer_code_history (
    new_customer_code_history_id TEXT NOT NULL,
    new_customer_code TEXT NOT NULL,
    student_id TEXT NOT NULL,
    bank_account_number TEXT NOT NULL,
    resource_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__new_customer_code_history PRIMARY KEY (new_customer_code_history_id)
);


CREATE TABLE IF NOT EXISTS invoicemgmt.billing_address (
    billing_address_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    student_payment_detail_id TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    city TEXT NOT NULL,
    street1 TEXT NOT NULL,
    street2 TEXT,
    prefecture_code TEXT,
    resource_path TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__billing_address PRIMARY KEY (billing_address_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
invoicemgmt.invoice_action_log,
invoicemgmt.new_customer_code_history,
invoicemgmt.billing_address;
