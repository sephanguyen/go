ALTER TABLE IF EXISTS public.bank
    ALTER COLUMN bank_code TYPE TEXT USING bank_code::text,
    ADD CONSTRAINT bank__bank_code__unique UNIQUE (bank_code, resource_path);

ALTER TABLE IF EXISTS public.school_level_grade
    DROP COLUMN IF EXISTS is_archived;

ALTER TABLE IF EXISTS public.school_info
    ALTER COLUMN school_partner_id SET NOT NULL;

ALTER TABLE IF EXISTS public.school_course
    ALTER COLUMN school_course_partner_id SET NOT NULL;

ALTER TABLE IF EXISTS public.user_tag
    ALTER COLUMN user_tag_partner_id SET NOT NULL;

ALTER TABLE IF EXISTS public.bank
    ALTER COLUMN bank_name_phonetic SET NOT NULL;

ALTER TABLE IF EXISTS public.bank_branch
    ALTER COLUMN bank_branch_phonetic_name SET NOT NULL,
    ALTER COLUMN bank_branch_code TYPE TEXT USING bank_branch_code::text;
    