CREATE OR REPLACE FUNCTION public.book_tree_fn() RETURNS TABLE(book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint)
    LANGUAGE sql STABLE
    AS $$
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
  JOIN chapters c USING (book_id)
  JOIN topics t USING (chapter_id)
  JOIN learning_material lm USING (topic_id)
WHERE
  COALESCE(
    b.deleted_at,
    c.deleted_at,
    t.deleted_at,
    lm.deleted_at
  ) IS NULL;
$$;