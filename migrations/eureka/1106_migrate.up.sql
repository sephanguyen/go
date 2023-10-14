DROP VIEW IF EXISTS master_study_plan_view;
DROP VIEW IF EXISTS study_plan_tree;

CREATE OR REPLACE FUNCTION study_plan_tree_fn()
RETURNS TABLE (
     study_plan_id text,
     book_id text,
     chapter_id text,
     chapter_display_order smallint,
     topic_id text,
     topic_display_order smallint,
     learning_material_id text,
     lm_display_order smallint
)
LANGUAGE SQL
SECURITY INVOKER
AS $func$
SELECT sp.study_plan_id , bt.book_id , bt.chapter_id  , bt.chapter_display_order , bt.topic_id, bt.topic_display_order, bt.learning_material_id, bt.lm_display_order
FROM study_plans sp JOIN book_tree bt USING (book_id);
$func$;

CREATE OR REPLACE VIEW study_plan_tree AS
SELECT * FROM study_plan_tree_fn();
--
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
