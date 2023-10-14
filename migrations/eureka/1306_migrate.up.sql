ALTER TABLE assessment_session
    DROP COLUMN IF EXISTS learning_material_id,
    DROP COLUMN IF EXISTS study_plan_id;
