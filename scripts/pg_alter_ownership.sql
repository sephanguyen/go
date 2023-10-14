-- Remember to grant "postgres" and migration SA to yourself first
GRANT "postgres" TO "atlantis@student-coach-e1e95.iam";
GRANT "stag-bob-m@staging-manabie-online.iam" TO "atlantis@student-coach-e1e95.iam";

-- Alter ownerships of tables
DO $$
DECLARE rec record;
BEGIN
    FOR rec IN
        SELECT * FROM pg_tables WHERE schemaname = 'public'
LOOP
    EXECUTE 'ALTER TABLE ' || quote_ident(rec.schemaname) || '.' || quote_ident(rec.tablename) || ' OWNER TO "stag-bob-m@staging-manabie-online.iam"';
END LOOP;
END
$$;

-- Alter ownerships of views
DO $$
DECLARE
    rec record;
BEGIN
    FOR rec IN
        SELECT * FROM information_schema.views WHERE table_schema = 'public'
LOOP
    EXECUTE 'ALTER VIEW ' || quote_ident(rec.schemaname) || '.' || quote_ident(rec.table_name) || ' OWNER TO "stag-bob-m@staging-manabie-online.iam"';
    END LOOP;
END
$$;

-- Alter ownerships of sequences
DO $$
DECLARE
    rec record;
BEGIN
    FOR rec IN
        SELECT * FROM information_schema.sequences WHERE sequence_schema = 'public'
LOOP
    EXECUTE 'ALTER SEQUENCE ' || quote_ident(rec.schemaname) || '.' || quote_ident(rec.sequence_name) || ' OWNER TO "stag-bob-m@staging-manabie-online.iam"';
    END LOOP;
END
$$;

-- Alter ownerships of publications
DO $$
DECLARE
    rec record;
BEGIN
    FOR rec IN
        SELECT * FROM pg_publication
LOOP
    EXECUTE 'ALTER PUBLICATION ' || quote_ident(rec.schemaname) || '.' || quote_ident(rec.pubname) || ' OWNER TO "stag-bob-m@staging-manabie-online.iam"';
    END LOOP;
END
$$;

-- Alter ownerships of functions
DO $$
DECLARE
    rec record;
BEGIN
    FOR rec IN
        SELECT * FROM information_schema.routines WHERE routine_type = 'FUNCTION' AND routine_schema = 'public'
LOOP
    EXECUTE 'ALTER FUNCTION ' || quote_ident(rec.schemaname) || '.' || quote_ident(rec.routine_name) || ' OWNER TO "stag-bob-m@staging-manabie-online.iam"';
    END LOOP;
END
$$;

-- Revoke "stag-bob-m@staging-manabie-online.iam" and migration SA from yourself
REVOKE "postgres" FROM "atlantis@student-coach-e1e95.iam";
REVOKE "stag-bob-m@staging-manabie-online.iam" FROM "atlantis@student-coach-e1e95.iam";

-- Get all functions owned by "stag-bob-m@staging-manabie-online.iam"
SELECT n.nspname as "Schema",
  p.proname as "Name",
  pg_catalog.pg_get_function_arguments(p.oid) as "Argument data types",
 pg_catalog.pg_get_userbyid(p.proowner) as "Owner"
FROM pg_catalog.pg_proc p
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
     LEFT JOIN pg_catalog.pg_language l ON l.oid = p.prolang
WHERE pg_catalog.pg_function_is_visible(p.oid)
      AND n.nspname <> 'pg_catalog'
      AND n.nspname <> 'information_schema'
      AND pg_catalog.pg_get_userbyid(p.proowner) = '"stag-bob-m@staging-manabie-online.iam"'
ORDER BY 1, 2, 3;

-- Get all functions owned by "stag-bob-m@staging-manabie-online.iam"
SELECT p.proname as "Name",
  pg_catalog.pg_get_function_arguments(p.oid) as "Argument data types"
FROM pg_catalog.pg_proc p
     LEFT JOIN pg_catalog.pg_namespace n ON n.oid = p.pronamespace
     LEFT JOIN pg_catalog.pg_language l ON l.oid = p.prolang
WHERE pg_catalog.pg_function_is_visible(p.oid)
      AND n.nspname <> 'pg_catalog'
      AND n.nspname <> 'information_schema'
      AND pg_catalog.pg_get_userbyid(p.proowner) = '"stag-bob-m@staging-manabie-online.iam"'
ORDER BY 1;
