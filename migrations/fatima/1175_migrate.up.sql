DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='alloydb_publication') THEN
      CREATE PUBLICATION alloydb_publication;
   END IF;
END
$do$;

ALTER PUBLICATION alloydb_publication ADD TABLE
public.leaving_reason,
public.order_item_course,
public.package_quantity_type_mapping,
public.bill_item_course,
public.student_product,
public.order_action_log,
public.product,
public.accounting_category,
public.product_setting,
public.package_course,
public.product_location;
