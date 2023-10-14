ALTER TABLE public.system_notifications ADD COLUMN IF NOT EXISTS status TEXT DEFAULT 'SYSTEM_NOTIFICATION_STATUS_NEW';

ALTER TABLE ONLY public.system_notifications DROP CONSTRAINT IF EXISTS system_notification_status_type_check;
ALTER TABLE public.system_notifications ADD CONSTRAINT system_notification_status_type_check CHECK (status = ANY (ARRAY[
		'SYSTEM_NOTIFICATION_STATUS_NEW',
		'SYSTEM_NOTIFICATION_STATUS_DONE'
]::TEXT[]));
