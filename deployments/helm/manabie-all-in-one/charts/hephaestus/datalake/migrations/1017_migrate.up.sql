
CREATE TABLE IF NOT EXISTS invoicemgmt.payment (
    payment_id text NOT NULL,
    payment_sequence_number INTEGER,
    invoice_id text NOT NULL,
    student_id text,
    payment_status text NOT NULL,
    payment_method text NOT NULL,
    payment_due_date timestamp with time zone NOT NULL,
    payment_expiry_date timestamp with time zone NOT NULL,
    payment_date timestamp with time zone,
    amount numeric(12,2),
    result_code text,
    resource_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT payment_pk PRIMARY KEY (payment_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_bill_item (
    invoice_bill_item_id text NOT NULL,
    invoice_id text NOT NULL,
    bill_item_sequence_number integer NOT NULL,
    past_billing_status text NOT NULL,
    resource_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT invoice_bill_item_pk PRIMARY KEY (invoice_bill_item_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.student_payment_detail (
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    payer_name text NOT NULL,
    payment_method text NOT NULL,
    resource_path text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT student_payment_detail__pk PRIMARY KEY (student_payment_detail_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.student_payment_detail_action_log (
    student_payment_detail_action_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    user_id text NOT NULL,
    action text NOT NULL,
    action_detail JSONB NOT NULL DEFAULT '{}'::jsonb, 
    resource_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT student_payment_detail_action_log__pk PRIMARY KEY (student_payment_detail_action_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
invoicemgmt.payment,
invoicemgmt.invoice_bill_item,
invoicemgmt.student_payment_detail,
invoicemgmt.student_payment_detail_action_log;
