-- migrate missing topics_learning_objectives
DO $$
DECLARE
    is_valid bool;
BEGIN
    INSERT INTO public.topics_learning_objectives(topic_id, lo_id, display_order, updated_at, created_at, deleted_at, resource_path)
    SELECT lo.topic_id, lo.lo_id, lo.display_order, lo.updated_at, lo.created_at, lo.deleted_at, lo.resource_path
    FROM public.learning_objectives lo
    WHERE lo_id not in (select lo_id from public.topics_learning_objectives);
    select (count(1)::int = 0::int) INTO is_valid from learning_objectives where lo_id not in(select lo_id from topics_learning_objectives);
    IF is_valid is false THEN
        ROLLBACK;
        raise 'wrong command sync `topics_learning_objectives` data';
    ELSE
        COMMIT;
    END IF;
END$$;
