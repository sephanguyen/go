ALTER TABLE public.order ADD COLUMN order_type text;

ALTER TABLE IF EXISTS public.order_product
    RENAME TO order_item;

ALTER TABLE order_item RENAME CONSTRAINT order_product_pk TO order_item_pk;

ALTER TABLE order_item RENAME CONSTRAINT fk_order_product_order_id TO fk_order_item_order_id;
ALTER TABLE order_item RENAME CONSTRAINT fk_order_product_product_id TO fk_order_item_product_id;
ALTER TABLE order_item RENAME CONSTRAINT fk_order_product_discount_id TO fk_order_item_discount_id;

DROP POLICY rls_order_product ON order_item;

CREATE POLICY rls_order_item ON "order_item" using (permission_check(resource_path, 'order_item')) with check (permission_check(resource_path, 'order_item'));

ALTER TABLE "order_item" ENABLE ROW LEVEL security;
ALTER TABLE "order_item" FORCE ROW LEVEL security;
