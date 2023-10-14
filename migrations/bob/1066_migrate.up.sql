ALTER TABLE public.users ALTER COLUMN phone_number DROP NOT NULL ;
ALTER TABLE public.student_parents ADD COLUMN IF NOT EXISTS relationship text NOT NULL ;