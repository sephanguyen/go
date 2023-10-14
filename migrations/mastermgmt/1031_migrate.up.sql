INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
VALUES
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483648'), --Manabie
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483647'), --JPREP
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483646'), --Synersia
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483645'), --Renseikai
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483644'), --Bestco
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483643'), --Bestco
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483642'), --KEC
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483641'), --AIC
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483640'), --NSG
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483636'), --withus
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483635'), --KEC
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'on', '-2147483631'), --Eishinkan
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483633'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483632'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483639'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483630'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '16091'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483629'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '16093'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483638'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483637'),
(uuid_generate_v4(), 'user.enrollment.update_status_manual', 'string', now(), now(), 'off', '-2147483634')
ON CONFLICT DO NOTHING;
