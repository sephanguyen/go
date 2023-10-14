CREATE OR REPLACE FUNCTION public.get_list_course_student_study_plans_by_filter_v2(_course_id text, search text, _book_ids text[], _status text, _grades integer[])
 RETURNS SETOF course_students
 LANGUAGE sql
 STABLE
AS $function$

SELECT DISTINCT c_student.* FROM public.course_students as c_student
LEFT JOIN student_study_plans as s_study_plans
    USING(student_id)
LEFT JOIN study_plans as st
    USING(course_id, study_plan_id)
WHERE c_student.course_id = _course_id
AND c_student.deleted_at IS NULL
AND (
    -- When user don't apply filter should return matched
    (_book_ids = '{}' AND _grades = '{}' AND search = '')
    OR (
        st.deleted_at IS NULL
        AND st.name ILIKE ('%' || search || '%')
        AND st.status = _status
        AND (
            _book_ids = '{}' OR
            st.book_id = ANY(_book_ids)
        )
        AND (
            _grades = '{}' OR
        	EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades))
        )
    )
)

$function$;