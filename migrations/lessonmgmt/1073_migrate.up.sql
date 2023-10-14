AlTER TABLE courses
    ADD COLUMN IF NOT EXISTS "course_type_id" text;

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

CREATE INDEX IF NOT EXISTS user_baic_info_name_gin_idx ON public.user_basic_info USING gin (nospace((name)::text) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS user_baic_info_full_name_phonetic_gin_idx ON public.user_basic_info USING gin (nospace((full_name_phonetic)::text) gin_trgm_ops);
