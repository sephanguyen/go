ALTER TABLE public.media ADD COLUMN IF NOT EXISTS converted_images JSONB;

CREATE TABLE IF NOT EXISTS public.conversion_tasks (
    task_uuid text NOT NULL PRIMARY KEY,
    resource_url text NOT NULL,
    status varchar(20) NOT NULL DEFAULT 'WAITING',
    conversion_response JSONB,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL
);

ALTER TABLE ONLY public.conversion_tasks
    DROP CONSTRAINT IF EXISTS conversion_task_resource_url_un;

ALTER TABLE ONLY public.conversion_tasks
    ADD CONSTRAINT conversion_task_resource_url_un UNIQUE (resource_url);
