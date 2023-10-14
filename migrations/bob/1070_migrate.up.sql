ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS user_group TEXT;
ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS parent_id TEXT;
ALTER TABLE public.users_info_notifications ADD COLUMN IF NOT EXISTS student_id TEXT;

ALTER TABLE public.parents ADD CONSTRAINT parents__parent_id_pk PRIMARY KEY (parent_id);
ALTER TABLE public.users_info_notifications DROP CONSTRAINT IF EXISTS users_info_notifications__parent_id_fk;
ALTER TABLE public.users_info_notifications 
ADD CONSTRAINT users_info_notifications__parent_id_fk FOREIGN KEY (parent_id) REFERENCES public.parents(parent_id);

ALTER TABLE public.users_info_notifications DROP CONSTRAINT IF EXISTS users_info_notifications__student_id_fk;
ALTER TABLE public.users_info_notifications 
ADD CONSTRAINT users_info_notifications__student_id_fk FOREIGN KEY (student_id) REFERENCES public.students(student_id);


ALTER TABLE public.users_info_notifications DROP CONSTRAINT IF EXISTS unique__user_id__notification_id; 
ALTER TABLE public.users_info_notifications ADD CONSTRAINT unique__user_id__notification_id UNIQUE(user_id, notification_id, parent_id,student_id);