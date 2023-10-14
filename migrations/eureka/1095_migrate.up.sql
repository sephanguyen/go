-- fix case properties of settings is empty
CREATE OR REPLACE FUNCTION create_task_assignment_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN IF new.type = 'ASSIGNMENT_TYPE_TASK' THEN
    INSERT INTO
        task_assignment (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            attachments,
            instruction,
            require_duration,
            require_complete_date,
            require_understanding_level,
            require_correctness,
            require_attachment,
            require_assignment_note,
            created_at,
            updated_at,
            deleted_at,
            resource_path
        )
    VALUES
        (
            new.assignment_id,
            new.content ->> 'topic_id',
            new.name,
            'LEARNING_MATERIAL_TASK_ASSIGNMENT',
            new.display_order,
            new.attachment,
            new.instruction,
            (COALESCE(new.settings ->> 'require_duration', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_complete_date', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_understanding_level', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_correctness', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
            (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
            new.created_at,
            new.updated_at,
            new.deleted_at,
            new.resource_path
        ) ON CONFLICT 
    ON CONSTRAINT task_assignment_pk
    DO UPDATE
    SET
        name = new.name,
        display_order = new.display_order,
        attachments = new.attachment,
        instruction = new.instruction,
        require_duration = (COALESCE(new.settings ->> 'require_duration', 'false'))::boolean,
        require_complete_date = (COALESCE(new.settings ->> 'require_complete_date', 'false'))::boolean,
        require_understanding_level = (COALESCE(new.settings ->> 'require_understanding_level', 'false'))::boolean,
        require_correctness = (COALESCE(new.settings ->> 'require_correctness', 'false'))::boolean,
        require_attachment = (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        require_assignment_note = (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        updated_at = new.updated_at,
        deleted_at = new.deleted_at;
    END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- migrate data from assignments to task_assignment
INSERT INTO
    task_assignment (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        attachments,
        instruction,
        require_duration,
        require_complete_date,
        require_understanding_level,
        require_correctness,
        require_attachment,
        require_assignment_note,
        created_at,
        updated_at,
        deleted_at,
        resource_path
    )
SELECT
    assignment_id,
    content ->> 'topic_id',
    name,
    'LEARNING_MATERIAL_TASK_ASSIGNMENT',
    display_order,
    attachment,
    instruction,
    (COALESCE(settings ->> 'require_duration', 'false'))::boolean,
    (COALESCE(settings ->> 'require_complete_date', 'false'))::boolean,
    (COALESCE(settings ->> 'require_understanding_level', 'false'))::boolean,
    (COALESCE(settings ->> 'require_correctness', 'false'))::boolean,
    (COALESCE(settings ->> 'require_attachment', 'false'))::boolean,
    (COALESCE(settings ->> 'require_assignment_note', 'false'))::boolean,
    created_at,
    updated_at,
    deleted_at,
    resource_path
FROM
    assignments
WHERE
type = 'ASSIGNMENT_TYPE_TASK' 
AND content ->> 'topic_id' <> ''
ON CONFLICT
ON CONSTRAINT task_assignment_pk
DO NOTHING;