ALTER TABLE IF EXISTS public.chapters
    ADD COLUMN IF NOT EXISTS book_id text NULL;


CREATE OR REPLACE FUNCTION update_book_id_for_chapters_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN 
-- IF new.book_id != old.book_id THEN
    UPDATE public.chapters 
    SET book_id = new.book_id
    WHERE chapter_id = new.chapter_id;
    -- END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_book_id_for_chapters ON public.books_chapters;

CREATE TRIGGER update_book_id_for_chapters AFTER INSERT OR UPDATE ON books_chapters FOR EACH ROW EXECUTE FUNCTION update_book_id_for_chapters_fn();

UPDATE public.chapters c
SET book_id = bc.book_id
FROM public.books_chapters bc
WHERE c.chapter_id = bc.chapter_id;