CREATE TABLE IF NOT EXISTS public.debezium_heartbeat (
  id INTEGER PRIMARY KEY,
  updated_at TIMESTAMPTZ
);

DO
$$
BEGIN
  IF NOT is_table_in_publication('debezium_publication', 'debezium_heartbeat') THEN
    ALTER PUBLICATION debezium_publication ADD TABLE public.debezium_heartbeat;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'debezium_heartbeat') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.debezium_heartbeat;
  END IF;
END;
$$;
