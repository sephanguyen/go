CREATE TABLE IF NOT EXISTS public.parents (
    parent_id text NOT NULL,
    school_id int,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);
