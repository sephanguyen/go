DROP INDEX IF EXISTS school_info__school_level_id__idx;

CREATE INDEX IF NOT EXISTS school_info__school_level_id__idx ON public.school_info(school_level_id);
