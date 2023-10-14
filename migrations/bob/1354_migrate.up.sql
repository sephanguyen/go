CREATE TABLE IF NOT EXISTS public.reallocation (
    student_id TEXT NOT NULL,
    original_lesson_id TEXT NOT NULL,
    new_lesson_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT reallocation__pk PRIMARY KEY (student_id,original_lesson_id),
    CONSTRAINT reallocation__original_lesson_id__fk FOREIGN KEY (original_lesson_id) REFERENCES public.lessons(lesson_id),
    CONSTRAINT reallocation__new_lesson_id__fk FOREIGN KEY (new_lesson_id) REFERENCES public.lessons(lesson_id)
);


CREATE POLICY rls_reallocation ON "reallocation"
USING (permission_check(resource_path, 'reallocation')) WITH CHECK (permission_check(resource_path, 'reallocation'));
CREATE POLICY rls_reallocation_restrictive ON "reallocation" AS RESTRICTIVE
USING (permission_check(resource_path, 'reallocation'))WITH CHECK (permission_check(resource_path, 'reallocation'));

ALTER TABLE "reallocation" ENABLE ROW LEVEL security;
ALTER TABLE "reallocation" FORCE ROW LEVEL security;
