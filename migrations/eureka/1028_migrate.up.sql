CREATE OR REPLACE FUNCTION update_content_structure_flatten_on_assignment_study_plan_item_created_fn()
RETURNS TRIGGER
AS $$
  DECLARE
    assignment_id text := NEW.assignment_id;
    content_structure jsonb := (SELECT content_structure FROM study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id);
    course_id text := content_structure->>'course_id';
    book_id text := content_structure->>'book_id';
    chapter_id text := content_structure->>'chapter_id';
    topic_id text := content_structure->>'topic_id';
  BEGIN
    UPDATE study_plan_items
    SET content_structure_flatten = 'book::' || book_id || 'topic::' || topic_id || 'chapter::' || chapter_id || 'course::' || course_id || 'assignment::' || assignment_id
    WHERE study_plan_item_id = NEW.study_plan_item_id;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_content_structure_flatten ON public.assignment_study_plan_items;
CREATE TRIGGER update_content_structure_flatten
  AFTER INSERT
  ON assignment_study_plan_items
  FOR EACH ROW
  EXECUTE FUNCTION update_content_structure_flatten_on_assignment_study_plan_item_created_fn();
