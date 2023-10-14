ALTER TABLE learning_material ADD COLUMN IF NOT EXISTS vendor_type TEXT NOT NULL CONSTRAINT vendor_type_check CHECK(vendor_type IN('MANABIE', 'LEARNOSITY')) DEFAULT 'MANABIE';

ALTER TABLE learning_objectives ADD COLUMN IF NOT EXISTS vendor_type TEXT NOT NULL CONSTRAINT vendor_type_check CHECK(vendor_type IN('MANABIE', 'LEARNOSITY')) DEFAULT 'MANABIE';

ALTER TABLE learning_material ADD COLUMN IF NOT EXISTS vendor_reference_id TEXT NULL;

ALTER TABLE learning_objectives ADD COLUMN IF NOT EXISTS vendor_reference_id TEXT NULL;

CREATE OR REPLACE FUNCTION public.create_learning_objective_fn()
 RETURNS trigger
 LANGUAGE plpgsql
AS $function$ 
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
            vendor_type,
            vendor_reference_id,
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
            new.vendor_type,
            new.vendor_reference_id,
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
        vendor_type = new.vendor_type,
        vendor_reference_id = new.vendor_reference_id,
        deleted_at = new.deleted_at;
        END IF;
RETURN NULL;
END;
$function$;
