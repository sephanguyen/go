ALTER TABLE public.partner_bank 
    ADD COLUMN IF NOT EXISTS record_limit INTEGER DEFAULT 0;