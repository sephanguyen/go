DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'invoice_action_log') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.invoice_action_log;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'new_customer_code_history') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.new_customer_code_history;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'billing_address') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.billing_address;
  END IF;
END;
$$
