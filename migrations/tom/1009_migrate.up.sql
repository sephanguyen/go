ALTER TABLE public.online_users
  DROP CONSTRAINT online_users_pk;

ALTER TABLE public.online_users
  ADD COLUMN IF NOT EXISTS online_user_id TEXT;

ALTER TABLE ONLY public.online_users
    ADD CONSTRAINT online_users_pk PRIMARY KEY (online_user_id);

CREATE INDEX IF NOT EXISTS idx__online_user__user_id ON public.online_users (user_id);

