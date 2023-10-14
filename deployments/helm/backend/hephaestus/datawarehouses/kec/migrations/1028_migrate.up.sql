CREATE TABLE IF NOT EXISTS public.invoice (
    invoice_id TEXT NOT NULL,
    invoice_sequence_number INTEGER,
    type text NOT NULL,
    status text NOT NULL,
    student_id text NOT NULL,
    sub_total numeric(12,2) NOT NULL,
    total numeric(12,2) NOT NULL,
    outstanding_balance numeric(12,2),
    amount_paid numeric(12,2),
    amount_refunded numeric(12,2),
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT pk__invoice PRIMARY KEY (invoice_id)
);

CREATE TABLE IF NOT EXISTS public.invoice_payment_list (
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

CREATE TABLE IF NOT EXISTS public.invoice_bill_item_list (
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

CREATE TABLE IF NOT EXISTS public.student_payment_detail (
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


ALTER PUBLICATION kec_publication ADD TABLE 
    public.invoice,
    public.invoice_payment_list,
    public.invoice_bill_item_list,
    public.student_payment_detail;
