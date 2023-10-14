CREATE TABLE IF NOT EXISTS public.tags (
    tag_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    is_archived bool DEFAULT false,

    CONSTRAINT pk__tags PRIMARY KEY (tag_id)
);

ALTER PUBLICATION kec_publication ADD TABLE public.tags;
