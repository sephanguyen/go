CREATE TABLE IF NOT EXISTS bob.prefecture (
    prefecture_id TEXT,
    prefecture_code TEXT,
    country TEXT,
    name TEXT,
    created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,

    CONSTRAINT pk__prefecture PRIMARY KEY (prefecture_id)
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE bob.prefecture;
