UPDATE public.configuration_key 
SET default_value = 'off', updated_at = now()
WHERE config_key = 'communication.notification.enable_delete_notification'
