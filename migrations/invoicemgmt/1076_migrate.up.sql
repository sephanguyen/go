ALTER TABLE IF EXISTS public.billing_address
    DROP COLUMN IF EXISTS prefecture_id;
ALTER TABLE IF EXISTS public.billing_address
    ADD COLUMN prefecture_id text;

DROP INDEX IF EXISTS billing_address__prefecture_id__idx;
CREATE INDEX billing_address__prefecture_id__idx ON public.billing_address (prefecture_id);