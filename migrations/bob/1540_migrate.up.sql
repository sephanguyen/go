-- course location access path
ALTER TABLE public.course_access_paths
    ADD COLUMN IF NOT EXISTS
    id VARCHAR(36) NOT NULL
    DEFAULT generate_ulid();
-- DEFAULT generate_ulid() will generate a new ulid for all old records.

ALTER TABLE public.course_access_paths
    ADD CONSTRAINT cap_unique_id UNIQUE (id);

