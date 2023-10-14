DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'package') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.package;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'fee') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.fee;
  END IF;
END;
$$
