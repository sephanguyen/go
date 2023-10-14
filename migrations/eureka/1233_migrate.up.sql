CREATE or replace FUNCTION public.individual_study_plan_fn() RETURNS TABLE(student_id text, study_plan_id text, book_id text, chapter_id text, chapter_display_order smallint, topic_id text, topic_display_order smallint, learning_material_id text, lm_display_order smallint, start_date timestamp with time zone, end_date timestamp with time zone, available_from timestamp with time zone, available_to timestamp with time zone, school_date timestamp with time zone, updated_at timestamp with time zone, status text, resource_path text)
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
    JOIN public.student_study_plans_fn() ssp ON m.study_plan_id = ssp.study_plan_id
    LEFT JOIN public.individual_study_plan isp
      ON (((ssp.student_id = isp.student_id) AND (m.learning_material_id = isp.learning_material_id) AND
           (m.study_plan_id = isp.study_plan_id))));
$$;


