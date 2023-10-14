update courses set end_date = start_date + INTERVAL '1 day'
	where end_date is null 
	and start_date is not null
	and start_date >= '2026-01-01 00:00:00+00:00'
	and resource_path <> '-2147483647';

update courses set end_date = '2026-01-01 00:00:00+00:00' 
where end_date is null 
	and (start_date is null or start_date < '2026-01-01 00:00:00+00:00')
	and resource_path <> '-2147483647';