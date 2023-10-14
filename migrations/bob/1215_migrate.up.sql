CREATE SEQUENCE IF NOT EXISTS import_user_event__import_user_event_id__seq;
DROP FUNCTION IF EXISTS import_user_event__import_user_event_id__next;

--Return next val based on sequence
--Also sharding the id
CREATE FUNCTION import_user_event__import_user_event_id__next(OUT result bigint) RETURNS bigint
    LANGUAGE plpgsql
AS
$$
DECLARE
    now_millis bigint;
    shard_id   bigint;
    seq_id     bigint;
BEGIN
    SELECT nextval('import_user_event__import_user_event_id__seq') % 1024 INTO seq_id;
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    SELECT current_setting('database.shard_id')::bigint INTO shard_id;

    result := (SELECT generate_sharded_id(now_millis, shard_id, seq_id));
END
$$;

CREATE TABLE IF NOT EXISTS public.import_user_event (
    import_user_event_id BIGINT DEFAULT import_user_event__import_user_event_id__next(),
    user_id TEXT NOT NULL,
    status TEXT NOT NULL,
    payload JSONB NOT NULL,
    importer_id TEXT NOT NULL,
    sequence_number BIGINT,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path TEXT DEFAULT autofillresourcepath(),
    CONSTRAINT pk__import_user_event PRIMARY KEY (import_user_event_id),
    CONSTRAINT fk__import_user_event__user_id FOREIGN KEY (user_id) REFERENCES public.users(user_id),
    CONSTRAINT fk__import_user_event__importer_id FOREIGN KEY (importer_id) REFERENCES public.users(user_id)
);
CREATE POLICY rls_import_user_event ON public.import_user_event using (permission_check(resource_path, 'import_user_event')) with check (permission_check(resource_path, 'import_user_event'));

ALTER TABLE public.import_user_event ENABLE ROW LEVEL security;
ALTER TABLE public.import_user_event FORCE ROW LEVEL security;

ALTER SEQUENCE import_user_event__import_user_event_id__seq OWNED BY import_user_event.import_user_event_id;

CREATE INDEX IF NOT EXISTS import_user_event__user_id__idx ON import_user_event (user_id);
