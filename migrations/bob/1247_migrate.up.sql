ALTER TABLE public.school_info
    ADD COLUMN IF NOT EXISTS school_partner_id TEXT,
    ADD CONSTRAINT school_info__school_partner_id__unique UNIQUE (school_partner_id, resource_path);

ALTER TABLE public.school_course
    ADD COLUMN IF NOT EXISTS school_course_partner_id TEXT,
    ADD CONSTRAINT school_course__school_course_partner_id__unique UNIQUE (school_course_partner_id, resource_path);

ALTER TABLE public.user_tag
    ADD COLUMN IF NOT EXISTS user_tag_partner_id TEXT,
    ADD CONSTRAINT user_tag__user_tag_partner_id__unique UNIQUE (user_tag_partner_id, resource_path);

ALTER TABLE public.school_level
    ALTER COLUMN sequence TYPE integer USING sequence::integer,
    ADD CONSTRAINT school_level__sequence__unique UNIQUE (sequence, resource_path);

