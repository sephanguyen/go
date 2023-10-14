ALTER TABLE public.user_group ADD COLUMN IF NOT EXISTS org_location_id TEXT;

ALTER TABLE public.user_group ADD CONSTRAINT fk__user_group__org_location_id FOREIGN KEY (org_location_id) REFERENCES public.locations(location_id);
