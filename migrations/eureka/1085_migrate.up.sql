CREATE INDEX IF NOT EXISTS books_chapters_book_id_idx ON public.books_chapters ("book_id");

CREATE INDEX IF NOT EXISTS learning_material_topic_id_idx ON public.learning_material ("topic_id");

CREATE OR REPLACE FUNCTION book_tree_fn()
RETURNS TABLE (
  book_id text,
  chapter_id text,
  chapter_display_order smallint,
  topic_id text,
  topic_display_order smallint,
  learning_material_id text,
  lm_display_order smallint
  )
LANGUAGE SQL
SECURITY INVOKER
AS $btf$
SELECT
  b.book_id,
  c.chapter_id,
  c.display_order AS chapter_display_order,
  t.topic_id,
  t.display_order AS topic_display_order,
  lm.learning_material_id,
  lm.display_order AS lm_display_order
FROM
  books b
  JOIN books_chapters bc USING (book_id)
  JOIN chapters c USING (chapter_id)
  JOIN topics t USING (chapter_id)
  JOIN learning_material lm USING (topic_id)
WHERE
  COALESCE(
    b.deleted_at,
    bc.deleted_at,
    c.deleted_at,
    t.deleted_at,
    lm.deleted_at
  ) IS NULL;
$btf$;

CREATE OR REPLACE VIEW book_tree AS SELECT * FROM book_tree_fn();