CREATE TABLE IF NOT EXISTS public.invoice_adjustment (
	invoice_adjustment_id TEXT NOT NULL,
    invoice_id TEXT NOT NULL,
    description TEXT NOT NULL,
    amount numeric(12,2) NOT NULL,
    student_id TEXT NOT NULL,
    invoice_adjustment_sequence_number INTEGER,
	invoice_adjustment_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
	invoice_adjustment_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
	invoice_adjustment_deleted_at TIMESTAMP WITH TIME ZONE,
    invoice_type TEXT NOT NULL,
    invoice_status TEXT NOT NULL,
    sub_total numeric(12,2) NOT NULL,
    total numeric(12,2) NOT NULL,
    outstanding_balance numeric(12,2),
    amount_paid numeric(12,2),
    amount_refunded numeric(12,2),
    invoice_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT pk__invoice_adjustment PRIMARY KEY (invoice_adjustment_id)
);

CREATE TABLE IF NOT EXISTS public.invoice_schedule_history (
	invoice_schedule_history_id TEXT NOT NULL,
    invoice_schedule_id TEXT NOT NULL,
    number_of_failed_invoices INTEGER NOT NULL,
    total_students INTEGER NOT NULL,
    execution_start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    execution_end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    invoice_schedule_history_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_history_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_history_deleted_at TIMESTAMP WITH TIME ZONE,
    invoice_date TIMESTAMP WITH TIME ZONE NOT NULL,
    scheduled_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status TEXT NOT NULL,
    is_archived boolean DEFAULT false,
    remarks TEXT,
    user_id TEXT NOT NULL,
    invoice_schedule_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT pk__invoice_schedule_history PRIMARY KEY (invoice_schedule_history_id)
);

CREATE TABLE IF NOT EXISTS public.invoice_schedule_student (
	invoice_schedule_student_id TEXT NOT NULL,
    invoice_schedule_history_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    error_details TEXT NOT NULL,
    actual_error_details TEXT,
    invoice_schedule_student_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_student_updated_at TIMESTAMP WITH TIME ZONE,
    invoice_schedule_student_deleted_at TIMESTAMP WITH TIME ZONE,
    invoice_schedule_id TEXT NOT NULL,
    number_of_failed_invoices INTEGER NOT NULL,
    total_students INTEGER NOT NULL,
    execution_start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    execution_end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    invoice_schedule_history_created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_history_updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    invoice_schedule_history_deleted_at TIMESTAMP WITH TIME ZONE,

	CONSTRAINT pk__invoice_schedule_student PRIMARY KEY (invoice_schedule_student_id)
);

CREATE TABLE IF NOT EXISTS public.invoice_schedule (
    invoice_schedule_id TEXT NOT NULL,
    invoice_date TIMESTAMP WITH TIME ZONE NOT NULL,
    scheduled_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status TEXT NOT NULL,
    is_archived boolean DEFAULT false,
    remarks TEXT,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT pk__invoice_schedule PRIMARY KEY (invoice_schedule_id)
);



ALTER PUBLICATION kec_publication ADD TABLE
    public.invoice_adjustment,
    public.invoice_schedule_history,
    public.invoice_schedule_student,
    public.invoice_schedule;
