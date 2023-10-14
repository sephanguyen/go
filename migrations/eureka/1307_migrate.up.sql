CREATE OR REPLACE FUNCTION is_table_in_publication(publication_name text, table_name text)
RETURNS BOOLEAN AS
$$
BEGIN
  RETURN EXISTS (
    SELECT 1
    FROM pg_publication_rel pr
    JOIN pg_class c ON pr.prrelid = c.oid
    JOIN pg_namespace n ON c.relnamespace = n.oid
    JOIN pg_publication p ON pr.prpubid = p.oid
    WHERE p.pubname = publication_name AND c.relname = table_name
  );
END;
$$
LANGUAGE plpgsql;


-- DO
-- $$
-- BEGIN
--   IF NOT is_table_in_publication('debezium_publication', 'test_table') THEN
--     ALTER PUBLICATION debezium_publication ADD TABLE public.test_table;
--   END IF;
-- END;
-- $$
