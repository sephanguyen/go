CREATE TABLE IF NOT EXISTS public.lesson_members_states (
    "lesson_id" TEXT NOT NULL,
    "user_id" TEXT NOT NULL,
    "state_type" TEXT NOT NULL,
    "created_at" timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    "updated_at" timestamp with time zone,
    "deleted_at" timestamp with time zone,
    "bool_value" BOOLEAN,
    "resource_path" TEXT,
    string_array_value TEXT[] DEFAULT NULL,
    CONSTRAINT lesson_id_fk FOREIGN KEY (lesson_id, user_id) REFERENCES public.lesson_members(lesson_id, user_id),
    CONSTRAINT lesson_members_states_pk PRIMARY KEY (lesson_id, user_id, state_type)
);
DROP POLICY IF EXISTS rls_lesson_members_states ON public.lesson_members_states;
CREATE POLICY rls_lesson_members_states ON "lesson_members_states" USING (permission_check(resource_path, 'lesson_members_states'::text)) WITH CHECK (permission_check(resource_path, 'lesson_members_states'::text));


DROP POLICY IF EXISTS rls_lesson_members_states_restrictive ON public.lesson_members_states;
CREATE POLICY rls_lesson_members_states_restrictive ON "lesson_members_states" AS RESTRICTIVE TO PUBLIC USING (permission_check(resource_path, 'lesson_members_states'::text)) WITH CHECK (permission_check(resource_path, 'lesson_members_states'::text));

ALTER TABLE "lesson_members_states" ENABLE ROW LEVEL security;
ALTER TABLE "lesson_members_states" FORCE ROW LEVEL security;
