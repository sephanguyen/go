CREATE SCHEMA IF NOT EXISTS invoicemgmt;

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_public_info (
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
    CONSTRAINT pk__invoice_public_info PRIMARY KEY (invoice_id)
);
