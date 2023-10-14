CREATE TABLE IF NOT EXISTS public.package(
    package_id TEXT NOT NULL,
    package_type TEXT NOT NULL,
    max_slot INTEGER NOT NULL,
    package_start_date timestamp with time zone,
    package_end_date timestamp with time zone,
    package_created_at timestamp with time zone NULL,
    package_updated_at timestamp with time zone NULL,
    package_deleted_at timestamp with time zone NULL,
    name TEXT NOT NULL,
    product_type TEXT NOT NULL,
    tax_id TEXT,
    available_from timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    available_until timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    remarks TEXT,
    custom_billing_period TIMESTAMP WITH TIME ZONE,
    billing_schedule_id TEXT,
    disable_pro_rating_flag BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    product_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    product_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    product_deleted_at timestamp with time zone,
    is_unique BOOLEAN DEFAULT false,
    CONSTRAINT pk_package_id PRIMARY KEY (package_id)
);

CREATE TABLE IF NOT EXISTS public.fee(
    fee_id TEXT NOT NULL,
    fee_type TEXT NOT NULL,
    fee_created_at timestamp with time zone NULL,
    fee_updated_at timestamp with time zone NULL,
    fee_deleted_at timestamp with time zone NULL,
    name TEXT NOT NULL,
    product_type TEXT NOT NULL,
    tax_id TEXT,
    available_from timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    available_until timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    remarks TEXT,
    custom_billing_period TIMESTAMP WITH TIME ZONE,
    billing_schedule_id TEXT,
    disable_pro_rating_flag BOOLEAN NOT NULL DEFAULT false,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    product_created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    product_updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    product_deleted_at timestamp with time zone,
    is_unique BOOLEAN DEFAULT false,
    CONSTRAINT pk_fee_id PRIMARY KEY (fee_id)
);

ALTER PUBLICATION kec_publication ADD TABLE 
public.package,
public.fee;
