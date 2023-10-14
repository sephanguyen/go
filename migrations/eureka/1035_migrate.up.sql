-- migrate missing topics_assignments
DO $$
DECLARE
    is_valid bool;
BEGIN
    INSERT INTO public.topics_assignments(topic_id, assignment_id, display_order, updated_at, created_at, deleted_at, resource_path)
    SELECT assignments.content->>'topic_id', assignments.assignment_id,assignments.display_order, assignments.updated_at, assignments.created_at, assignments.deleted_at, assignments.resource_path
    FROM public.assignments
    WHERE assignment_id not in (select assignment_id from public.topics_assignments);
    select (count(1)::int = 0::int) INTO is_valid from assignments where assignment_id not in(select assignment_id from topics_assignments);
    IF is_valid is false THEN
        ROLLBACK;
        raise 'wrong command sync `topics_assignments` data';
    ELSE
        COMMIT;
    END IF;
END$$;
