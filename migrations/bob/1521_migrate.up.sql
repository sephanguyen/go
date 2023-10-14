ALTER TABLE public.notification_class_filter
    ALTER COLUMN deleted_at SET DEFAULT NULL;

ALTER TABLE public.notification_course_filter
    ALTER COLUMN deleted_at SET DEFAULT NULL;

ALTER TABLE public.notification_location_filter
    ALTER COLUMN deleted_at SET DEFAULT NULL;