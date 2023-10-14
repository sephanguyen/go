ALTER TABLE ONLY public.conversion_tasks
  DROP CONSTRAINT IF EXISTS conversion_task_resource_url_un;

ALTER TABLE public.conversion_tasks
  ALTER COLUMN status TYPE varchar(255) USING status::varchar;
