ALTER TABLE public.students
  ADD COLUMN IF NOT EXISTS additional_data JSONB;

INSERT INTO public.configs
(config_key, config_group, country, config_value, updated_at, created_at, deleted_at)
VALUES
('planName', 'class_plan', 'COUNTRY_JP', 'School', NOW(), NOW(), NULL),
('planPeriod', 'class_plan', 'COUNTRY_JP', 3000, NOW(), NOW(), NULL)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS public.jpref_sync_data_logs (
    jpref_sync_data_log_id text NOT NULL,
    signature text NOT NULL,
    payload JSONB NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__jpref_sync_data_logs PRIMARY KEY (jpref_sync_data_log_id)
);

CREATE TABLE IF NOT EXISTS public.lesson_members (
    lesson_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT pk__lesson_members PRIMARY KEY (lesson_id,user_id),
    CONSTRAINT fk__lesson_members__lesson_id FOREIGN KEY (lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT fk__lesson_members__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);


INSERT INTO public.schools
(school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge)
VALUES(-2147483647, 'Manabie School', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false)
ON CONFLICT DO NOTHING;
