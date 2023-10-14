ALTER TABLE public.user_group DROP CONSTRAINT IF EXISTS fk__user_group__org_location_id;

ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__user_group_id;
ALTER TABLE public.granted_role DROP CONSTRAINT IF EXISTS fk__granted_role__role_id;

ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_id;
ALTER TABLE public.user_group_member DROP CONSTRAINT IF EXISTS fk__user_group_member__user_group_id;
