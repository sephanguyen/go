INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'invoice.invoicemgmt.enable_auto_default_convenience_store', 
    'string', 
    'off', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);
