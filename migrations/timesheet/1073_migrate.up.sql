
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='alloydb_publication') THEN
      CREATE PUBLICATION alloydb_publication;
   END IF;
END
$do$;

ALTER PUBLICATION alloydb_publication SET TABLE
public.alloydb_dbz_signal;