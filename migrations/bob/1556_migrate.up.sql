-- create index for deleted_at of student_enrollment_status_history table
DROP INDEX IF EXISTS student_enrollment_status_history__deleted_at_idx;
CREATE INDEX IF NOT EXISTS student_enrollment_status_history__deleted_at_idx ON public.student_enrollment_status_history (deleted_at);

-- create get_active_student_ids_with_grade_sequence function
DROP FUNCTION IF EXISTS get_active_student_ids_with_grade_sequence;

CREATE OR REPLACE FUNCTION public.get_active_student_ids_with_grade_sequence(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL
) RETURNS TABLE(student_id TEXT, grade_sequence INT)
    LANGUAGE SQL STABLE
    AS $$
        WITH active_student_ids AS (
            SELECT DISTINCT esh.student_id FROM student_enrollment_status_history AS esh
            WHERE 
                (enrollment_statuses IS NULL OR esh.enrollment_status = ANY(enrollment_statuses)) AND
                esh.start_date < now() AND
                (esh.end_date IS null OR esh.end_date > now()) AND
                (location_ids IS NULL OR esh.location_id = ANY(location_ids)) AND
                deleted_at IS NULL
        ),
        active_student_ids_with_grade_sequence AS (
            SELECT s2.student_id, g.sequence grade_sequence FROM active_student_ids AS s
            INNER JOIN students AS s2 ON s.student_id = s2.student_id
            LEFT JOIN grade AS g ON g.grade_id = s2.grade_id
            WHERE (grade_ids IS NULL OR g.grade_id = ANY(grade_ids))
        )
        SELECT student_id, grade_sequence FROM active_student_ids_with_grade_sequence
    $$;

-- update get_sorted_students_list function
DROP FUNCTION IF EXISTS get_sorted_students_list;

CREATE OR REPLACE FUNCTION public.get_sorted_students_list(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL
) RETURNS SETOF public.users
    LANGUAGE SQL STABLE
    AS $$
        WITH active_student_ids_with_grade_sequence AS (
            SELECT student_id
            FROM 
                get_active_student_ids_with_grade_sequence(location_ids, enrollment_statuses, grade_ids)
        ),
        sorted_students AS (
            SELECT u.* FROM users AS u
            INNER JOIN active_student_ids_with_grade_sequence AS s ON s.student_id = u.user_id
            ORDER BY u.created_at DESC, u.user_id DESC
        )
        SELECT * FROM sorted_students
    $$;

-- update get_sorted_students_list_v2 function
DROP FUNCTION IF EXISTS get_sorted_students_list_v2;

CREATE OR REPLACE FUNCTION public.get_sorted_students_list_v2(
    location_ids TEXT[] DEFAULT NULL, 
    enrollment_statuses TEXT[] DEFAULT NULL, 
    grade_ids TEXT[] DEFAULT NULL,
    has_sort BOOLEAN DEFAULT true
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
            ORDER BY
                CASE WHEN has_sort THEN s.grade_sequence END ASC NULLS LAST, 
                CASE WHEN has_sort THEN u.full_name_phonetic END COLLATE japanese_collation ASC NULLS LAST
        )
        SELECT * FROM sorted_students
    $$;

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
            WHERE (search_text IS NULL OR nospace(u.name) ILIKE nospace(search_text) OR nospace(u.full_name_phonetic) ILIKE nospace(search_text))
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
