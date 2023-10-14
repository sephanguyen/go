CREATE SEQUENCE IF NOT EXISTS public.user_upsertion_event__user_upsertion_event_id__seq AS integer;

DROP FUNCTION IF EXISTS user_upsertion_event__user_upsertion_event_id__next;

--Return next val based on sequence
--Also sharding the id
CREATE FUNCTION user_upsertion_event__user_upsertion_event_id__next(OUT result bigint) RETURNS bigint
    LANGUAGE plpgsql
AS
$$
DECLARE
    now_millis bigint;
    shard_id   bigint;
    seq_id     bigint;
BEGIN
    SELECT nextval('user_upsertion_event__user_upsertion_event_id__seq') % 1024 INTO seq_id;
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    SELECT current_setting('database.shard_id')::bigint INTO shard_id;

    result := (SELECT generate_sharded_id(now_millis, shard_id, seq_id));
END
$$;

CREATE TABLE IF NOT EXISTS public.user_upsertion_event (
    user_upsertion_event_id bigint DEFAULT user_upsertion_event__user_upsertion_event_id__next(),
    user_id text,
    event_type text,
    action_type text,
    status text,
    message text,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc' :: text, now()),
    deleted_at timestamp with time zone,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT user_upsertion_event_pk PRIMARY KEY (user_upsertion_event_id),
    CONSTRAINT user_upsertion_event__user_id__event_type__action_type__unique UNIQUE (user_id, event_type, action_type)
);

ALTER SEQUENCE public.user_upsertion_event__user_upsertion_event_id__seq OWNED BY public.user_upsertion_event.user_upsertion_event_id;

CREATE POLICY rls_user_upsertion_event ON "user_upsertion_event"
USING (permission_check(resource_path, 'user_upsertion_event')) WITH CHECK (permission_check(resource_path, 'user_upsertion_event'));
CREATE POLICY rls_user_upsertion_event_restrictive ON "user_upsertion_event" AS RESTRICTIVE
USING (permission_check(resource_path, 'user_upsertion_event'))WITH CHECK (permission_check(resource_path, 'user_upsertion_event'));

ALTER TABLE "user_upsertion_event" ENABLE ROW LEVEL security;
ALTER TABLE "user_upsertion_event" FORCE ROW LEVEL security;
