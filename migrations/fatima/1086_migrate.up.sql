--- change master data ids type from int to text (uuids) and update relevant tables ---

-- product
ALTER TABLE public.fee DROP CONSTRAINT fk__fee__id;
ALTER TABLE public.material DROP CONSTRAINT fk__material__id;
ALTER TABLE public.package DROP CONSTRAINT fk__package__id;
ALTER TABLE public.bill_item DROP CONSTRAINT fk_bill_item_product_id;
ALTER TABLE public.order_item DROP CONSTRAINT fk_order_item_product_id;
ALTER TABLE public.package_course DROP CONSTRAINT fk_package_course_id;
ALTER TABLE public.product_accounting_category DROP CONSTRAINT fk_product_id;
ALTER TABLE public.product_grade DROP CONSTRAINT fk_product_id;
ALTER TABLE public.product_location DROP CONSTRAINT fk_product_location_id;
ALTER TABLE public.product_setting DROP CONSTRAINT fk_product_setting_product_id;
ALTER TABLE public.student_product DROP CONSTRAINT fk_student_product_product_id;
ALTER TABLE public.product_price DROP CONSTRAINT product_fk;

ALTER TABLE public.product ALTER COLUMN product_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.product_id_seq;
ALTER TABLE public.product ALTER COLUMN product_id TYPE text;

-- package
ALTER TABLE public.package_course_material DROP CONSTRAINT fk_package_id;
ALTER TABLE public.package_course_fee DROP CONSTRAINT fk_package_id;
ALTER TABLE public.student_package_by_order DROP CONSTRAINT fk_student_package_package_id;
ALTER TABLE public.package ALTER COLUMN package_id TYPE text;

-- material
ALTER TABLE public.package_course_material DROP CONSTRAINT fk_material_id;
ALTER TABLE public.material ALTER COLUMN material_id TYPE text;

-- fee
ALTER TABLE public.package_course_fee DROP CONSTRAINT fk_fee_id;
ALTER TABLE public.fee ALTER COLUMN fee_id TYPE text;

-- accounting_category
ALTER TABLE public.product_accounting_category DROP CONSTRAINT fk_accounting_category_id;

ALTER TABLE public.accounting_category ALTER COLUMN accounting_category_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.accounting_category_id_seq;
ALTER TABLE public.accounting_category ALTER COLUMN accounting_category_id TYPE text;


-- tax
ALTER TABLE public.bill_item DROP CONSTRAINT fk_bill_item_tax_id;
ALTER TABLE public.product DROP CONSTRAINT fk_product_tax_id;

ALTER TABLE public.tax ALTER COLUMN tax_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.tax_id_seq;
ALTER TABLE public.tax ALTER COLUMN tax_id TYPE text;

-- discount
ALTER TABLE public.bill_item DROP CONSTRAINT fk_bill_item_discount_id;
ALTER TABLE public.order_item DROP CONSTRAINT fk_order_item_discount_id;

ALTER TABLE public.discount ALTER COLUMN discount_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.discount_id_seq;
ALTER TABLE public.discount ALTER COLUMN discount_id TYPE text;

-- billing_schedule
ALTER TABLE public.billing_schedule_period DROP CONSTRAINT fk_billing_schedule_id;
ALTER TABLE public.product DROP CONSTRAINT fk_product_billing_schedule_id;

ALTER TABLE public.billing_schedule ALTER COLUMN billing_schedule_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.billing_schedule_id_seq;
ALTER TABLE public.billing_schedule ALTER COLUMN billing_schedule_id TYPE text;


-- billing_schedule_period
ALTER TABLE public.product_price DROP CONSTRAINT billing_schedule_period_fk;
ALTER TABLE public.bill_item DROP CONSTRAINT fk_bill_item_billing_schedule_period_id;
ALTER TABLE public.billing_ratio DROP CONSTRAINT fk_billing_schedule_period_id;

ALTER TABLE public.billing_schedule_period ALTER COLUMN billing_schedule_period_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.billing_schedule_period_id_seq;
ALTER TABLE public.billing_schedule_period ALTER COLUMN billing_schedule_period_id TYPE text;


-- billing_ratio

ALTER TABLE public.billing_ratio ALTER COLUMN billing_ratio_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.billing_ratio_id_seq;
ALTER TABLE public.billing_ratio ALTER COLUMN billing_ratio_id TYPE text;


-- leaving_reason

ALTER TABLE public.leaving_reason ALTER COLUMN leaving_reason_id DROP DEFAULT;
DROP SEQUENCE IF EXISTS public.leaving_reason_id_seq;
ALTER TABLE public.leaving_reason ALTER COLUMN leaving_reason_id TYPE text;



--- add CONSTRAINT back ---

-- product
ALTER TABLE public.fee ADD CONSTRAINT fk_fee_fee_id FOREIGN KEY (fee_id) REFERENCES public.product(product_id);
ALTER TABLE public.material ADD CONSTRAINT fk_material_material_id FOREIGN KEY (material_id) REFERENCES public.product(product_id);
ALTER TABLE public.package ADD CONSTRAINT fk_package_package_id FOREIGN KEY (package_id) REFERENCES public.product(product_id);

