TRUNCATE TABLE public.withus_mapping_exam_lo_id;

ALTER TABLE public.withus_mapping_exam_lo_id 
ADD CONSTRAINT exam_lo_id_fk FOREIGN KEY (exam_lo_id) REFERENCES public.exam_lo(learning_material_id);

INSERT INTO public.withus_mapping_exam_lo_id (
    exam_lo_id,
    resource_path
)
SELECT 
    learning_material_id,
    resource_path
FROM public.exam_lo
ON CONFLICT ON CONSTRAINT withus_mapping_exam_lo_id_pk DO NOTHING;

TRUNCATE TABLE public.withus_mapping_question_tag;

ALTER TABLE public.withus_mapping_question_tag
ADD CONSTRAINT question_tag_id_fk FOREIGN KEY (manabie_tag_id) REFERENCES public.question_tag(question_tag_id);

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

CREATE OR REPLACE FUNCTION public.withus_check_valid_course_id() RETURNS TRIGGER
LANGUAGE plpgsql
AS $BODY$
BEGIN
    IF NEW.manabie_course_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM public.course_students WHERE course_id = NEW.manabie_course_id) THEN
            RAISE EXCEPTION 'manabie_course_id % does not exist', NEW.manabie_course_id;
        END IF;
    END IF;
    RETURN NEW;
END;
$BODY$;

DROP TRIGGER IF EXISTS withus_check_valid_course_id ON public.withus_mapping_course_id;
CREATE TRIGGER withus_check_valid_course_id
BEFORE INSERT OR UPDATE ON public.withus_mapping_course_id
FOR EACH ROW
EXECUTE FUNCTION public.withus_check_valid_course_id();

TRUNCATE TABLE public.withus_mapping_course_id;

INSERT INTO public.withus_mapping_course_id (
    manabie_course_id,
    resource_path
)
SELECT 
    course_id,
    resource_path
FROM public.course_students
ON CONFLICT ON CONSTRAINT withus_mapping_course_id_pk DO NOTHING;
