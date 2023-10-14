CREATE TABLE IF NOT EXISTS public.snapshot_datawarehouse_signal (
    id TEXT PRIMARY KEY, 
    type TEXT, 
    data TEXT
);


DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='publication_for_datawarehouse') THEN
      CREATE PUBLICATION publication_for_datawarehouse;
   END IF;
END
$do$;


ALTER PUBLICATION publication_for_datawarehouse SET TABLE 
public.snapshot_datawarehouse_signal;

CREATE SCHEMA IF NOT EXISTS bob;