ALTER TABLE public.bill_item ALTER COLUMN product_id TYPE text;
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.order_item ALTER COLUMN product_id TYPE text;
ALTER TABLE public.order_item ADD CONSTRAINT fk_order_item_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.product_accounting_category ALTER COLUMN product_id TYPE text;
ALTER TABLE public.product_accounting_category ADD CONSTRAINT fk_product_accounting_category_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.product_grade ALTER COLUMN product_id TYPE text;
ALTER TABLE public.product_grade ADD CONSTRAINT fk_product_grade_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.product_location ALTER COLUMN product_id TYPE text;
ALTER TABLE public.product_location ADD CONSTRAINT fk_product_location_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.product_setting ALTER COLUMN product_id TYPE text;
ALTER TABLE public.product_setting ADD CONSTRAINT fk_product_setting_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.student_product ALTER COLUMN product_id TYPE text;
ALTER TABLE public.student_product ADD CONSTRAINT fk_student_product_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);

ALTER TABLE public.product_price ALTER COLUMN product_id TYPE text;
ALTER TABLE public.product_price ADD CONSTRAINT fk_product_price_product_id FOREIGN KEY (product_id) REFERENCES public.product(product_id);


-- package
ALTER TABLE public.package_course_material ALTER COLUMN package_id TYPE text;
ALTER TABLE public.package_course_material ADD CONSTRAINT fk_package_course_material_package_id FOREIGN KEY (package_id) REFERENCES public.package(package_id);

ALTER TABLE public.package_course_fee ALTER COLUMN package_id TYPE text;
ALTER TABLE public.package_course_fee ADD CONSTRAINT fk_package_course_fee_package_id FOREIGN KEY (package_id) REFERENCES public.package(package_id);

ALTER TABLE public.student_package_by_order ALTER COLUMN package_id TYPE text;
ALTER TABLE public.student_package_by_order ADD CONSTRAINT fk_student_package_by_order_package_id FOREIGN KEY (package_id) REFERENCES public.package(package_id);

ALTER TABLE public.package_course ALTER COLUMN package_id TYPE text;
ALTER TABLE public.package_course ADD CONSTRAINT fk_package_course_package_id FOREIGN KEY (package_id) REFERENCES public.package(package_id);


-- material
ALTER TABLE public.package_course_material ALTER COLUMN material_id TYPE text;
ALTER TABLE public.package_course_material ADD CONSTRAINT fk_package_course_material_material_id FOREIGN KEY (material_id) REFERENCES public.material(material_id);


-- fee
ALTER TABLE public.package_course_fee ALTER COLUMN fee_id TYPE text;
ALTER TABLE public.package_course_fee ADD CONSTRAINT fk_package_course_fee_fee_id FOREIGN KEY (fee_id) REFERENCES public.fee(fee_id);


-- accounting_category
ALTER TABLE public.product_accounting_category ALTER COLUMN accounting_category_id TYPE text;
ALTER TABLE public.product_accounting_category ADD CONSTRAINT fk_product_accounting_category_accounting_category_id FOREIGN KEY (accounting_category_id) REFERENCES public.accounting_category(accounting_category_id);


-- tax
ALTER TABLE public.bill_item ALTER COLUMN tax_id TYPE text;
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_tax_id FOREIGN KEY (tax_id) REFERENCES public.tax(tax_id);

ALTER TABLE public.product ALTER COLUMN tax_id TYPE text;
ALTER TABLE public.product ADD CONSTRAINT fk_product_tax_id FOREIGN KEY (tax_id) REFERENCES public.tax(tax_id);


-- discount
ALTER TABLE public.bill_item ALTER COLUMN discount_id TYPE text;
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_discount_id FOREIGN KEY (discount_id) REFERENCES public.discount(discount_id);

ALTER TABLE public.order_item ALTER COLUMN discount_id TYPE text;
ALTER TABLE public.order_item ADD CONSTRAINT fk_order_item_discount_id FOREIGN KEY (discount_id) REFERENCES public.discount(discount_id);


-- billing_schedule
ALTER TABLE public.billing_schedule_period ALTER COLUMN billing_schedule_id TYPE text;
ALTER TABLE public.billing_schedule_period ADD CONSTRAINT fk_billing_schedule_period_billing_schedule_id FOREIGN KEY (billing_schedule_id) REFERENCES public.billing_schedule(billing_schedule_id);

ALTER TABLE public.product ALTER COLUMN billing_schedule_id TYPE text;
ALTER TABLE public.product ADD CONSTRAINT fk_product_billing_schedule_id FOREIGN KEY (billing_schedule_id) REFERENCES public.billing_schedule(billing_schedule_id);


-- billing_schedule_period
DROP INDEX IF EXISTS product_price_uni_idx;
ALTER TABLE public.product_price ALTER COLUMN billing_schedule_period_id TYPE text;
CREATE UNIQUE INDEX product_price_uni_idx ON public.product_price (product_id, COALESCE(billing_schedule_period_id, ''), COALESCE(quantity, -1));
ALTER TABLE public.product_price ADD CONSTRAINT fk_product_price_billing_schedule_period_id FOREIGN KEY (billing_schedule_period_id) REFERENCES public.billing_schedule_period(billing_schedule_period_id);

ALTER TABLE public.bill_item ALTER COLUMN billing_schedule_period_id TYPE text;
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_billing_schedule_period_id FOREIGN KEY (billing_schedule_period_id) REFERENCES public.billing_schedule_period(billing_schedule_period_id);

ALTER TABLE public.billing_ratio ALTER COLUMN billing_schedule_period_id TYPE text;
ALTER TABLE public.billing_ratio ADD CONSTRAINT fk_billing_ratio_billing_schedule_period_id FOREIGN KEY (billing_schedule_period_id) REFERENCES public.billing_schedule_period(billing_schedule_period_id);


-- others
ALTER TABLE public.order_item_course ALTER COLUMN package_id TYPE text;
