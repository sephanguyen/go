INSERT INTO
    assignment (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        attachments,
        max_grade,
        status,
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
SELECT
    assignment_id,
    content ->> 'topic_id',
    name,
    'LEARNING_MATERIAL_GENERAL_ASSIGNMENT',
    display_order,
    attachment,
    max_grade,
    status,
    instruction,
    is_required_grade,
    (COALESCE(settings ->> 'allow_resubmission', 'false'))::boolean,
    (COALESCE(settings ->> 'require_attachment', 'false'))::boolean,
    (COALESCE(settings ->> 'allow_late_submission', 'false'))::boolean,
    (COALESCE(settings ->> 'require_assignment_note', 'false'))::boolean,
    (COALESCE(settings ->> 'require_video_submission', 'false'))::boolean,
    created_at,
    updated_at,
    deleted_at,
    resource_path
FROM
    assignments
WHERE
type = 'ASSIGNMENT_TYPE_LEARNING_OBJECTIVE'
AND content ->> 'topic_id' <> ''
ON CONFLICT
ON CONSTRAINT assignment_pk
DO NOTHING;