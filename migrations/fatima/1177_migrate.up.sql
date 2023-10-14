-- ALTER PUBLICATION alloydb_publication ADD TABLE
-- public.bill_item,
-- public.order,
-- public.tax;


DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'bill_item') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.bill_item;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'order') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.order;
  END IF;
    IF NOT is_table_in_publication('alloydb_publication', 'tax') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.tax;
  END IF;
END;
$$
