ALTER TABLE public.school_configs DROP CONSTRAINT IF EXISTS school_configs_privileges_check;
ALTER TABLE public.school_configs DROP CONSTRAINT IF EXISTS school_configs_privileges_check1;
ALTER TABLE public.school_configs DROP CONSTRAINT IF EXISTS school_configs_privileges_check2;

ALTER TABLE public.school_configs
ADD CONSTRAINT school_configs_privileges_check CHECK (privileges <@ ARRAY[
            'CAN_ACCESS_LEARNING_TOPICS',
            'CAN_ACCESS_PRACTICE_TOPICS',
            'CAN_ACCESS_MOCK_EXAMS',
            'CAN_ACCESS_ALL_LOS',
            'CAN_ACCESS_SOME_LOS',
            'CAN_WATCH_VIDEOS',
            'CAN_READ_STUDY_GUIDES',
            'CAN_SKIP_VIDEOS',
            'CAN_CHAT_WITH_TEACHER'
        ]); 
