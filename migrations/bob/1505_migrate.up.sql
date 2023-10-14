DROP TABLE IF EXISTS public.subjects;

CREATE TABLE IF NOT EXISTS public.subject (
    subject_id VARCHAR(36) NOT NULL PRIMARY KEY,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath()
);


CREATE POLICY rls_subject ON "subject" using (
	permission_check(resource_path, 'subject')
) 
with check (
	permission_check(resource_path, 'subject')
);


CREATE POLICY rls_subject_restrictive ON "subject" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'subject')
) with check (
    permission_check(resource_path, 'subject')
);

ALTER TABLE "subject" ENABLE ROW LEVEL security;
ALTER TABLE "subject" FORCE ROW LEVEL security;


-- MANY TO MANY WITH COURSE
CREATE TABLE IF NOT EXISTS public.course_subject (
	course_id varchar(36) NULL,
	subject_id varchar(36) NULL,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL,
	deleted_at timestamp with time zone,
	resource_path text NULL DEFAULT autofillresourcepath(),
	--- Remember to keep the order when filtering in query
	PRIMARY KEY (course_id, subject_id)
);

CREATE POLICY rls_course_subject ON "course_subject" 
using (
	permission_check(resource_path, 'course_subject')
) 
with check (
	permission_check(resource_path, 'course_subject')
);

CREATE POLICY rls_course_subject_restrictive ON "course_subject"  AS RESTRICTIVE TO PUBLIC using 
(
	permission_check(resource_path, 'course_subject')
) 
with check (
	permission_check(resource_path, 'course_subject')
);

ALTER TABLE "course_subject" ENABLE ROW LEVEL security;
ALTER TABLE "course_subject" FORCE ROW LEVEL security;
