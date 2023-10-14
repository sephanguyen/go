ALTER TABLE public.user_basic_info
    RENAME COLUMN first_phonetic_name TO first_name_phonetic;

ALTER TABLE public.user_basic_info
    RENAME COLUMN last_phonetic_name TO last_name_phonetic;

ALTER TABLE public.user_basic_info
    RENAME COLUMN full_phonetic_name TO full_name_phonetic;

ALTER TABLE public.user_basic_info ALTER COLUMN current_grade TYPE smallint USING (current_grade::smallint);
