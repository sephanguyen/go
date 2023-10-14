-- public.lesson_student_subscriptions definition

-- Drop table

-- DROP TABLE public.lesson_student_subscriptions;

CREATE TABLE IF NOT EXISTS public.lesson_student_subscriptions (
	student_subscription_id text NOT NULL,
	course_id text NOT NULL,
	student_id text NOT NULL,
	subscription_id text NOT NULL,
	start_at timestamptz NULL,
	end_at timestamptz NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	course_slot int4 NULL,
	course_slot_per_week int4 NULL,
	student_first_name text NULL,
	student_last_name text NULL,
	package_type text NULL,
	CONSTRAINT lesson_student_subscriptions_pkey PRIMARY KEY (student_subscription_id),
	CONSTRAINT lesson_student_subscriptions_uniq UNIQUE (subscription_id, course_id, student_id)
);
CREATE INDEX lesson_student_subscripton__package_type__idx ON public.lesson_student_subscriptions USING btree (package_type);


CREATE POLICY rls_lesson_student_subscriptions ON "lesson_student_subscriptions" USING (permission_check(resource_path, 'lesson_student_subscriptions')) WITH CHECK (permission_check(resource_path, 'lesson_student_subscriptions'));
CREATE POLICY rls_lesson_student_subscriptions_restrictive ON "lesson_student_subscriptions" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'lesson_student_subscriptions')) with check (permission_check(resource_path, 'lesson_student_subscriptions'));

ALTER TABLE "lesson_student_subscriptions" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_student_subscriptions" FORCE ROW LEVEL security;