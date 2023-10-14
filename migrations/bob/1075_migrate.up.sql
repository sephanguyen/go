WITH t AS (
	SELECT pspw.lesson_id, pspw.start_date, pspw.end_date, t."name" 
	FROM preset_study_plans_weekly pspw
		JOIN topics t ON t.topic_id=pspw.topic_id)
UPDATE lessons AS l SET "name" = t."name" , start_time = t.start_date, end_time = t.end_date
FROM t 
WHERE l."name" IS NULL 
	AND l.start_time IS NULL 
	AND l.end_time IS NULL 
	AND l.lesson_id = t.lesson_id