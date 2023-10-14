ALTER TABLE IF EXISTS public.lo_progression 
    ALTER COLUMN deleted_at DROP NOT NULL;