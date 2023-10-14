CREATE TABLE IF NOT EXISTS public.assign_study_plan_tasks (
	id text NOT NULL,
	study_plan_ids text[] NOT NULL,
	status text,
	course_id text,

    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,

	CONSTRAINT assign_study_plan_tasks_pk PRIMARY KEY (id)
);
