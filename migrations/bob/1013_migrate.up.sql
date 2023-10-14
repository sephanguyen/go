CREATE TABLE IF NOT EXISTS public.apple_users (
    apple_user_id text NOT NULL,
    user_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__apple_users PRIMARY KEY (apple_user_id),
    CONSTRAINT fk__apple_users__users FOREIGN KEY (user_id) REFERENCES public.users (user_id)
);
