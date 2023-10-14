-- drop fk for replicate tables
ALTER TABLE public.lesson_members
	DROP CONSTRAINT IF EXISTS fk__lesson_members__lesson_id;

ALTER TABLE public.lessons_teachers
	DROP CONSTRAINT IF EXISTS lessons_fk;

ALTER TABLE  public.location_types
	DROP CONSTRAINT IF EXISTS location_type_id_fk;