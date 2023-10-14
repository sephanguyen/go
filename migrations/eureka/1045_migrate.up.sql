WITH TMP AS (
	SELECT DISTINCT ON (sp.study_plan_id) 
    sp.study_plan_id,spi.content_structure ->> 'book_id'  AS book_id FROM public.study_plans AS sp
	JOIN public.study_plan_items AS spi
	USING(study_plan_id)
	WHERE sp.deleted_at IS NULL
	AND spi.deleted_at IS NULL
	ORDER BY sp.study_plan_id, spi.updated_at DESC
)
UPDATE public.study_plans AS sp
SET book_id = tmp.book_id
FROM TMP AS tmp
WHERE sp.study_plan_id = tmp.study_plan_id;
