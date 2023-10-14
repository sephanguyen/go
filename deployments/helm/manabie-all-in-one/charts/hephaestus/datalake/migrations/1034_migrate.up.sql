CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_adjustment (
    invoice_adjustment_id TEXT NOT NULL,
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    amount numeric(12,2) NOT NULL,
    student_id TEXT NOT NULL,
    invoice_adjustment_sequence_number INTEGER,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT NOT NULL,

    CONSTRAINT pk__invoice_adjustment PRIMARY KEY (invoice_adjustment_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_schedule_history (
    invoice_schedule_history_id TEXT NOT NULL,
    invoice_schedule_id TEXT NOT NULL,
    number_of_failed_invoices INTEGER NOT NULL,
    total_students INTEGER NOT NULL,
    execution_start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    execution_end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    resource_path TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT pk__invoice_schedule_history PRIMARY KEY (invoice_schedule_history_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_schedule (
    invoice_schedule_id TEXT NOT NULL,
    invoice_date TIMESTAMP WITH TIME ZONE NOT NULL,
    scheduled_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status TEXT NOT NULL,
    is_archived boolean DEFAULT false,
    remarks TEXT,
    user_id TEXT NOT NULL,
    resource_path TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__invoice_schedule PRIMARY KEY (invoice_schedule_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.invoice_schedule_student (
    invoice_schedule_student_id TEXT NOT NULL,
    invoice_schedule_history_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    error_details TEXT NOT NULL,
    resource_path TEXT NOT NULL,
    actual_error_details TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__invoice_schedule_student PRIMARY KEY (invoice_schedule_student_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
invoicemgmt.invoice_adjustment,
invoicemgmt.invoice_schedule_history,
invoicemgmt.invoice_schedule,
invoicemgmt.invoice_schedule_student;
