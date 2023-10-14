DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'invoice_adjustment') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.invoice_adjustment;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'invoice_schedule_history') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.invoice_schedule_history;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'invoice_schedule') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.invoice_schedule;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'invoice_schedule_student') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.invoice_schedule_student;
  END IF;
END;
$$
