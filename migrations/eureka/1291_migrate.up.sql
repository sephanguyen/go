ALTER TABLE public.books
	ADD COLUMN IF NOT EXISTS is_v2 boolean DEFAULT false;
