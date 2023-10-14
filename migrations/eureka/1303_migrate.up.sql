ALTER TABLE public.assessment DROP CONSTRAINT IF EXISTS fk__learning_material_id;
ALTER TABLE public.assessment
    ADD COLUMN IF NOT EXISTS ref_table VARCHAR(20) NOT NULL;
