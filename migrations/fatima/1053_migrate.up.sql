ALTER TABLE public.discount RENAME COLUMN id TO discount_id;
ALTER TABLE public.billing_schedule RENAME COLUMN id TO billing_schedule_id;
ALTER TABLE public.billing_schedule_period RENAME COLUMN id TO billing_schedule_period_id;
ALTER TABLE public.tax RENAME COLUMN id TO tax_id;
ALTER TABLE public.billing_ratio RENAME COLUMN id TO billing_ratio_id;
ALTER TABLE public.leaving_reason RENAME COLUMN id TO leaving_reason_id;
ALTER TABLE public.accounting_category RENAME COLUMN id TO accounting_category_id;
ALTER TABLE public.product_price RENAME COLUMN id TO product_price_id;

ALTER TABLE public.order RENAME COLUMN id TO order_id;
ALTER TABLE public.order_action_log RENAME COLUMN id TO order_action_log_id;

ALTER TABLE public.product RENAME COLUMN id TO product_id;

ALTER TABLE public.product_fee DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_product_fee ON product_fee ;
ALTER TABLE public.product_fee RENAME TO fee;

ALTER TABLE public.product_material DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_product_material ON product_material ;
ALTER TABLE public.product_material RENAME TO material;

ALTER TABLE public.product_package DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_product_package ON product_package ;
ALTER TABLE public.product_package RENAME TO package;

ALTER TABLE public.fee RENAME COLUMN id TO fee_id;
ALTER TABLE public.material RENAME COLUMN id TO material_id;
ALTER TABLE public.package RENAME COLUMN id TO package_id;

ALTER TABLE public.fee RENAME CONSTRAINT product_fee_pk TO fee_pk;
ALTER TABLE public.material RENAME CONSTRAINT product_material_pk TO material_pk;
ALTER TABLE public.package RENAME CONSTRAINT product_package_pk TO package_pk;

ALTER TABLE public.fee RENAME CONSTRAINT fk__product_fee__id TO fk__fee__id;
ALTER TABLE public.material RENAME CONSTRAINT fk__product_material__id TO fk__material__id;
ALTER TABLE public.package RENAME CONSTRAINT fk_product_package_id TO fk__package__id;

CREATE POLICY rls_fee ON "fee" USING (permission_check(resource_path, 'fee')) WITH CHECK (permission_check(resource_path, 'fee'));
ALTER TABLE "fee" ENABLE ROW LEVEL security;
ALTER TABLE "fee" FORCE ROW LEVEL security;

CREATE POLICY rls_material ON "material" USING (permission_check(resource_path, 'material')) WITH CHECK (permission_check(resource_path, 'material'));
ALTER TABLE "material" ENABLE ROW LEVEL security;
ALTER TABLE "material" FORCE ROW LEVEL security;

CREATE POLICY rls_package ON "package" USING (permission_check(resource_path, 'package')) WITH CHECK (permission_check(resource_path, 'package'));
ALTER TABLE "package" ENABLE ROW LEVEL security;
ALTER TABLE "package" FORCE ROW LEVEL security;
