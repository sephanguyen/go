CREATE TABLE IF NOT EXISTS user_address(
    user_address_id     TEXT NOT NULL,
    user_id             TEXT NOT NULL,
    type                TEXT NOT NULL,
    postal_code         TEXT,
    prefecture          TEXT,
    city                TEXT,

    created_at          timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at          timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at          timestamp with time zone,
    resource_path       TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT user_address__pk PRIMARY KEY (user_address_id),
    CONSTRAINT user_address__user_id__fk FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);

CREATE POLICY rls_user_address ON "user_address"
USING (permission_check(resource_path, 'user_address'))
WITH CHECK (permission_check(resource_path, 'user_address'));

ALTER TABLE "user_address" ENABLE ROW LEVEL security;
ALTER TABLE "user_address" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS address_street(
    address_street_id       TEXT NOT NULL,
    user_address_id         TEXT NOT NULL,
    street_name             TEXT NOT NULL,
    street_level            integer NOT NULL,

    created_at              timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at              timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at              timestamp with time zone,
    resource_path           TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT address_street__pk PRIMARY KEY (address_street_id),
    CONSTRAINT address_street__user_address_id__fk FOREIGN KEY (user_address_id) REFERENCES public.user_address(user_address_id)
);

CREATE POLICY rls_address_street ON "address_street"
USING (permission_check(resource_path, 'address_street'))
WITH CHECK (permission_check(resource_path, 'address_street'));

ALTER TABLE "address_street" ENABLE ROW LEVEL security;
ALTER TABLE "address_street" FORCE ROW LEVEL security;


