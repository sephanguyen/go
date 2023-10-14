package repo

const DistinctKeyword string = "distinct"

const getLessonQueryAscending = `
	select fl.lesson_id, fl."name", fl.start_time, fl.end_time, fl.teaching_method,
	fl.teaching_medium, fl.center_id, fl.course_id, fl.class_id, fl.scheduling_status, fl.lesson_capacity,
	fl.end_at, fl.zoom_link
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) > ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time ASC, fl.end_time ASC, fl.lesson_id ASC
	LIMIT $%d
	`

const getLessonQueryDescending = `
	select fl.lesson_id, fl."name", fl.start_time, fl.end_time, fl.teaching_method,
	fl.teaching_medium, fl.center_id, fl.course_id, fl.class_id, fl.scheduling_status, fl.lesson_capacity,
	fl.end_at, fl.zoom_link
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) < ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time DESC, fl.end_time DESC, fl.lesson_id DESC
	LIMIT $%d
	`

const previousLessonQueryAscending = `
	, previous_sort as (select fl.lesson_id, count(*) OVER() AS total, fl.start_time, fl.end_time
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) < ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time DESC, fl.end_time DESC, fl.lesson_id DESC
	LIMIT $%d) select ps.lesson_id, ps.total
		FROM previous_sort ps
		order by ps.start_time ASC, ps.end_time ASC, ps.lesson_id ASC
		LIMIT 1
	`

const previousLessonQueryDescending = `
	, previous_sort as (select fl.lesson_id, count(*) OVER() AS total, fl.start_time, fl.end_time
	from filter_lesson fl
	where $%d::text IS NULL
				OR (fl.start_time, fl.end_time, fl.lesson_id) > ((SELECT start_time FROM lessons WHERE lesson_id = $%d LIMIT 1), (SELECT end_time FROM lessons WHERE lesson_id = $%d LIMIT 1), $%d)
	order by fl.start_time ASC, fl.end_time ASC, fl.lesson_id ASC
	LIMIT $%d) select ps.lesson_id, ps.total
		FROM previous_sort ps
		order by ps.start_time DESC, ps.end_time DESC, ps.lesson_id DESC
		LIMIT 1
	`
