ALTER TABLE IF EXISTS public.lesson_members_states
    ALTER COLUMN resource_path SET DEFAULT autofillresourcepath(),
    ALTER COLUMN resource_path SET NOT NULL;

ALTER TABLE IF EXISTS public.lesson_room_states
    ALTER COLUMN resource_path SET NOT NULL;
