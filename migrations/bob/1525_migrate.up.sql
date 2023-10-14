-- create index for enrollment status history table
DROP INDEX IF EXISTS student_enrollment_status_history_student_id_idx;
DROP INDEX IF EXISTS student_enrollment_status_history_location_id_idx;
CREATE INDEX IF NOT EXISTS student_enrollment_status_history__student_id_idx ON public.student_enrollment_status_history (student_id);
CREATE INDEX IF NOT EXISTS student_enrollment_status_history__location_id_idx ON public.student_enrollment_status_history (location_id);

-- create index (created_at desc, user_id desc) in users table
DROP INDEX IF EXISTS users__created_at_desc__user_id_desc__idx;
CREATE INDEX IF NOT EXISTS users__created_at_desc__user_id_desc__idx ON public.users (created_at desc, user_id desc);

-- create function get sorted students list by sequece grade and full_name_phonetic
DROP FUNCTION IF EXISTS get_sorted_students_list;

CREATE OR REPLACE FUNCTION public.get_sorted_students_list(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL
) RETURNS SETOF public.users
    LANGUAGE SQL STABLE
    AS $$
        WITH active_student_ids AS (
            SELECT DISTINCT(esh.student_id) FROM student_enrollment_status_history AS esh
            WHERE 
                (enrollment_statuses IS NULL OR esh.enrollment_status = ANY(enrollment_statuses)) AND
                esh.start_date < now() AND
                (esh.end_date IS null OR esh.end_date > now()) AND
                (location_ids IS NULL OR esh.location_id = ANY(location_ids))
            GROUP BY esh.student_id
        ),
        active_student_ids_grade_sequence AS (
            SELECT s2.*, g.sequence grade_sequence FROM active_student_ids AS s
            INNER JOIN students AS s2 ON s.student_id = s2.student_id
            INNER JOIN grade AS g ON g.grade_id = s2.grade_id
            WHERE (grade_ids IS NULL OR g.grade_id = ANY(grade_ids))
        ),
        sorted_students AS (
            SELECT u.* FROM users AS u
            INNER JOIN active_student_ids_grade_sequence AS s ON s.student_id = u.user_id
            ORDER BY u.created_at DESC, u.user_id DESC
        )
        SELECT * FROM sorted_students
    $$
