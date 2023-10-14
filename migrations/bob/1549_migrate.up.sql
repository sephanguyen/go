CREATE OR REPLACE FUNCTION nospace(value TEXT) 
    RETURNS TEXT AS
    $BODY$
        DECLARE
        BEGIN
            value = TRANSLATE(value,' ','');
            RETURN value;
        END
    $BODY$
LANGUAGE 'plpgsql' IMMUTABLE;

CREATE INDEX IF NOT EXISTS users__name_gin__idx ON public.users USING gin (nospace((name)::text) gin_trgm_ops);

CREATE INDEX IF NOT EXISTS users__full_name_phonetic_gin__idx ON public.users USING gin (nospace((full_name_phonetic)::text) gin_trgm_ops);

-- create function get student list with sort w/o space
DROP FUNCTION IF EXISTS get_sorted_students_list_v3;
DROP TYPE IF EXISTS sort_type_enum;

CREATE TYPE sort_type_enum AS ENUM ('none', 'lms', 'erp');

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
        WITH active_student_ids AS (
            SELECT DISTINCT esh.student_id FROM student_enrollment_status_history AS esh
            WHERE 
                (enrollment_statuses IS NULL OR esh.enrollment_status = ANY(enrollment_statuses)) AND
                esh.start_date < now() AND
                (esh.end_date IS null OR esh.end_date > now()) AND
                (location_ids IS NULL OR esh.location_id = ANY(location_ids))
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
    $$
