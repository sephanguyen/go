CREATE TABLE IF NOT EXISTS public.student_associated_product (
	student_product_id TEXT NOT NULL,
    associated_product_id TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
	updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
	deleted_at timestamp with time zone,
	CONSTRAINT student_associated_product_pk PRIMARY KEY (student_product_id, associated_product_id)
);

ALTER PUBLICATION kec_publication ADD TABLE public.student_associated_product;
