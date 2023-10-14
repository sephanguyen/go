DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'alloydb_dbz_signal') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.alloydb_dbz_signal;
  END IF;
END;
$$;
