INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    description,
    created_at,
    updated_at
)
VALUES(
    'syllabus.to_review.feedback_item_reference_id',
    'string',
    -- This is ID for PROD
    'MANA_b14efb94-3a5e-4b06-a9a7-9c1c830f3a02',
    'CONFIGURATION_TYPE_INTERNAL',
    'Store an item reference id used to render feedback form in the to review',
    NOW(),
    NOW()
);
