\connect calendar;

INSERT INTO public.day_type (day_type_id,resource_path,created_at,updated_at,deleted_at,display_name,is_archived) VALUES	
	 ('regular','-2147483642',now(),now(),NULL,NULL,false) ON CONFLICT DO NOTHING;	

INSERT INTO public.day_type (day_type_id,resource_path,created_at,updated_at,deleted_at,display_name,is_archived) VALUES	
	 ('seasonal','-2147483642',now(),now(),NULL,NULL,false) ON CONFLICT DO NOTHING;	

INSERT INTO public.day_type (day_type_id,resource_path,created_at,updated_at,deleted_at,display_name,is_archived) VALUES	
	 ('spare','-2147483642',now(),now(),NULL,NULL,false) ON CONFLICT DO NOTHING;	

INSERT INTO public.day_type (day_type_id,resource_path,created_at,updated_at,deleted_at,display_name,is_archived) VALUES	
	 ('closed','-2147483642',now(),now(),NULL,NULL,false) ON CONFLICT DO NOTHING;	

INSERT INTO public.day_info("date",location_id,day_type_id,opening_time,status,resource_path,created_at,updated_at,deleted_at,time_zone) VALUES	
	 ('2022-10-07','01FR4M51XJY9E77GSN4QZ1Q9N7','regular','8:00',NULL,'-2147483642',now(),now(),NULL,'Asia/Ho_Chi_Minh') ON CONFLICT DO NOTHING;	

INSERT INTO public.scheduler (scheduler_id,start_date,end_date,freq,created_at,updated_at,deleted_at,resource_path) 	
VALUES ('0122TAS9VKTBWTJVYSIWD09KA1','2023-03-06 00:15:00+07','2023-03-31 01:15:00+07','weekly', now(),	
now(),NULL,'-2147483642') ON CONFLICT DO NOTHING;
