DROP INDEX IF EXISTS student_packages__properties__can_do_quiz__idx;

CREATE INDEX IF NOT EXISTS student_packages__properties__can_do_quiz__idx ON public.student_packages USING gin ((properties->'can_do_quiz'));
