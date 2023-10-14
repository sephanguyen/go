DROP TABLE public.bill_item_custom;

ALTER TABLE public.bill_item ALTER COLUMN discount_amount_value DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN discount_amount DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN student_product_id DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN product_pricing DROP NOT NULL;
ALTER TABLE public.bill_item ALTER COLUMN product_id DROP NOT NULL;

DROP TABLE public.order_item_custom;

ALTER TABLE public.order_item_course DROP CONSTRAINT IF EXISTS order_item_course_pk;
ALTER TABLE public.order_item_course DROP CONSTRAINT IF EXISTS order_item_course_order_item_fk;
ALTER TABLE public.order_item_course ALTER COLUMN package_id DROP NOT NULL;
ALTER TABLE public.order_item_course ADD COLUMN order_item_course_id text NOT NULL;
ALTER TABLE public.order_item_course ADD CONSTRAINT order_item_course_id_pk PRIMARY KEY (order_item_course_id);

ALTER TABLE public.order_item DROP CONSTRAINT IF EXISTS order_item_pk;
ALTER TABLE public.order_item ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE public.order_item ALTER COLUMN student_product_id DROP NOT NULL;
ALTER TABLE public.order_item ADD COLUMN product_name text;
ALTER TABLE public.order_item ADD CONSTRAINT order_item_id_pk PRIMARY KEY (order_item_id);

ALTER TABLE public.order_item_course ADD CONSTRAINT order_item_course_order_id_fk FOREIGN KEY (order_id) REFERENCES public.order(order_id);
ALTER TABLE public.bill_item ADD CONSTRAINT bill_item_location_id_fk FOREIGN KEY (location_id) REFERENCES public.locations(location_id);
