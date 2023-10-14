DROP VIEW IF EXISTS student_study_plans_view;
DROP VIEW IF EXISTS master_study_plan_view;
DROP VIEW IF EXISTS study_plan_tree;

DROP FUNCTION IF EXISTS master_study_plan_fn;
DROP FUNCTION IF EXISTS study_plan_tree_fn;

CREATE OR REPLACE FUNCTION study_plan_tree_fn()
RETURNS TABLE (
     study_plan_id text,
     book_id text,
     chapter_id text,
     chapter_display_order smallint,
     topic_id text,
     topic_display_order smallint,
     learning_material_id text,
     lm_display_order smallint,
     resource_path text
)
LANGUAGE SQL
SECURITY INVOKER
AS $func$
SELECT sp.study_plan_id , bt.book_id , bt.chapter_id  , bt.chapter_display_order , bt.topic_id, bt.topic_display_order, bt.learning_material_id, bt.lm_display_order, sp.resource_path
FROM study_plans sp JOIN book_tree bt USING (book_id);
$func$;

CREATE OR REPLACE VIEW study_plan_tree AS
SELECT * FROM study_plan_tree_fn();

CREATE OR REPLACE FUNCTION master_study_plan_fn()
RETURNS TABLE (
    study_plan_id text,
    book_id text,
    chapter_id text,
    chapter_display_order smallint,
    topic_id text,
    topic_display_order smallint,
    learning_material_id text,
    lm_display_order smallint,
    resource_path text,
    start_date timestamptz ,
    end_date timestamptz ,
    available_from timestamptz ,
    available_to timestamptz ,
    school_date timestamptz ,
    updated_at timestamptz ,
    status text
)
LANGUAGE SQL
SECURITY INVOKER
AS $func$
SELECT sp.*, m.start_date, m.end_date, m.available_from, m.available_to, m.school_date, m.updated_at, m.status
FROM study_plan_tree sp
         LEFT JOIN master_study_plan m USING (study_plan_id, learning_material_id);
$func$;

CREATE OR REPLACE VIEW master_study_plan_view AS
SELECT * FROM master_study_plan_fn();

CREATE
OR REPLACE FUNCTION individual_study_plan_fn()
RETURNS TABLE (
    student_id text,
    study_plan_id text,
    book_id text,
    chapter_id text,
    chapter_display_order smallint,
    topic_id text,
    topic_display_order smallint,
    learning_material_id text,
    lm_display_order smallint,
    start_date timestamptz ,
    end_date timestamptz ,
    available_from timestamptz ,
    available_to timestamptz ,
    school_date timestamptz ,
    updated_at timestamptz ,
    status text,
    resource_path text
)
LANGUAGE SQL
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
FROM ((public.master_study_plan_view m
    JOIN public.student_study_plans ssp ON (((m.study_plan_id = ssp.master_study_plan_id) OR
                                             ((ssp.master_study_plan_id IS NULL) AND
                                              (m.study_plan_id = ssp.study_plan_id)))))
    LEFT JOIN public.individual_study_plan isp
      ON (((ssp.student_id = isp.student_id) AND (m.learning_material_id = isp.learning_material_id) AND
           (m.study_plan_id = isp.study_plan_id))));
$$;

CREATE
OR REPLACE VIEW public.individual_study_plans_view AS
SELECT *
FROM individual_study_plan_fn();
