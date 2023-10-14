CREATE TABLE IF NOT EXISTS prefecture (
    prefecture_code       TEXT NOT NULL,
    country               TEXT NOT NULL,
    name                  TEXT NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,

    CONSTRAINT prefecture__pk PRIMARY KEY (prefecture_code)
);

ALTER TABLE user_address RENAME prefecture TO prefecture_code;
ALTER TABLE user_address ADD CONSTRAINT user_address__prefecture_code__fk FOREIGN KEY (prefecture_code) REFERENCES public.prefecture(prefecture_code);

