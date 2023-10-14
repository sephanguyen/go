UPDATE public.learning_objective SET type = 'LEARNING_MATERIAL_LEARNING_OBJECTIVE';

ALTER TABLE ONLY public.learning_objective
    ADD CONSTRAINT learning_objective_type_check CHECK (type = 'LEARNING_MATERIAL_LEARNING_OBJECTIVE');

INSERT INTO
    learning_objective (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        resource_path,
        video,
        study_guide,
        video_script,
        deleted_at
    )
SELECT
    lo_id,
    topic_id,
    name,
    'LEARNING_MATERIAL_LEARNING_OBJECTIVE',
    display_order,
    created_at,
    updated_at,
    resource_path,
    video,
    study_guide,
    video_script,
    deleted_at
FROM
    learning_objectives
WHERE
type = 'LEARNING_OBJECTIVE_TYPE_LEARNING'
ON CONFLICT 
ON CONSTRAINT learning_objective_pk
DO NOTHING