CREATE TABLE IF NOT EXISTS fatima.order_item (
    order_item_id TEXT NOT NULL,
    order_id TEXT NOT NULL,
    product_id TEXT,
    discount_id TEXT,
    start_date timestamp with time zone,
    student_product_id TEXT,
    product_name TEXT,
    effective_date timestamp with time zone,
    cancellation_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT order_item_pk PRIMARY KEY (order_item_id)
);

CREATE TABLE IF NOT EXISTS fatima.discount (
    discount_id TEXT NOT NULL,
    name TEXT NOT NULL,
    discount_type TEXT NOT NULL,
    discount_amount_type TEXT NOT NULL,
    discount_amount_value numeric(12,2) NOT NULL,
    recurring_valid_duration INTEGER,
    available_from timestamp with time zone NOT NULL,
    available_until timestamp with time zone NOT NULL,
    remarks TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    student_tag_id_validation TEXT,
    parent_tag_id_validation TEXT,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT discount_pk PRIMARY KEY (discount_id)
);

CREATE TABLE IF NOT EXISTS fatima.material (
    material_id TEXT NOT NULL,
    material_type TEXT NOT NULL,
    custom_billing_date timestamp with time zone,
    resource_path TEXT NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    CONSTRAINT material_pk PRIMARY KEY (material_id)
);

CREATE TABLE IF NOT EXISTS fatima.product (
    product_id TEXT NOT NULL,
    name TEXT NOT NULL,
    product_type TEXT NOT NULL,
    tax_id TEXT,
    available_from timestamp with time zone NOT NULL,
    available_until timestamp with time zone NOT NULL,
    remarks TEXT,
    custom_billing_period timestamp with time zone,
    billing_schedule_id TEXT,
    disable_pro_rating_flag BOOLEAN NOT NULL DEFAULT FALSE,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    is_unique BOOLEAN DEFAULT FALSE,
    resource_path TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    CONSTRAINT product_pk PRIMARY KEY (product_id)
);

CREATE TABLE IF NOT EXISTS fatima.billing_ratio (
    billing_ratio_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    billing_schedule_period_id TEXT NOT NULL,
    billing_ratio_numerator INTEGER NOT NULL,
    billing_ratio_denominator INTEGER NOT NULL,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT billing_ratio_pk PRIMARY KEY (billing_ratio_id)
);

CREATE TABLE IF NOT EXISTS fatima.billing_schedule_period (
    billing_schedule_period_id TEXT NOT NULL,
    name TEXT NOT NULL,
    billing_schedule_id TEXT NOT NULL,
    start_date timestamp with time zone NOT NULL,
    end_date timestamp with time zone NOT NULL,
    billing_date timestamp with time zone NOT NULL,
    remarks TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT billing_schedule_period_pk PRIMARY KEY (billing_schedule_period_id)
);

CREATE TABLE IF NOT EXISTS fatima.billing_schedule (
    billing_schedule_id TEXT NOT NULL,
    name TEXT NOT NULL,
    remarks TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT,
    deleted_at timestamp with time zone,
    CONSTRAINT billing_schedule_pk PRIMARY KEY (billing_schedule_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE 
fatima.discount,
fatima.order_item,
fatima.product,
fatima.material,
fatima.billing_ratio,
fatima.billing_schedule_period,
fatima.billing_schedule;
