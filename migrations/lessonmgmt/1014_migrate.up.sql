CREATE TABLE IF NOT EXISTS public.courses (
	course_id text NOT NULL,
	name text NOT NULL,
	country text,
	subject text,
	grade smallint,
    display_order smallint,
	updated_at timestamp WITH TIME ZONE NOT NULL,
	created_at timestamp WITH TIME ZONE NOT NULL,
    deleted_at timestamp WITH TIME ZONE,
	school_id integer NOT NULL,
	course_type text,
	start_date timestamp WITH TIME ZONE,
	end_date timestamp WITH TIME ZONE,
	teacher_ids text[],
	preset_study_plan_id text,
	icon text,
	status text,
	resource_path text DEFAULT autofillresourcepath(),
	teaching_method text,
	CONSTRAINT courses_pk PRIMARY KEY (course_id)
);


CREATE POLICY rls_courses ON "courses" USING (permission_check(resource_path, 'courses')) WITH CHECK (permission_check(resource_path, 'courses'));
CREATE POLICY rls_courses_restrictive ON "courses" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'courses')) with check (permission_check(resource_path, 'courses'));

ALTER TABLE "courses" ENABLE ROW LEVEL security;
ALTER TABLE "courses" FORCE ROW LEVEL security;
