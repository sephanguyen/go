ALTER TABLE public.product_setting ADD COLUMN IF NOT EXISTS is_added_to_enrollment_by_default boolean DEFAULT false;
