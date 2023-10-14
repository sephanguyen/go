CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_payment_list_public_info (
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
    payment_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    payment_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    CONSTRAINT pk__invoice_payment_list_public_info PRIMARY KEY (payment_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_bill_item_list_public_info (
    invoice_bill_item_id text NOT NULL,
    invoice_id text NOT NULL,
    invoice_sequence_number INTEGER,
    student_id text NOT NULL,
    bill_item_sequence_number integer NOT NULL,
    invoice_bill_item_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_bill_item_updated_at timestamp with time zone DEFAULT (now() at time zone 'utc'),
    invoice_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    CONSTRAINT pk__invoice_bill_item_list_public_info PRIMARY KEY (invoice_bill_item_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.student_payment_detail_history_info (
    student_payment_detail_action_id text NOT NULL,
    student_payment_detail_id text NOT NULL,
    student_id text NOT NULL,
    payment_method text NOT NULL,
    staff_id text NOT NULL,
    action_type text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    CONSTRAINT pk__student_payment_detail_history_info PRIMARY KEY (student_payment_detail_action_id)
);
