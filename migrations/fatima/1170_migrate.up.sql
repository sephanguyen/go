ALTER TABLE public.discount ADD COLUMN IF NOT EXISTS discount_tag_id TEXT;

ALTER TABLE public.discount
    ADD CONSTRAINT fk_discount_discount_tag_id FOREIGN KEY (discount_tag_id) REFERENCES public.discount_tag(discount_tag_id);
