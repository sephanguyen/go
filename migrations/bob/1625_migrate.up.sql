DO
$$
BEGIN
  IF NOT is_table_in_publication('debezium_publication', 'school_history') THEN
    ALTER PUBLICATION debezium_publication ADD TABLE public.school_history;
  END IF;
END;
$$
