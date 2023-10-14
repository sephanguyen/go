
DO
$do$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_publication WHERE pubname='alloydb_publication') THEN
      CREATE PUBLICATION alloydb_publication;
   END IF;
END
$do$;

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


-- ALTER PUBLICATION alloydb_publication ADD TABLE
-- public.grade,
-- public.academic_year;



DO
$$
BEGIN
  IF NOT is_table_in_publication('alloydb_publication', 'grade') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.grade;
  END IF;
  IF NOT is_table_in_publication('alloydb_publication', 'academic_year') THEN
    ALTER PUBLICATION alloydb_publication ADD TABLE public.academic_year;
  END IF;
END;
$$
