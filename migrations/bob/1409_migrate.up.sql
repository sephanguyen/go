ALTER TABLE IF EXISTS user_access_paths 
	DROP CONSTRAINT IF EXISTS user_access_paths_users_fk,
	ADD CONSTRAINT user_access_paths_users_fk FOREIGN KEY (user_id) REFERENCES public.users(user_id) DEFERRABLE INITIALLY DEFERRED;
