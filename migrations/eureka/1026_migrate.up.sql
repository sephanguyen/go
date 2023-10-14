ALTER TABLE IF EXISTS study_plan_items
  ADD COLUMN IF NOT EXISTS content_structure_flatten TEXT;

CREATE INDEX IF NOT EXISTS study_plan_content_structure_idx
  ON public.study_plan_items (study_plan_id, content_structure_flatten);

CREATE OR REPLACE FUNCTION update_content_structure_flatten_fn()
RETURNS TRIGGER
AS $$
  DECLARE
    lo_id text := NEW.lo_id;
    content_structure jsonb := (SELECT content_structure FROM study_plan_items WHERE study_plan_item_id = NEW.study_plan_item_id);
    course_id text := content_structure->>'course_id';
    book_id text := content_structure->>'book_id';
    chapter_id text := content_structure->>'chapter_id';
    topic_id text := content_structure->>'topic_id';
  BEGIN
    UPDATE study_plan_items
    SET content_structure_flatten = 'book::' || book_id || 'topic::' || topic_id || 'chapter::' || chapter_id || 'course::' || course_id || 'lo::' || lo_id
    WHERE study_plan_item_id = NEW.study_plan_item_id;
    RETURN NULL;
  END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_content_structure_flatten ON public.lo_study_plan_items;
CREATE TRIGGER update_content_structure_flatten
  AFTER INSERT
  ON lo_study_plan_items
  FOR EACH ROW
  EXECUTE FUNCTION update_content_structure_flatten_fn();

INSERT INTO study_plan_items(study_plan_item_id,study_plan_id,available_from,start_date,end_date,deleted_at,available_to,created_at,updated_at,
copy_study_plan_item_id,content_structure,completed_at,display_order,content_structure_flatten)
SELECT spi.study_plan_item_id,study_plan_id,available_from,start_date,end_date,spi.deleted_at,available_to,spi.created_at,spi.updated_at,
copy_study_plan_item_id,content_structure,completed_at,display_order,
format('book::%stopic::%schapter::%scourse::%slo::%s', content_structure->>'book_id', content_structure->>'topic_id', content_structure->>'chapter_id', content_structure->>'course_id', lo_id)
FROM study_plan_items spi
INNER JOIN lo_study_plan_items ON lo_study_plan_items.study_plan_item_id = spi.study_plan_item_id
ON CONFLICT ON CONSTRAINT study_plan_items_pk
DO UPDATE SET content_structure_flatten = EXCLUDED.content_structure_flatten;
