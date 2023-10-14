CREATE TABLE IF NOT EXISTS public.books (
    book_id text NOT NULL,
    name text NOT NULL,
    country text,
    subject text,
    grade smallint,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    school_id integer DEFAULT '-2147483648'::integer NOT NULL,
    CONSTRAINT books_pk PRIMARY KEY (book_id),
    CONSTRAINT books_school_id_fk FOREIGN KEY (school_id) REFERENCES public.schools(school_id)
);

CREATE TABLE IF NOT EXISTS public.courses_books (
    book_id text NOT NULL,
    course_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT courses_books_pk PRIMARY KEY (book_id, course_id),
    CONSTRAINT courses_books_un UNIQUE (book_id, course_id)
);

ALTER TABLE public.chapters DROP CONSTRAINT IF EXISTS chapters_book_id_fk;

ALTER TABLE public.chapters
  ADD COLUMN IF NOT EXISTS book_id TEXT,
  ADD CONSTRAINT chapters_book_id_fk FOREIGN KEY (book_id) REFERENCES public.books(book_id);


-- --migration data of books and courses_books
DO $$DECLARE 
	dCourse_id text;
	dBook_id text;
BEGIN
  -- check chapter_ids exist to run second migration
	IF EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'courses' AND COLUMN_NAME = 'chapter_ids') 
	THEN 
      FOR dBook_id, dCourse_id IN 
        SELECT md5(random()::text || clock_timestamp()::text)::uuid as book_id, c1.course_id FROM courses c1 
        WHERE c1.course_type  = 'COURSE_TYPE_CONTENT' AND c1.chapter_ids IS NOT NULL 
        AND c1.course_id NOT IN (SELECT cb.course_id FROM courses_books cb WHERE cb.course_id  = c1.course_id)
      LOOP
    
      INSERT INTO courses_books (course_id , book_id, deleted_at, created_at , updated_at )
      SELECT c.course_id , dBook_id as book_id, c.deleted_at , c.created_at , c.updated_at 
      FROM courses c WHERE c.course_id  = dCourse_id 
      ON CONFLICT DO NOTHING;	

      INSERT INTO books (book_id, name, country, subject, grade, school_id , deleted_at, created_at , updated_at )
      SELECT dBook_id as book_id, c."name" , c.country , c.subject , c.grade , c.school_id , c.deleted_at , c.created_at , c.updated_at 
      FROM courses c
      WHERE c.course_type  = 'COURSE_TYPE_CONTENT' AND c.course_id  = dCourse_id 
      ON CONFLICT DO NOTHING;
      END LOOP;
	END IF;
END$$;


-- --migration data book_id of chapters
DO $$DECLARE 
	dChapter_id text;
	dBook_id text;
BEGIN
  -- check chapter_ids exist to run second migration
	IF EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'courses' AND COLUMN_NAME = 'chapter_ids') 
	THEN 
    FOR dBook_id, dChapter_id IN 
    	SELECT cb.book_id, unnest(c1.chapter_ids) 
    	FROM courses c1 
    	JOIN courses_books cb ON cb.course_id = c1.course_id 
    loop
		UPDATE chapters SET book_id = dBook_id WHERE chapter_id = dChapter_id;	
    END LOOP;
	END IF;
END$$;

