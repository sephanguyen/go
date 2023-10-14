ALTER TABLE public."course_type" 
    ADD COLUMN IF NOT EXISTS "remarks" text,
    ADD COLUMN IF NOT EXISTS "is_archived" boolean DEFAULT FALSE;