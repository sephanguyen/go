WITH deleted_chapters AS (
    SELECT chapter_id 
    FROM public.chapters
    JOIN (
        SELECT DISTINCT(chapter_id)
        FROM public.books_chapters
        WHERE deleted_at IS NOT NULL
        GROUP BY chapter_id
    ) AS dc
    USING(chapter_id)
    WHERE chapters.deleted_at IS NULL
)

UPDATE public.chapters
SET deleted_at = NOW() 
WHERE chapter_id IN (SELECT * FROM deleted_chapters);