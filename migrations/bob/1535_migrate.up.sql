-- create japanese_collation 
CREATE COLLATION IF NOT EXISTS japanese_collation (provider = icu, locale = 'en-u-kn-true-kr-digit-en-ja_JP');

-- create index for grade_id in students
DROP INDEX IF EXISTS students__grade_id__idx;
CREATE INDEX IF NOT EXISTS students__grade_id__idx ON public.students (grade_id);

-- create function get sorted students list by sequece grade and full_name_phonetic
DROP FUNCTION IF EXISTS get_sorted_students_list_v2;

CREATE OR REPLACE FUNCTION public.get_sorted_students_list_v2(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL,
    has_sort BOOLEAN DEFAULT true
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
            LEFT JOIN grade AS g ON g.grade_id = s2.grade_id
            WHERE (grade_ids IS NULL OR g.grade_id = ANY(grade_ids))
        ),
        sorted_students AS (
            SELECT u.* FROM users AS u
            INNER JOIN active_student_ids_grade_sequence AS s ON s.student_id = u.user_id
            ORDER BY
                CASE WHEN has_sort THEN s.grade_sequence END ASC NULLS LAST, 
                CASE WHEN has_sort THEN u.full_name_phonetic END COLLATE japanese_collation ASC NULLS LAST
        )
        SELECT * FROM sorted_students
    $$
