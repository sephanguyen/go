DROP INDEX IF EXISTS users__email_gin__idx;
CREATE INDEX IF NOT EXISTS users__email_gin__idx ON public.users USING gin (email gin_trgm_ops);

-- update get_sorted_students_list_v3 function
DROP FUNCTION IF EXISTS get_sorted_students_list_v3;

CREATE OR REPLACE FUNCTION public.get_sorted_students_list_v3(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL,
    search_text TEXT DEFAULT NULL,
    student_ids_by_phone_number TEXT[] DEFAULT NULL,
    sort_type sort_type_enum DEFAULT 'none'
) RETURNS SETOF public.users
    LANGUAGE SQL STABLE
    AS $$
        WITH active_student_ids_with_grade_sequence AS (
            SELECT student_id, grade_sequence
            FROM 
                get_active_student_ids_with_grade_sequence(location_ids, enrollment_statuses, grade_ids)
        ),
        sorted_students AS (
            SELECT u.* FROM users AS u
            INNER JOIN active_student_ids_with_grade_sequence AS s ON s.student_id = u.user_id
            WHERE (search_text IS NULL OR nospace(u.name) ILIKE nospace(search_text) OR nospace(u.full_name_phonetic) ILIKE nospace(search_text) OR u.email ILIKE search_text)
                OR (u.user_id = ANY(COALESCE(student_ids_by_phone_number, '{}')))
            ORDER BY
                CASE
                    WHEN sort_type = 'erp' THEN s.grade_sequence
                END ASC NULLS LAST,
                CASE
                    WHEN sort_type = 'erp' THEN u.full_name_phonetic COLLATE japanese_collation
                END ASC NULLS LAST,
                CASE
                    WHEN sort_type = 'lms' THEN u.created_at
                END DESC,
                CASE
                    WHEN sort_type = 'lms' THEN u.user_id
                END DESC
        )

        SELECT * FROM sorted_students
    $$;
