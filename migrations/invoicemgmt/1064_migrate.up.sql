CREATE TABLE IF NOT EXISTS public.prefecture (
    prefecture_id text NOT NULL,
    prefecture_code text NOT NULL,
    country text NOT NULL,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamptz NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamptz NULL,

    CONSTRAINT prefecture__pk PRIMARY KEY (prefecture_id)
);