---- Supporting system account (schedule job,...) ----
ALTER TABLE public.users ADD COLUMN IF NOT EXISTS is_system BOOLEAN DEFAULT false;
