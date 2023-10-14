INSERT INTO public.configs
(config_key, config_group, country, config_value, updated_at, created_at, deleted_at)
VALUES
('limitDuration', 'ask_tutor', 'COUNTRY_VN', 'THIS_WEEK', NOW(), NOW(), NULL), -- value can be THIS_DAY, THIS_WEEK, THIS_MONTH, THIS_YEAR
('limitTotalQuestion', 'ask_tutor', 'COUNTRY_VN', '5', NOW(), NOW(), NULL)
ON CONFLICT DO NOTHING;

