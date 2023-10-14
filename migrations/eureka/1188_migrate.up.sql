-- modify table
ALTER TABLE IF EXISTS public.learning_objectives
    ADD COLUMN IF NOT EXISTS maximum_attempt INTEGER DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS approve_grading BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS grade_capping BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS review_option TEXT NOT NULL DEFAULT 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::TEXT,
    DROP CONSTRAINT IF EXISTS learning_objectives_review_option_check,
    ADD CONSTRAINT learning_objectives_review_option_check CHECK ((review_option = ANY(ARRAY['EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::TEXT, 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE'::TEXT])));

ALTER TABLE IF EXISTS public.exam_lo
    ADD COLUMN IF NOT EXISTS maximum_attempt INTEGER DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS approve_grading BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS grade_capping BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS review_option TEXT NOT NULL DEFAULT 'EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::TEXT,
    DROP CONSTRAINT IF EXISTS exam_lo_review_option_check,
    ADD CONSTRAINT exam_lo_review_option_check CHECK ((review_option = ANY(ARRAY['EXAM_LO_REVIEW_OPTION_IMMEDIATELY'::TEXT, 'EXAM_LO_REVIEW_OPTION_AFTER_DUE_DATE'::TEXT])));

--trigger function
CREATE OR REPLACE FUNCTION public.migrate_learning_objectives_to_exam_lo_fn() 
    RETURNS trigger
    LANGUAGE 'plpgsql'
AS $BODY$
BEGIN
    INSERT INTO exam_lo (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        deleted_at,
        resource_path,
        instruction,
        grade_to_pass,
        manual_grading,
        time_limit,
        maximum_attempt,
        approve_grading,
        grade_capping,
        review_option
    )
    VALUES (
        NEW.lo_id,
        NEW.topic_id,
        NEW.name,
        'LEARNING_MATERIAL_EXAM_LO',
        NEW.display_order,
        NEW.created_at,
        NEW.updated_at,
        NEW.deleted_at,
        NEW.resource_path,
        NEW.instruction,
        NEW.grade_to_pass,
        NEW.manual_grading,
        NEW.time_limit,
        NEW.maximum_attempt,
        NEW.approve_grading,
        NEW.grade_capping,
        NEW.review_option
    )
    ON CONFLICT ON CONSTRAINT exam_lo_pk DO UPDATE SET
        topic_id = NEW.topic_id,
        name = NEW.name,
        display_order = NEW.display_order,
        updated_at = NEW.updated_at,
        deleted_at = NEW.deleted_at,
        instruction = NEW.instruction,
        grade_to_pass = NEW.grade_to_pass,
        manual_grading = NEW.manual_grading,
        time_limit = NEW.time_limit,
        maximum_attempt = NEW.maximum_attempt,
        approve_grading = NEW.approve_grading,
        grade_capping = NEW.grade_capping,
        review_option = NEW.review_option;
    RETURN NULL;
END;
$BODY$;
