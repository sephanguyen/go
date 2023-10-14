-- migrate data for chapter
-- because the previous version missing deleted_at
WITH tmp AS (SELECT book_id, count(1) AS total_chapter FROM books_chapters WHERE deleted_at IS NULL group by book_id)
UPDATE books SET current_chapter_display_order = total_chapter, updated_at=now() FROM tmp WHERE books.book_id = tmp.book_id;
-- migrate for topic
ALTER TABLE IF EXISTS public.chapters ADD COLUMN IF NOT EXISTS current_topic_display_order INTEGER DEFAULT 0;

WITH tmp AS (SELECT chapter_id, count(1) AS total_topic FROM topics where deleted_at IS NULL group by chapter_id)
UPDATE chapters SET current_topic_display_order = tmp.total_topic , updated_at=now() FROM tmp WHERE chapters.chapter_id = tmp.chapter_id;
