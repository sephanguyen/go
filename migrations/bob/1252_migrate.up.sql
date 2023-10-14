CREATE TABLE IF NOT EXISTS user_phone_number (
     user_phone_number_id TEXT NOT NULL,
     user_id TEXT NOT NULL,
     phone_number TEXT NOT NULL,
     type TEXT NOT NULL,
     updated_at timestamp with time zone NOT NULL,
     created_at timestamp with time zone NOT NULL,
     deleted_at timestamp with time zone,
     resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT user_phone_number__pk PRIMARY KEY (user_phone_number_id),
    CONSTRAINT user_phone_number_user_id__fk FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);

CREATE POLICY rls_user_phone_number ON "user_phone_number"
USING (permission_check(resource_path, 'user_phone_number'))
WITH CHECK (permission_check(resource_path, 'user_phone_number'));

ALTER TABLE IF EXISTS user_phone_number ENABLE ROW LEVEL security;
ALTER TABLE IF EXISTS user_phone_number FORCE ROW LEVEL security;

CREATE INDEX user_phone_number_phone_number_idx ON user_phone_number USING btree(phone_number);

ALTER TABLE IF EXISTS students ADD contact_preference TEXT;