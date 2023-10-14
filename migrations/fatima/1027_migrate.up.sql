CREATE TABLE public.order (
    id text NOT NULL,
    student_id text NOT NULL,
    location_id text NOT NULL,
    order_sequence_number int NOT NULL,
    order_comment text,
    order_status text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.order ADD CONSTRAINT order_pk PRIMARY KEY (id);

CREATE SEQUENCE public.order_sequence_number_seq
    AS integer;

ALTER SEQUENCE public.order_sequence_number_seq OWNED BY public.order.order_sequence_number;

ALTER TABLE ONLY public.order ALTER COLUMN order_sequence_number SET DEFAULT nextval('public.order_sequence_number_seq'::regclass);

CREATE POLICY rls_order ON "order" using (permission_check(resource_path, 'order')) with check (permission_check(resource_path, 'order'));

ALTER TABLE "order" ENABLE ROW LEVEL security;
ALTER TABLE "order" FORCE ROW LEVEL security;

CREATE TABLE public.order_product (
    order_id text NOT NULL,
    product_id integer NOT NULL,
    discount_id integer,
    start_date timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath()
);

ALTER TABLE ONLY public.order_product ADD CONSTRAINT order_product_pk PRIMARY KEY (order_id,product_id);
ALTER TABLE public.order_product ADD CONSTRAINT fk_order_product_order_id FOREIGN KEY(order_id) REFERENCES public.order(id);
ALTER TABLE public.order_product ADD CONSTRAINT fk_order_product_product_id FOREIGN KEY(product_id) REFERENCES product(id);
ALTER TABLE public.order_product ADD CONSTRAINT fk_order_product_discount_id FOREIGN KEY(discount_id) REFERENCES discount(id);

CREATE POLICY rls_order_product ON "order_product" using (permission_check(resource_path, 'order_product')) with check (permission_check(resource_path, 'order_product'));

ALTER TABLE "order_product" ENABLE ROW LEVEL security;
ALTER TABLE "order_product" FORCE ROW LEVEL security;
