ALTER TABLE public.role ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
ALTER TABLE public.user_group ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
