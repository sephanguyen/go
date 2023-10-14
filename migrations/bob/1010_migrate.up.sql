CREATE VIEW public.lesson_schedules AS
 SELECT to_char(p.start_date , 'YYYY-MM-DD') AS "formatted_date", to_char(p.start_date , 'YYYY') AS "formatted_year", to_char(p.start_date , 'MM') as "formatted_month", to_char(p.start_date , 'DD') as "formatted_day", p.start_date 
 FROM public.preset_study_plans_weekly p JOIN lessons l ON l.lesson_id = p.lesson_id ORDER BY formatted_date DESC;
 
 
CREATE VIEW public.preset_study_plans_weekly_format AS
 SELECT to_char(p.start_date , 'YYYY-MM-DD') AS "formatted_date", to_char(p.start_date , 'YYYY') AS "formatted_year", to_char(p.start_date , 'MM') as "formatted_month", to_char(p.start_date , 'DD') as "formatted_day", p.* 
 FROM public.preset_study_plans_weekly p;
 
 
 
