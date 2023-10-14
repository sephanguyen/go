-- create table withus_mapping_course_id
CREATE TABLE IF NOT EXISTS public.withus_mapping_course_id (
    manabie_course_id text NOT NULL,
    withus_course_id text NOT NULL DEFAULT '',
    created_date timestamp with time zone,
    created_by text,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT withus_mapping_course_id_pk PRIMARY KEY (manabie_course_id)
);

/* set RLS */
CREATE POLICY rls_withus_mapping_course_id ON "withus_mapping_course_id" using (
    permission_check(resource_path, 'withus_mapping_course_id')
) with check (
    permission_check(resource_path, 'withus_mapping_course_id')
);

CREATE POLICY rls_withus_mapping_course_id_restrictive ON "withus_mapping_course_id" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'withus_mapping_course_id')
) with check (
    permission_check(resource_path, 'withus_mapping_course_id')
);

ALTER TABLE IF EXISTS "withus_mapping_course_id" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "withus_mapping_course_id" FORCE ROW LEVEL security;

-- trigger to migrate data from course_students to withus_mapping_course_id
CREATE OR REPLACE FUNCTION public.migrate_withus_mapping_course_id() 
RETURNS TRIGGER
LANGUAGE plpgsql
AS $BODY$
BEGIN
    INSERT INTO public.withus_mapping_course_id (
        manabie_course_id,
        resource_path
    )
    SELECT 
        NEW.course_id,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_course_id_pk DO NOTHING;
RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_withus_mapping_course_id ON public.course_students;
CREATE TRIGGER migrate_withus_mapping_course_id
AFTER INSERT ON public.course_students
FOR EACH ROW
EXECUTE FUNCTION public.migrate_withus_mapping_course_id();

-- create table withus_mapping_question_tag
CREATE TABLE IF NOT EXISTS public.withus_mapping_question_tag (
    manabie_tag_id text NOT NULL,
    manabie_tag_name text NOT NULL,
    withus_tag_name text NOT NULL DEFAULT '',
    created_date timestamp with time zone,
    created_by text,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT withus_mapping_question_tag_pk PRIMARY KEY (manabie_tag_id)
);

/* set RLS */
CREATE POLICY rls_withus_mapping_question_tag ON "withus_mapping_question_tag" using (
    permission_check(resource_path, 'withus_mapping_question_tag')
) with check (
    permission_check(resource_path, 'withus_mapping_question_tag')
);

CREATE POLICY rls_withus_mapping_question_tag_restrictive ON "withus_mapping_question_tag" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'withus_mapping_question_tag')
) with check (
    permission_check(resource_path, 'withus_mapping_question_tag')
);

ALTER TABLE IF EXISTS "withus_mapping_question_tag" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "withus_mapping_question_tag" FORCE ROW LEVEL security;

-- trigger to migrate data from question_tag to withus_mapping_question_tag
CREATE OR REPLACE FUNCTION public.migrate_withus_mapping_question_tag()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $BODY$
BEGIN
    INSERT INTO public.withus_mapping_question_tag (
        manabie_tag_id,
        manabie_tag_name,
        resource_path
    )
    SELECT 
        NEW.question_tag_id,
        NEW.name,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_question_tag_pk DO UPDATE SET
        manabie_tag_name = NEW.name;
RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_withus_mapping_question_tag ON public.question_tag;
CREATE TRIGGER migrate_withus_mapping_question_tag
AFTER INSERT OR UPDATE ON public.question_tag
FOR EACH ROW
EXECUTE FUNCTION public.migrate_withus_mapping_question_tag();

-- create table withus_mapping_exam_lo_id
CREATE TABLE IF NOT EXISTS public.withus_mapping_exam_lo_id (
    exam_lo_id text NOT NULL,
    material_code text NOT NULL DEFAULT '',
    created_date timestamp with time zone,
    created_by text,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT withus_mapping_exam_lo_id_pk PRIMARY KEY (exam_lo_id)
);

