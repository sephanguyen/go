ALTER TABLE ONLY public.invoice
    ADD COLUMN IF NOT EXISTS is_expired boolean DEFAULT false;

ALTER TABLE ONLY public.payment
    ADD COLUMN IF NOT EXISTS is_expired boolean DEFAULT false;