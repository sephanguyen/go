DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'discount') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.discount;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'order_item') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.order_item;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'material') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.material;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'billing_ratio') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.billing_ratio;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'billing_schedule_period') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.billing_schedule_period;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'billing_schedule') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.billing_schedule;
  END IF;
END;
$$
