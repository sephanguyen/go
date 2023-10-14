DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'bank') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.bank;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'bank_branch') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.bank_branch;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'partner_bank') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.partner_bank;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'bank_mapping') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.bank_mapping;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'bank_account') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.bank_account;
  END IF;
END;
$$
