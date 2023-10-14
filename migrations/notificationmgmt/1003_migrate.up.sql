CREATE TABLE IF NOT EXISTS important_events (
    important_event_id TEXT NOT NULL,
    reference_id TEXT NOT NULL,
    content TEXT DEFAULT NULL,
    url TEXT DEFAULT NULL,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    valid_to TIMESTAMP WITH TIME ZONE DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__important_events PRIMARY KEY (important_event_id),
    CONSTRAINT uk__important_events__reference_id UNIQUE (reference_id)
);

CREATE POLICY rls_important_events ON "important_events" USING (permission_check(resource_path, 'important_events')) with check (permission_check(resource_path, 'important_events'));
CREATE POLICY rls_important_events_restrictive ON "important_events" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'important_events')) WITH CHECK (permission_check(resource_path, 'important_events'));

ALTER TABLE "important_events" ENABLE ROW LEVEL security;
ALTER TABLE "important_events" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS important_event_recipients (
    important_event_recipient_id TEXT NOT NULL,
    important_event_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__important_event_recipients PRIMARY KEY (important_event_recipient_id),
    CONSTRAINT fk__important_event_recipients__important_events FOREIGN KEY (important_event_id) REFERENCES important_events(important_event_id)
);

CREATE POLICY rls_important_event_recipients ON "important_event_recipients" USING (permission_check(resource_path, 'important_event_recipients')) with check (permission_check(resource_path, 'important_event_recipients'));
CREATE POLICY rls_important_event_recipients_restrictive ON "important_event_recipients" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'important_event_recipients')) WITH CHECK (permission_check(resource_path, 'important_event_recipients'));

ALTER TABLE "important_event_recipients" ENABLE ROW LEVEL security;
ALTER TABLE "important_event_recipients" FORCE ROW LEVEL security;