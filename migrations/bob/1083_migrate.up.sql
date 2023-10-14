ALTER TABLE public.books ADD COLUMN IF NOT EXISTS current_chapter_display_order INTEGER DEFAULT 0 NOT NULL;
