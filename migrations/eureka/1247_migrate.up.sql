drop function if exists public.student_study_plans_fn();

CREATE or Replace FUNCTION public.individual_study_plan_fn() RETURNS TABLE(student_id text, study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, start_date timestamp with time zone, end_date timestamp with time zone, available_from timestamp with time zone, available_to timestamp with time zone, school_date timestamp with time zone, updated_at timestamp with time zone, status text, resource_path text)
    LANGUAGE sql STABLE
    AS $$
SELECT ssp.student_id,
       m.study_plan_id,
       m.book_id,
       m.chapter_id,
       m.chapter_display_order,
       m.topic_id,
       m.topic_display_order,
       m.learning_material_id,
       m.lm_display_order,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.start_date, isp.start_date) AS start_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.end_date,
                                         isp.end_date)                                               AS end_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_from,
                                         isp.available_from)                                         AS available_from,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.available_to,
                                         isp.available_to)                                           AS available_to,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.school_date,
                                         isp.school_date)                                            AS school_date,
       public.check_study_plan_item_time(m.updated_at, isp.updated_at, m.updated_at,
                                         isp.updated_at)                                             AS updated_at,
       CASE
           WHEN ((m.updated_at IS NULL) OR (isp.updated_at IS NULL)) THEN COALESCE(m.status, isp.status)
           ELSE
               CASE
                   WHEN (m.updated_at >= isp.updated_at) THEN m.status
                   ELSE isp.status
                   END
           END                                                                                       AS status,
       m.resource_path
FROM (public.master_study_plan_view m
    join (select cs.course_id, cs.student_id, sp.study_plan_id, sp.study_plan_type, cs.deleted_at
            from study_plans sp
            join course_students cs on cs.course_id = sp.course_id
            where sp.master_study_plan_id is null
        ) ssp ON m.study_plan_id = ssp.study_plan_id
    left join student_study_plans ssp_a
       on ssp.student_id = ssp_a.student_id and ssp.study_plan_id = ssp_a.study_plan_id
    LEFT JOIN public.individual_study_plan isp
      ON (((ssp.student_id = isp.student_id) AND (m.learning_material_id = isp.learning_material_id) AND
           (m.study_plan_id = isp.study_plan_id))))
    where (ssp_a.student_id is not null OR ssp.study_plan_type = 'STUDY_PLAN_TYPE_COURSE')
$$;

CREATE or replace FUNCTION public.get_list_course_student_study_plans_by_filter_v2(_course_id text, search text, _book_ids text[], _status text, _grades integer[]) RETURNS SETOF public.course_students
    LANGUAGE sql STABLE
    AS $$

SELECT DISTINCT c_student.*
FROM public.course_students as c_student
         LEFT JOIN individual_study_plan_fn() as s_study_plans
                   USING (student_id)
         LEFT JOIN study_plans as st
                   USING (study_plan_id)

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

$$;

CREATE or replace FUNCTION public.get_student_study_plans_by_filter_v2() RETURNS TABLE(study_plan_id text, master_study_plan_id text, name text, study_plan_type text, school_id integer, created_at timestamp with time zone, updated_at timestamp with time zone, deleted_at timestamp with time zone, course_id text, resource_path text, book_id text, status text, track_school_progress boolean, grades integer[], student_id text)
    LANGUAGE sql STABLE
    AS $$

select st.*, s_study_plans.student_id
from public.individual_study_plan_fn() as s_study_plans
         join study_plans as st
              USING (study_plan_id)
$$;


