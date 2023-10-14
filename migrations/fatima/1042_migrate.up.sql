CREATE TABLE IF NOT EXISTS public.order_action_log (
    id integer NOT NULL,
    user_id text NOT NULL,
    order_id text NOT NULL,
    action text,
    comment text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),

    CONSTRAINT order_action_log_pk PRIMARY KEY (id),
    CONSTRAINT order_action_log_users_fk FOREIGN KEY (user_id) REFERENCES "users"(user_id),
    CONSTRAINT order_action_log_order_fk FOREIGN KEY (order_id) REFERENCES "order"(id)
);

CREATE SEQUENCE public.order_action_log_id_seq
    AS integer;

ALTER SEQUENCE public.order_action_log_id_seq OWNED BY public.order_action_log.id;

ALTER TABLE ONLY public.order_action_log ALTER COLUMN id SET DEFAULT nextval('public.order_action_log_id_seq'::regclass);

CREATE POLICY rls_order_action_log ON "order_action_log" using (permission_check(resource_path, 'order_action_log')) with check (permission_check(resource_path, 'order_action_log'));

ALTER TABLE "order_action_log" ENABLE ROW LEVEL security;
ALTER TABLE "order_action_log" FORCE ROW LEVEL security;