do
$do$
begin
   if not exists (select from pg_publication where pubname='debezium_publication') then
      create publication debezium_publication;
   end if;
end
$do$;

ALTER PUBLICATION debezium_publication SET TABLE 
public.dbz_signals,
public.bill_item,
public.order,
public.product,
public.order_item,
public.discount,
public.student_course;