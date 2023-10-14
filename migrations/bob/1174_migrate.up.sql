CREATE SEQUENCE IF NOT EXISTS usr_email__import_id__seq;

DROP FUNCTION IF EXISTS generate_sharded_id;

CREATE FUNCTION generate_sharded_id(millis bigint, shardID bigint, sequenceVal bigint, OUT result bigint) RETURNS bigint
    LANGUAGE plpgsql
AS
$$
BEGIN
    result := millis << 21;
    result := result | (shardID << 10);
    result := result | (sequenceVal << 0);
END
$$;

DROP FUNCTION IF EXISTS usr_email__import_id__next;

--Return next val based on sequence
--Also sharding the id
CREATE FUNCTION usr_email__import_id__next(OUT result bigint) RETURNS bigint
    LANGUAGE plpgsql
AS
$$
DECLARE
    now_millis bigint;
    shard_id   bigint;
    seq_id     bigint;
BEGIN
    SELECT nextval('usr_email__import_id__seq') % 1024 INTO seq_id;
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    SELECT current_setting('database.shard_id')::bigint INTO shard_id;

    result := (SELECT generate_sharded_id(now_millis, shard_id, seq_id));
END
$$;

CREATE TABLE IF NOT EXISTS usr_email
(
    email         text                      NOT NULL,
    usr_id        text,
    create_at     timestamptz               NOT NULL,
    updated_at    timestamptz default now() NOT NULL,
    delete_at     timestamptz,
    resource_path text        default autofillresourcepath(),
    import_id     bigint      default usr_email__import_id__next(),

    CONSTRAINT usr_email__pkey PRIMARY KEY (email),

    --Prevent emails have upper-case and mixed-case characters
    CONSTRAINT usr_email__email__lower_case__check CHECK (email = lower(email)),
    --Prevent emails have leading and trailing spaces
    CONSTRAINT usr_email__email__leading_and_trailing_spaces__check CHECK (email = (trim(BOTH FROM email))),
    --Prevent emails have empty values
    CONSTRAINT usr_email__email__empty__check CHECK ('' != (trim(BOTH FROM email)))
);

CREATE POLICY rls_usr_email ON "usr_email" using (permission_check(resource_path, 'usr_email')) with check (permission_check(resource_path, 'usr_email'));
ALTER TABLE "usr_email"
    ENABLE ROW LEVEL security;
ALTER TABLE "usr_email"
    FORCE ROW LEVEL security;

ALTER SEQUENCE usr_email__import_id__seq OWNED BY usr_email.import_id;

CREATE INDEX IF NOT EXISTS usr_email__usr_id__idx ON usr_email (usr_id);
