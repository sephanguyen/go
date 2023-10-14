CREATE TABLE IF NOT EXISTS public.fourkeys_commit_data (
    commit_hash text NOT NULL,
    author_email text NOT NULL,
    feature text NOT NULL,
    folder text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deployed_at timestamp with time zone NOT NULL,
    CONSTRAINT fourkeys_commit_feature_folder PRIMARY KEY (commit_hash,feature,folder)
);

CREATE TABLE IF NOT EXISTS public.fourkeys_feature_data (
    feature text NOT NULL,
    folder text NOT NULL,
    last_deployed_commit_hash text NOT NULL,
    CONSTRAINT fourkeys_feature_folder UNIQUE (feature,folder)
);