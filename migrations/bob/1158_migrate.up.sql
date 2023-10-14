CREATE INDEX IF NOT EXISTS lessons__start_time__idx_asc_nulls_last ON public.lessons (start_time ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS lessons__start_time__idx_desc_nulls_first ON public.lessons (start_time DESC NULLS FIRST);
CREATE INDEX IF NOT EXISTS lessons__end_time__idx_asc_nulls_last ON public.lessons (end_time ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS lessons__end_time__idx_desc_nulls_first ON public.lessons (end_time DESC NULLS FIRST);
