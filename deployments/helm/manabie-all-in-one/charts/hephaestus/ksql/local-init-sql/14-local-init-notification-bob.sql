\connect bob;

INSERT INTO public.tags
(tag_id, tag_name, created_at, updated_at, deleted_at, is_archived, resource_path)
VALUES('01GV032YZ8FF4JGEAM4XXQX6L8', 'Notification Tag DWH test', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, false, '-2147483642');
