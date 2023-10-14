CREATE TYPE rating_type AS ENUM ('neutral', 'positive', 'negative');

CREATE TABLE IF NOT EXISTS lo_video_rating (
    lo_id text NOT NULL,
    video_id text NOT NULL,
    learner_id text NOT NULL,
    rating_value rating_type NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT lo_video_rating_pk PRIMARY KEY (lo_id, video_id)
);

CREATE POLICY rls_lo_video_rating ON "lo_video_rating"
USING (permission_check(resource_path, 'lo_video_rating'))
WITH CHECK (permission_check(resource_path, 'lo_video_rating'));

CREATE POLICY rls_lo_video_rating_restrictive ON "lo_video_rating"
AS RESTRICTIVE TO public
USING (permission_check(resource_path, 'lo_video_rating'))
WITH CHECK (permission_check(resource_path, 'lo_video_rating'));

ALTER TABLE "lo_video_rating" ENABLE ROW LEVEL security;
ALTER TABLE "lo_video_rating" FORCE ROW LEVEL security;
