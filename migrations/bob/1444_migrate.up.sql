DROP FUNCTION IF EXISTS public.find_parents_or_enrolled_students;

CREATE OR REPLACE FUNCTION find_parents_or_enrolled_students(keyword TEXT, user_ids TEXT[])
RETURNS SETOF users
LANGUAGE SQL
STABLE
AS $FUNCTION$
    SELECT DISTINCT u.*
    FROM users u
        JOIN user_group_member ugm ON u.user_id = ugm.user_id AND u.resource_path = ugm.resource_path
        JOIN user_group ug ON ug.user_group_id = ugm.user_group_id  AND u.resource_path = ug.resource_path
        JOIN granted_role gr ON gr.user_group_id = ug.user_group_id AND u.resource_path = gr.resource_path
        JOIN "role" r ON gr.role_id = r.role_id AND u.resource_path = r.resource_path
        LEFT JOIN student_enrollment_status_history sesh ON u.user_id = sesh.student_id AND u.resource_path = sesh.resource_path
    WHERE (
            r.role_name = 'Parent'
            OR (
                r.role_name = 'Student'
                AND sesh.enrollment_status = 'STUDENT_ENROLLMENT_STATUS_ENROLLED'
                AND sesh.start_date <= now()
                AND (sesh.end_date >= now() OR sesh.end_date IS NULL)
            )
        )
        AND (u."name" ILIKE concat('%', $1, '%') OR u.email ILIKE concat('%', $1, '%'))
        AND ($2::TEXT[] IS NULL OR u.user_id = ANY($2))
        AND u.deleted_at IS NULL
        AND ugm.deleted_at IS NULL
        AND ug.deleted_at IS NULL
        AND gr.deleted_at IS NULL
        AND r.deleted_at IS NULL
        AND sesh.deleted_at IS NULL
    ORDER BY u.created_at DESC;
$FUNCTION$
;
