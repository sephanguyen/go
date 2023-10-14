INSERT INTO public.github_repo_state (org,repo,is_blocked) 
VALUES
	('manabie-com','school-portal-admin',false), 
    ('manabie-com','student-app',false),
    ('manabie-com','eibanam',false)
ON CONFLICT DO NOTHING;