
-- add not null constraint for table info_notifications
ALTER TABLE public.info_notifications ALTER COLUMN type SET NOT NULL;

-- add not null constraint for users_info_notifications
ALTER TABLE public.users_info_notifications ALTER COLUMN notification_id SET NOT NULL;
ALTER TABLE public.users_info_notifications ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE public.users_info_notifications ALTER COLUMN status SET NOT NULL;

-- add new column owner 
ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS owner INTEGER NOT NULL;

-- add new column courses and current grade for users_info_notifications
ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS course_ids TEXT[];
ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS current_grade SMALLINT;
ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS is_individual BOOLEAN;

-- add foreign key for info_notifications
ALTER TABLE public.info_notifications ADD CONSTRAINT fk__info_notification__editor_id FOREIGN KEY(editor_id) REFERENCES users(user_id);
ALTER TABLE public.info_notifications ADD CONSTRAINT fk__info_notification__school_id FOREIGN KEY(owner) REFERENCES schools(school_id);
ALTER TABLE public.users_info_notifications ADD CONSTRAINT unique__user_id__notification_id UNIQUE(user_id, notification_id);

ALTER TABLE ONLY public.info_notifications DROP CONSTRAINT IF EXISTS notification_type_check;
ALTER TABLE public.info_notifications
    ADD CONSTRAINT notification_type_check CHECK (type = ANY (ARRAY[
		'NOTIFICATION_TYPE_NONE',
		'NOTIFICATION_TYPE_TEXT',
		'NOTIFICATION_TYPE_PROMO_CODE',
		'NOTIFICATION_TYPE_ASSIGNMENT',
		'NOTIFICATION_TYPE_COMPOSED'
]::text[]));


ALTER TABLE ONLY public.info_notifications DROP CONSTRAINT IF EXISTS notification_status_type_check;
ALTER TABLE public.info_notifications
    ADD CONSTRAINT notification_status_type_check CHECK (status = ANY (ARRAY[
		'NOTIFICATION_STATUS_NONE',
		'NOTIFICATION_STATUS_DRAFT',
		'NOTIFICATION_STATUS_SCHEDULED',
		'NOTIFICATION_STATUS_SENT',
		'NOTIFICATION_STATUS_DISCARD'
]::text[]));


ALTER TABLE ONLY public.info_notifications DROP CONSTRAINT IF EXISTS notification_event_type_check;
ALTER TABLE public.info_notifications
    ADD CONSTRAINT notification_event_type_check CHECK (event = ANY (ARRAY[
		'NOTIFICATION_EVENT_NONE',
		'NOTIFICATION_EVENT_X_LO_COMPLETED',
		'NOTIFICATION_EVENT_TEACHER_GIVE_ASSIGNMENT',
		'NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT',
		'NOTIFICATION_EVENT_STUDENT_SUBMIT_ASSIGNMENT',
		'NOTIFICATION_EVENT_ASSIGNMENT_UPDATED'
]::text[]));


ALTER TABLE ONLY public.users_info_notifications DROP CONSTRAINT IF EXISTS users_notification_status_type_check;
ALTER TABLE public.users_info_notifications
    ADD CONSTRAINT users_notification_status_type_check CHECK (status = ANY (ARRAY[
		'USER_NOTIFICATION_STATUS_NONE',
		'USER_NOTIFICATION_STATUS_NEW',
		'USER_NOTIFICATION_STATUS_SEEN',
		'USER_NOTIFICATION_STATUS_READ',
		'USER_NOTIFICATION_STATUS_FAILED'
]::text[]));