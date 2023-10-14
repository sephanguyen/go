CREATE TABLE IF NOT EXISTS public.flash_card (
    CONSTRAINT flash_card_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id),
    CONSTRAINT learning_objective_type_check CHECK (type = 'LEARNING_MATERIAL_FLASH_CARD')
) INHERITS (public.learning_material);
/* set RLS */
CREATE POLICY rls_flash_card ON "flash_card" using (
    permission_check(resource_path, 'flash_card')
) with check (
    permission_check(resource_path, 'flash_card')
);

ALTER TABLE
    "flash_card" ENABLE ROW LEVEL security;
ALTER TABLE
    "flash_card" FORCE ROW LEVEL security;

CREATE OR REPLACE FUNCTION create_flash_card_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN IF new.type = 'LEARNING_OBJECTIVE_TYPE_FLASH_CARD' THEN
    INSERT INTO
        flash_card (
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
            'LEARNING_MATERIAL_FLASH_CARD',
            new.display_order,
            new.created_at,
            new.updated_at,
            new.resource_path,
            new.deleted_at
        ) ON CONFLICT 
    ON CONSTRAINT flash_card_pk
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

DROP TRIGGER IF EXISTS create_flash_card ON public.learning_objectives;

CREATE TRIGGER create_flash_card AFTER INSERT OR UPDATE ON learning_objectives FOR EACH ROW EXECUTE FUNCTION create_flash_card_fn();


INSERT INTO
    flash_card (
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
    'LEARNING_MATERIAL_FLASH_CARD',
    display_order,
    created_at,
    updated_at,
    resource_path,
    deleted_at
FROM
    learning_objectives
WHERE
type = 'LEARNING_OBJECTIVE_TYPE_FLASH_CARD'
ON CONFLICT 
ON CONSTRAINT flash_card_pk
DO NOTHING;