CREATE TABLE IF NOT EXISTS public.applied_slot (
	id int4 NOT NULL,
	"year" int4 NULL,
	"period" int4 NULL,
	center_num int4 NULL,
	student_id text NULL,
	enrollment_status int4 NULL,
	grade int4 NULL,
	student_name text NULL,
	applied_slot int4 NULL,
	literature_slot int4 NULL,
	math_slot int4 NULL,
	en_slot int4 NULL,
	science_slot int4 NULL,
	social_science_slot int4 NULL,
	other_slot_1 int4 NULL,
	other_slot_2 int4 NULL,
	other_slot_3 int4 NULL,
	other_slot_4 int4 NULL,
	other_slot_5 int4 NULL,
	other_slot_6 int4 NULL,
	other_slot_7 int4 NULL,
	other_slot_8 int4 NULL,
	other_slot_9 int4 NULL,
	other_slot_10 int4 NULL,
	sd_literature_slot int4 NULL,
	sd_math_slot int4 NULL,
	sd_en_slot int4 NULL,
	sd_science_slot int4 NULL,
	sd_social_slot int4 NULL,
	sd_other_slot_1 int4 NULL,
	sd_other_slot_2 int4 NULL,
	sd_other_slot_3 int4 NULL,
	sd_other_slot_4 int4 NULL,
	sd_other_slot_5 int4 NULL,
	sd_other_slot_6 int4 NULL,
	sd_other_slot_7 int4 NULL,
	sd_other_slot_8 int4 NULL,
	sd_other_slot_9 int4 NULL,
	sd_other_slot_10 int4 NULL,
	preferred_gender int4 NULL,
	sibling_should_be_same_time int4 NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT applied_slot_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.center_opening_slot (
	id int4 NOT NULL,
	"year" int4 NULL,
	"period" int4 NULL,
	center_num int4 NULL,
	"date" date NULL,
	time_period int4 NULL,
	open_or_not bool NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT center_opening_slot_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.student_available_slot_master (
	id int4 NOT NULL,
	"year" int4 NULL,
	"period" int4 NULL,
	center_num int4 NULL,
	student_id text NULL,
	"date" date NULL,
	time_period int4 NULL,
	open_or_not bool NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT student_available_slot_master_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.teacher_available_slot_master (
	id int4 NOT NULL,
	"year" int4 NULL,
	"period" int4 NULL,
	center_num int4 NULL,
	teacher_id text NULL,
	"date" date NULL,
	time_period int4 NULL,
	open_or_not bool NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT teacher_available_slot_master_pk PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.teacher_subject (
	id int4 NOT NULL,
	teacher_id text NULL,
	grade_div int4 NULL,
	subject_id int4 NULL,
	available_or_not bool NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	CONSTRAINT teacher_subject_pk PRIMARY KEY (id)
);

CREATE POLICY rls_applied_slot ON "applied_slot" using (permission_check(resource_path, 'applied_slot')) with check (permission_check(resource_path, 'applied_slot'));
CREATE POLICY rls_applied_slot_restrictive ON "applied_slot" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'applied_slot')) with check (permission_check(resource_path, 'applied_slot'));

CREATE POLICY rls_center_opening_slot ON "center_opening_slot" using (permission_check(resource_path, 'center_opening_slot')) with check (permission_check(resource_path, 'center_opening_slot'));
CREATE POLICY rls_center_opening_slot_restrictive ON "center_opening_slot" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'center_opening_slot')) with check (permission_check(resource_path, 'center_opening_slot'));

CREATE POLICY rls_student_available_slot_master ON "student_available_slot_master" using (permission_check(resource_path, 'student_available_slot_master')) with check (permission_check(resource_path, 'student_available_slot_master'));
CREATE POLICY rls_student_available_slot_master_restrictive ON "student_available_slot_master" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'student_available_slot_master')) with check (permission_check(resource_path, 'student_available_slot_master'));

CREATE POLICY rls_teacher_available_slot_master ON "teacher_available_slot_master" using (permission_check(resource_path, 'teacher_available_slot_master')) with check (permission_check(resource_path, 'teacher_available_slot_master'));
CREATE POLICY rls_teacher_available_slot_master_restrictive ON "teacher_available_slot_master" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'teacher_available_slot_master')) with check (permission_check(resource_path, 'teacher_available_slot_master'));

CREATE POLICY rls_teacher_subject ON "teacher_subject" using (permission_check(resource_path, 'teacher_subject')) with check (permission_check(resource_path, 'teacher_subject'));
CREATE POLICY rls_teacher_subject_restrictive ON "teacher_subject" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'teacher_subject')) with check (permission_check(resource_path, 'teacher_subject'));

ALTER TABLE "applied_slot" ENABLE ROW LEVEL security;
ALTER TABLE "applied_slot" FORCE ROW LEVEL security;

ALTER TABLE "center_opening_slot" ENABLE ROW LEVEL security;
ALTER TABLE "center_opening_slot" FORCE ROW LEVEL security;

ALTER TABLE "student_available_slot_master" ENABLE ROW LEVEL security;
ALTER TABLE "student_available_slot_master" FORCE ROW LEVEL security;

ALTER TABLE "teacher_available_slot_master" ENABLE ROW LEVEL security;
ALTER TABLE "teacher_available_slot_master" FORCE ROW LEVEL security;

ALTER TABLE "teacher_subject" ENABLE ROW LEVEL security;
ALTER TABLE "teacher_subject" FORCE ROW LEVEL security;
