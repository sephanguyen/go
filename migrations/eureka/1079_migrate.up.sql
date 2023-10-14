CREATE TABLE IF NOT EXISTS public.learning_objective (
    video text,
    study_guide text,
    video_script text,
    CONSTRAINT learning_objective_pk PRIMARY KEY (learning_material_id),
    CONSTRAINT topic_id_fk FOREIGN KEY (topic_id) REFERENCES public.topics(topic_id)
) INHERITS (public.learning_material);
/* set RLS */
CREATE POLICY rls_learning_objective ON "learning_objective" using (
    permission_check(resource_path, 'learning_objective')
) with check (
    permission_check(resource_path, 'learning_objective')
);

ALTER TABLE
    "learning_objective" ENABLE ROW LEVEL security;
ALTER TABLE
    "learning_objective" FORCE ROW LEVEL security;

CREATE OR REPLACE FUNCTION create_learning_objective_fn() 
RETURNS TRIGGER 
AS $$ 
BEGIN IF new.type = 'LEARNING_OBJECTIVE_TYPE_LEARNING' THEN
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
    VALUES
        (
            new.lo_id,
            new.topic_id,
            new.name,
            'LEARNING_MATERIAL_LEARNING_OBJECTIVE',
            new.display_order,
            new.created_at,
            new.updated_at,
            new.resource_path,
            new.video,
            new.study_guide,
            new.video_script,
            new.deleted_at
        ) ON CONFLICT 
    ON CONSTRAINT learning_objective_pk
    DO UPDATE
    SET
        updated_at = new.updated_at,
        name = new.name,
        display_order = new.display_order,
        video = new.video,
        study_guide = new.study_guide,
        video_script = new.video_script,
        deleted_at = new.deleted_at;
        END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS create_learning_objective ON public.learning_objectives;

CREATE TRIGGER create_learning_objective AFTER INSERT OR UPDATE ON learning_objectives FOR EACH ROW EXECUTE FUNCTION create_learning_objective_fn();
