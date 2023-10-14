CREATE TABLE IF NOT EXISTS dbz_signals (
    id TEXT PRIMARY KEY, 
    type TEXT, 
    data TEXT
);

CREATE TABLE IF NOT EXISTS public.debezium_heartbeat (
  id INTEGER PRIMARY KEY,
  updated_at TIMESTAMPTZ
);

DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='debezium_publication') THEN
      CREATE PUBLICATION debezium_publication;
   END IF;
END
$do$;

ALTER PUBLICATION debezium_publication SET TABLE 
public.dbz_signals,
public.debezium_heartbeat,
public.bank,
public.bank_account;
