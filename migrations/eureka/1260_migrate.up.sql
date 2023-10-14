CREATE OR REPLACE FUNCTION public.update_allocate_marker_when_exam_lo_submission_was_created()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS $FUNCTION$
BEGIN
    UPDATE allocate_marker
    SET teacher_id = NULL
    WHERE study_plan_id = NEW.study_plan_id
    AND student_id = NEW.student_id 
    AND learning_material_id = NEW.learning_material_id;

    RETURN NEW;
END;
$FUNCTION$;

DROP TRIGGER IF EXISTS update_allocate_marker_once_created ON public.exam_lo_submission;
CREATE TRIGGER update_allocate_marker_once_created
    AFTER INSERT ON public.exam_lo_submission
    FOR EACH ROW
    EXECUTE FUNCTION public.update_allocate_marker_when_exam_lo_submission_was_created();