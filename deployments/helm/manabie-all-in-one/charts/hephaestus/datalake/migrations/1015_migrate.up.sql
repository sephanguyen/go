CREATE TABLE IF NOT EXISTS invoicemgmt.invoice (
    invoice_id text NOT NULL,
    invoice_sequence_number INTEGER,
    type text NOT NULL,
    status text NOT NULL,
    student_id text NOT NULL,
    sub_total numeric(12,2) NOT NULL,
    total numeric(12,2) NOT NULL,
    outstanding_balance numeric(12,2),
    amount_paid numeric(12,2),
    amount_refunded numeric(12,2),
    resource_path text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT invoice_pk PRIMARY KEY (invoice_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.invoice;
