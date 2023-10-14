CREATE TABLE IF NOT EXISTS public.offline_learning (
    CONSTRAINT offline_learning_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT offline_learning_topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id),
    CONSTRAINT offline_learning_type_check CHECK (type = 'LEARNING_MATERIAL_OFFLINE_LEARNING')
) INHERITS (public.learning_material);
/* set RLS */
CREATE POLICY rls_offline_learning ON "offline_learning" using (
    permission_check(resource_path, 'offline_learning')
) with check (
    permission_check(resource_path, 'offline_learning')
);

ALTER TABLE
    "offline_learning" ENABLE ROW LEVEL security;
ALTER TABLE
    "offline_learning" FORCE ROW LEVEL security;

CREATE OR REPLACE FUNCTION create_offline_learning_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN IF new.type = 'LEARNING_OBJECTIVE_TYPE_OFFLINE_LEARNING' THEN
    INSERT INTO
        offline_learning (
            learning_material_id,
            topic_id,
            name,
            type,
            display_order,
            created_at,
            updated_at,
            resource_path,
            deleted_at
        )
    VALUES
        (
            new.lo_id,
            new.topic_id,
            new.name,
            'LEARNING_MATERIAL_OFFLINE_LEARNING',
            new.display_order,
            new.created_at,
            new.updated_at,
            new.resource_path,
            new.deleted_at
        ) ON CONFLICT 
    ON CONSTRAINT offline_learning_pk
    DO UPDATE
    SET
        updated_at = new.updated_at,
        name = new.name,
        display_order = new.display_order,
        deleted_at = new.deleted_at;
        END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS create_offline_learning ON public.learning_objectives;

CREATE TRIGGER create_offline_learning AFTER INSERT OR UPDATE ON learning_objectives FOR EACH ROW EXECUTE FUNCTION create_offline_learning_fn();


INSERT INTO
    offline_learning (
        learning_material_id,
        topic_id,
        name,
        type,
        display_order,
        created_at,
        updated_at,
        resource_path,
        deleted_at
    )
SELECT
    lo_id,
    topic_id,
    name,
    'LEARNING_MATERIAL_OFFLINE_LEARNING',
    display_order,
    created_at,
    updated_at,
    resource_path,
    deleted_at
FROM
    learning_objectives
WHERE
type = 'LEARNING_OBJECTIVE_TYPE_OFFLINE_LEARNING'
ON CONFLICT 
ON CONSTRAINT offline_learning_pk
DO NOTHING; 