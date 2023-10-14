ALTER TABLE public.school_configs ALTER COLUMN plan_expired_at TYPE timestamptz USING plan_expired_at::timestamptz;

ALTER TABLE public.classes ALTER COLUMN plan_expired_at TYPE timestamptz USING plan_expired_at::timestamptz;

UPDATE public.configs
SET config_value='3000-06-30 23:59:59', updated_at=now() 
WHERE config_key='planPeriod' AND config_group='class_plan' AND country='COUNTRY_VN';
