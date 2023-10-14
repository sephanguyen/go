
-- lo_progression table
CREATE TABLE IF NOT EXISTS public.lo_progression (
	progression_id TEXT NOT NULL,
	shuffled_quiz_set_id TEXT NOT NULL,

	student_id TEXT NOT NULL,
	study_plan_id TEXT NOT NULL,
	learning_material_id TEXT NOT NULL,

	quiz_external_ids TEXT[],
    last_index INT4 NOT NULL,

	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NOT NULL,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT lo_progression_pk PRIMARY KEY (progression_id)
);
CREATE INDEX IF NOT EXISTS lo_progression_study_plan_item_identity_idx 
ON public.lo_progression(student_id, study_plan_id, learning_material_id);

/* set RLS */
CREATE POLICY rls_lo_progression ON "lo_progression" using (
    permission_check(resource_path, 'lo_progression')
) with check (
    permission_check(resource_path, 'lo_progression')
);

CREATE POLICY rls_lo_progression_restrictive ON "lo_progression" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lo_progression')
) with check (
    permission_check(resource_path, 'lo_progression')
);

ALTER TABLE IF EXISTS "lo_progression" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "lo_progression" FORCE ROW LEVEL security;


-- lo_progression_answer table
CREATE TABLE IF NOT EXISTS public.lo_progression_answer (
	progression_answer_id TEXT NOT NULL,
	shuffled_quiz_set_id TEXT NOT NULL,
	quiz_external_id TEXT NOT NULL,
	progression_id TEXT NOT NULL,
	
	student_id TEXT NOT NULL,
	study_plan_id TEXT NOT NULL,
	learning_material_id TEXT NOT NULL,
	
	student_text_answer TEXT[] NULL,
	student_index_answer INT4[] NULL,

	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE NULL,
	resource_path TEXT NOT NULL DEFAULT autofillresourcepath(),

	CONSTRAINT lo_progression_answer_pk PRIMARY KEY (progression_answer_id),
	CONSTRAINT lo_progression_answer_fk FOREIGN KEY(progression_id) REFERENCES public.lo_progression(progression_id),
	CONSTRAINT lo_progression_answer_un UNIQUE (progression_id, quiz_external_id)
);

/* set RLS */
CREATE POLICY rls_lo_progression_answer ON "lo_progression_answer" using (
    permission_check(resource_path, 'lo_progression_answer')
) with check (
    permission_check(resource_path, 'lo_progression_answer')
);

CREATE POLICY rls_lo_progression_answer_restrictive ON "lo_progression_answer" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lo_progression_answer')
) with check (
    permission_check(resource_path, 'lo_progression_answer')
);

ALTER TABLE IF EXISTS "lo_progression_answer" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "lo_progression_answer" FORCE ROW LEVEL security;