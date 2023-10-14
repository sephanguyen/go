CREATE TABLE IF NOT EXISTS bob.tags (
    tag_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at TIMESTAMP WITH TIME ZONE,
    is_archived bool DEFAULT false,
    resource_path TEXT,

    CONSTRAINT pk__tags PRIMARY KEY (tag_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.tags;
