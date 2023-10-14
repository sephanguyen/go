CREATE SCHEMA IF NOT EXISTS fatima;
CREATE TABLE IF NOT EXISTS fatima.bill_item (
    order_id TEXT NOT NULL,
    bill_item_sequence_number INTEGER,
    product_id TEXT,
    product_description TEXT NOT NULL,
    product_pricing INTEGER,
    discount_amount_type TEXT,
    discount_amount_value numeric(12,2),
    tax_id TEXT,
    tax_category TEXT,
    tax_percentage INTEGER,
    bill_type TEXT NOT NULL,
    billing_status TEXT NOT NULL,
    billing_date timestamp with time zone,
    billing_from timestamp with time zone,
    billing_to timestamp with time zone,
    billing_schedule_period_id text,
    discount_amount numeric(12,2),
    tax_amount numeric(12,2),
    final_price numeric(12,2) NOT NULL,
    student_id TEXT NOT NULL,
    student_product_id TEXT,
    billing_approval_status TEXT,
    billing_item_description JSONB NULL,
    location_id TEXT NOT NULL,
    discount_id TEXT,
    previous_bill_item_sequence_number INTEGER,
    previous_bill_item_status TEXT,
    adjustment_price numeric(12,2),
    is_latest_bill_item boolean NULL,
    price numeric(12,2),
    old_price numeric(12,2),
    billing_ratio_numerator INTEGER,
    billing_ratio_denominator INTEGER,
    is_reviewed BOOLEAN DEFAULT FALSE,
    raw_discount_amount numeric(12,2),
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT bill_item_order_bill_item_sequence_number_pk PRIMARY KEY (order_id, bill_item_sequence_number)
);

CREATE TABLE IF NOT EXISTS fatima.order (
    order_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    order_sequence_number INTEGER,
    order_comment TEXT,
    order_status TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    order_type TEXT,
    student_full_name TEXT NOT NULL,
    is_reviewed BOOLEAN DEFAULT FALSE,
    withdrawal_effective_date timestamp with time zone,
    background TEXT,
    future_measures TEXT,
    loa_start_date timestamp with time zone,
    loa_end_date timestamp with time zone,
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT order_pk PRIMARY KEY (order_id)
);

CREATE TABLE IF NOT EXISTS fatima.tax (
    tax_id TEXT NOT NULL,
    name TEXT NOT NULL,
    tax_percentage INTEGER NOT NULL,
    tax_category TEXT NOT NULL,
    default_flag BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT tax_pk PRIMARY KEY (tax_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE fatima.bill_item;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE fatima.order;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE fatima.tax;
