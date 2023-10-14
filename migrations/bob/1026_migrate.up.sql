CREATE TABLE IF NOT EXISTS public.media (
    media_id text NOT NULL,
    name text,
    resource text,
    comments jsonb,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    type text,
    CONSTRAINT media_pk PRIMARY KEY (media_id)
);
