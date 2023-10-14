DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'file') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.file;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'order_item') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.order_item;
  END IF;
END;
$$
