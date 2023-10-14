
CREATE TABLE IF NOT EXISTS public.debezium_heartbeat (
  id INTEGER PRIMARY KEY,
  updated_at TIMESTAMPTZ
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE public.debezium_heartbeat;
