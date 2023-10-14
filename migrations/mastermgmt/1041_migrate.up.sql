INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'invoice.invoicemgmt.enable_invoice_manager', 
    'string', 
    'off', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);


-- Enable in e2e-tokyo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'invoice.invoicemgmt.enable_invoice_manager' and resource_path ='-2147483639';

-- Enable in kec-demo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'invoice.invoicemgmt.enable_invoice_manager' and resource_path ='-2147483635';
