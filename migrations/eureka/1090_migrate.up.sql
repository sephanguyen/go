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
            (new.settings ->> 'require_duration')::boolean,
        	(new.settings ->> 'require_complete_date')::boolean,
        	(new.settings ->> 'require_understanding_level')::boolean,
        	(new.settings ->> 'require_correctness')::boolean,
        	(new.settings ->> 'require_attachment')::boolean,
            (new.settings ->> 'require_assignment_note')::boolean,
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
        require_duration = (new.settings ->> 'require_duration')::boolean,
        require_complete_date = (new.settings ->> 'require_complete_date')::boolean,
        require_understanding_level = (new.settings ->> 'require_understanding_level')::boolean,
        require_correctness = (new.settings ->> 'require_correctness')::boolean,
        require_attachment = (new.settings ->> 'require_attachment')::boolean,
        require_assignment_note = (new.settings ->> 'require_assignment_note')::boolean,
        updated_at = new.updated_at,
        deleted_at = new.deleted_at;
    END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS create_task_assignment ON public.assignments;

CREATE TRIGGER create_task_assignment AFTER INSERT OR UPDATE ON assignments FOR EACH ROW EXECUTE FUNCTION create_task_assignment_fn();