CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx__activity_logs__request_at__action_type" ON public.activity_logs USING btree (request_at, action_type); 
