CREATE TABLE IF NOT EXISTS public.file (
	file_id text NOT NULL,
	file_name text NOT NULL,
	file_type text NOT NULL,
	download_link text NOT NULL,
	updated_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	created_at  timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	CONSTRAINT file_pk PRIMARY KEY (file_id)
);

CREATE TABLE IF NOT EXISTS public.order_item (
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
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
	CONSTRAINT order_item_pk PRIMARY KEY (order_item_id)
);

ALTER PUBLICATION kec_publication ADD TABLE 
public.file,
public.order_item;
