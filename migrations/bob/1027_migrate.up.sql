CREATE TABLE IF NOT EXISTS public.books_chapters (
    book_id text NOT NULL,
    chapter_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT books_chapters_pk PRIMARY KEY (book_id, chapter_id),
    CONSTRAINT books_chapters_un UNIQUE (book_id, chapter_id)
);



-- --migration data book_id of chapters
DO $$DECLARE 
	dChapter_id text;
	dBook_id text;
BEGIN
    FOR dBook_id, dChapter_id IN 
    	SELECT c1.book_id, c1.chapter_id
    	FROM chapters c1 
    	JOIN books cb ON cb.book_id = c1.book_id 
    loop
		INSERT INTO books_chapters (chapter_id , book_id, deleted_at, created_at , updated_at )
		SELECT c.chapter_id , dBook_id as book_id, c.deleted_at , c.created_at , c.updated_at 
		FROM chapters c WHERE c.chapter_id  = dChapter_id 
		ON CONFLICT DO NOTHING;	
    END LOOP;
END$$;

ALTER TABLE chapters DROP COLUMN IF EXISTS book_id;