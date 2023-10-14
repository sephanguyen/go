CREATE INDEX IF NOT EXISTS school_info__address_gin__idx ON public.school_info USING gin (nospace((address)::text) gin_trgm_ops);
CREATE INDEX IF NOT EXISTS school_info__school_name_gin__idx ON public.school_info USING gin (nospace((school_name)::text) gin_trgm_ops);

-- update get_school_info_list function
DROP FUNCTION IF EXISTS get_school_info_list;
CREATE OR REPLACE FUNCTION public.get_school_info_list(
    search_text TEXT DEFAULT NULL,
    level_id TEXT DEFAULT NULL
) RETURNS SETOF public.school_info
    LANGUAGE SQL STABLE
    AS $$
        SELECT s.*
        FROM school_info AS s
        WHERE (
            search_text IS NULL
            OR nospace(s.school_name) ILIKE nospace(search_text)
            OR nospace(s.address) ILIKE nospace(search_text)
        )
        AND (level_id IS NULL OR s.school_level_id = level_id)
    $$;
