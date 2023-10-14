CREATE OR REPLACE FUNCTION public.get_list_course_study_plan_by_filter(_course_id text, search text, _book_ids text[], _status text[], _grades integer[])
 RETURNS SETOF course_study_plans
 LANGUAGE sql
 STABLE
AS $function$
SELECT  cs.* FROM public.course_study_plans as cs  JOIN study_plans as st 
USING(study_plan_id,course_id)

WHERE (_status = '{}' OR st.status = ANY(_status))
AND (_book_ids = '{}' OR st.book_id = ANY(_book_ids))
AND cs.deleted_at IS NULL
AND st.deleted_at IS NULL
AND cs.course_id = _course_id
AND st.name ilike ('%' || search || '%')
AND (
	EXISTS (SELECT * FROM UNNEST(grades) WHERE unnest = ANY(_grades)) 
	OR _grades = '{}'
)

$function$;