/* set RLS */
CREATE POLICY rls_withus_mapping_exam_lo_id ON "withus_mapping_exam_lo_id" using (
    permission_check(resource_path, 'withus_mapping_exam_lo_id')
) with check (
    permission_check(resource_path, 'withus_mapping_exam_lo_id')
);

CREATE POLICY rls_withus_mapping_exam_lo_id_restrictive ON "withus_mapping_exam_lo_id" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'withus_mapping_exam_lo_id')
) with check (
    permission_check(resource_path, 'withus_mapping_exam_lo_id')
);

ALTER TABLE IF EXISTS "withus_mapping_exam_lo_id" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "withus_mapping_exam_lo_id" FORCE ROW LEVEL security;

-- trigger to migrate data from exam_lo to withus_mapping_exam_lo_id
CREATE OR REPLACE FUNCTION public.migrate_withus_mapping_exam_lo_id()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $BODY$
BEGIN
    INSERT INTO public.withus_mapping_exam_lo_id (
        exam_lo_id,
        resource_path
    )
    SELECT 
        NEW.learning_material_id,
        NEW.resource_path
    ON CONFLICT ON CONSTRAINT withus_mapping_exam_lo_id_pk DO NOTHING;
RETURN NULL;
END;
$BODY$;

DROP TRIGGER IF EXISTS migrate_withus_mapping_exam_lo_id ON public.exam_lo;
CREATE TRIGGER migrate_withus_mapping_exam_lo_id
AFTER INSERT ON public.exam_lo
FOR EACH ROW
EXECUTE FUNCTION public.migrate_withus_mapping_exam_lo_id();

-- migrate data from course_students to withus_mapping_course_id
INSERT INTO public.withus_mapping_course_id (
    manabie_course_id,
    resource_path
)
SELECT 
    course_id,
    resource_path
FROM public.course_students
ON CONFLICT ON CONSTRAINT withus_mapping_course_id_pk DO NOTHING;

-- migrate data from question_tag to withus_mapping_question_tag
INSERT INTO public.withus_mapping_question_tag (
    manabie_tag_id,
    manabie_tag_name,
    resource_path
)
SELECT 
    question_tag_id,
    name,
    resource_path
FROM public.question_tag
ON CONFLICT ON CONSTRAINT withus_mapping_question_tag_pk DO NOTHING;

-- migrate data from exam_lo to withus_mapping_exam_lo_id
INSERT INTO public.withus_mapping_exam_lo_id (
    exam_lo_id,
    resource_path
)
SELECT 
    learning_material_id,
    resource_path
FROM public.exam_lo
ON CONFLICT ON CONSTRAINT withus_mapping_exam_lo_id_pk DO NOTHING;

-- create table withus_failed_sync_email_recipient
CREATE TABLE IF NOT EXISTS public.withus_failed_sync_email_recipient (
    recipient_id text NOT NULL,
    email_address text NOT NULL,
    created_date timestamp with time zone,
    created_by text,
    last_updated_date timestamp with time zone,
    last_updated_by text,
    is_archived boolean DEFAULT false,
    resource_path text DEFAULT public.autofillresourcepath(),
    CONSTRAINT withus_failed_sync_email_recipient_pk PRIMARY KEY (recipient_id)
);

/* set RLS */
CREATE POLICY rls_withus_failed_sync_email_recipient ON "withus_failed_sync_email_recipient" using (
    permission_check(resource_path, 'withus_failed_sync_email_recipient')
) with check (
    permission_check(resource_path, 'withus_failed_sync_email_recipient')
);

CREATE POLICY rls_withus_failed_sync_email_recipient_restrictive ON "withus_failed_sync_email_recipient" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'withus_failed_sync_email_recipient')
) with check (
    permission_check(resource_path, 'withus_failed_sync_email_recipient')
);

ALTER TABLE IF EXISTS "withus_failed_sync_email_recipient" ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS "withus_failed_sync_email_recipient" FORCE ROW LEVEL security;