ALTER TABLE public.lessons
    ADD COLUMN IF NOT EXISTS "classdo_owner_id" TEXT,
    ADD COLUMN IF NOT EXISTS "classdo_link" TEXT
;