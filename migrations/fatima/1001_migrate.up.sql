CREATE TABLE IF NOT EXISTS public.packages (
    package_id TEXT NOT NULL,
    country text,
    name text NOT NULL,
    descriptions text[],
    price integer NOT NULL,
    discounted_price integer,
    start_at timestamp with time zone,
    end_at timestamp with time zone,
    duration integer, -- by day, if null with read from start_at, end_at
    prioritize_level integer DEFAULT 0,
    properties JSONB NOT NULL, -- format like this https://github.com/manabie-com/manabie-online/issues/2283#issuecomment-683614940
    is_recommended boolean NOT NULL,
    is_active boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__packages PRIMARY KEY (package_id)
);

CREATE TABLE IF NOT EXISTS public.student_packages (
    student_package_id TEXT NOT NULL,
    student_id TEXT NOT NULL,
    package_id TEXT NOT NULL,
    start_at timestamp with time zone NOT NULL,
    end_at timestamp with time zone NOT NULL,
    properties JSONB NOT NULL, -- format like this https://github.com/manabie-com/manabie-online/issues/2283#issuecomment-683614940
    is_active boolean NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__student_packages PRIMARY KEY (student_package_id),
    CONSTRAINT fk__student_packages__package_id FOREIGN KEY (package_id) REFERENCES public.packages (package_id)
);

CREATE INDEX idx__student_packages__student_id__start_at__end_at ON public.student_packages USING btree (student_id, start_at, end_at);

