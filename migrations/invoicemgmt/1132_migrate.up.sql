DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'partner_convenience_store') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.partner_convenience_store;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'company_detail') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.company_detail;
  END IF;
END;
$$
