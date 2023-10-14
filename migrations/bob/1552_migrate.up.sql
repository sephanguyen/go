DROP INDEX IF EXISTS users__name_gin__idx;
DROP INDEX IF EXISTS users__full_name_phonetic_gin__idx;
DROP FUNCTION IF EXISTS nospace;

CREATE OR REPLACE FUNCTION nospace(value TEXT) 
    RETURNS TEXT AS
    $BODY$
        DECLARE
        BEGIN
            value = REGEXP_REPLACE(value, '[[:space:]]|　', '', 'g');
            
            RETURN value;
        END
    $BODY$
LANGUAGE 'plpgsql' IMMUTABLE;

CREATE INDEX IF NOT EXISTS users__name_gin__idx ON public.users USING gin (nospace((name)::text) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS users__full_name_phonetic_gin__idx ON public.users USING gin (nospace((full_name_phonetic)::text) gin_trgm_ops);
