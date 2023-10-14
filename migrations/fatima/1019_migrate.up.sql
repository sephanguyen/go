CREATE TABLE public.product_price (
    id integer NOT NULL,
    product_id integer NOT NULL,
    billing_schedule_period_id integer NULL,
    quantity_type text NOT NULL,
    quantity integer NULL,
    price numeric(12,2) NOT NULL,
    created_at timestamp with time zone NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT product_price_quantity CHECK (quantity >= 0),
    CONSTRAINT product_price_price CHECK (price >= 0),
    CONSTRAINT product_fk FOREIGN KEY (product_id) REFERENCES public.product(id),
    CONSTRAINT billing_schedule_period_fk FOREIGN KEY (billing_schedule_period_id) REFERENCES public.billing_schedule_period(id)
);

ALTER TABLE ONLY public.product_price
    ADD CONSTRAINT product_price_pk PRIMARY KEY (id);

CREATE SEQUENCE public.product_price_id_seq
    AS integer;

ALTER SEQUENCE public.product_price_id_seq OWNED BY public.product_price.id;

ALTER TABLE ONLY public.product_price ALTER COLUMN id SET DEFAULT nextval('public.product_price_id_seq'::regclass);

ALTER TABLE ONLY public.product_price
    ADD CONSTRAINT check_quantity_type
    CHECK (quantity_type = 'PRODUCT_PRICE_QUANTITY_TYPE_NONE' OR quantity_type = 'PRODUCT_PRICE_QUANTITY_TYPE_COURSE_WEIGHT' OR quantity_type = 'PRODUCT_PRICE_QUANTITY_TYPE_SLOT' OR quantity_type = 'PRODUCT_PRICE_QUANTITY_TYPE_SLOT_PER_WEEK');

ALTER TABLE ONLY public.product_price
    ADD CONSTRAINT check_quantity_and_quantity_type 
    CHECK ((quantity_type = 'PRODUCT_PRICE_QUANTITY_TYPE_NONE' AND quantity IS NULL) OR (quantity_type <> 'PRODUCT_PRICE_QUANTITY_TYPE_NONE' AND quantity IS NOT NULL));

CREATE UNIQUE INDEX product_price_uni_idx ON public.product_price
(product_id, COALESCE(billing_schedule_period_id, -1), COALESCE(quantity, -1));

CREATE POLICY rls_product_price ON "product_price" USING (permission_check(resource_path, 'product_price')) WITH CHECK (permission_check(resource_path, 'product_price'));

ALTER TABLE "product_price" ENABLE ROW LEVEL security;
ALTER TABLE "product_price" FORCE ROW LEVEL security;
