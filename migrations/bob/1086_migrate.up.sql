WITH tmp AS (SELECT book_id, count(1) AS total_chapter FROM books_chapters group by book_id)
UPDATE books SET current_chapter_display_order = total_chapter, updated_at=now() FROM tmp WHERE books.book_id = tmp.book_id;
