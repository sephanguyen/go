ALTER TABLE assignment 
DROP COLUMN IF EXISTS status;

CREATE OR REPLACE FUNCTION create_assignment_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN IF new.type = 'ASSIGNMENT_TYPE_LEARNING_OBJECTIVE' THEN
    INSERT INTO
        assignment (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            attachments,
            max_grade,
            instruction,
            is_required_grade,
            allow_resubmission,
            require_attachment,
            allow_late_submission,
            require_assignment_note,
            require_video_submission,
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
            'LEARNING_MATERIAL_GENERAL_ASSIGNMENT',
            new.display_order,
            new.attachment,
            new.max_grade,
            new.instruction,
            new.is_required_grade,
            (COALESCE(new.settings ->> 'allow_resubmission', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'allow_late_submission', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        	(COALESCE(new.settings ->> 'require_video_submission', 'false'))::boolean,
            new.created_at,
            new.updated_at,
            new.deleted_at,
            new.resource_path
        ) ON CONFLICT 
    ON CONSTRAINT assignment_pk
    DO UPDATE
    SET
        name = new.name,
        display_order = new.display_order,
        attachments = new.attachment,
        max_grade = new.max_grade,
        instruction = new.instruction,
        is_required_grade = new.is_required_grade,
        allow_resubmission = (COALESCE(new.settings ->> 'allow_resubmission', 'false'))::boolean,
        require_attachment = (COALESCE(new.settings ->> 'require_attachment', 'false'))::boolean,
        allow_late_submission = (COALESCE(new.settings ->> 'allow_late_submission', 'false'))::boolean,
        require_assignment_note = (COALESCE(new.settings ->> 'require_assignment_note', 'false'))::boolean,
        require_video_submission = (COALESCE(new.settings ->> 'require_video_submission', 'false'))::boolean,
        updated_at = new.updated_at,
        deleted_at = new.deleted_at;
    END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;