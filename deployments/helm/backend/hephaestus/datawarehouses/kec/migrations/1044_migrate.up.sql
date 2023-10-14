CREATE TABLE IF NOT EXISTS public.prefecture (
    prefecture_id TEXT,
    prefecture_code TEXT,
    country TEXT,
    name TEXT,
    created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,

    CONSTRAINT pk__prefecture PRIMARY KEY (prefecture_id)
);

ALTER PUBLICATION kec_publication ADD TABLE public.prefecture;
