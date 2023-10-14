CREATE TABLE IF NOT EXISTS emails (
    email_id TEXT NOT NULL,
    sg_message_id TEXT DEFAULT NULL,
    subject TEXT DEFAULT NULL,
    content JSONB DEFAULT NULL,
    email_from TEXT DEFAULT NULL,
    status TEXT DEFAULT 'EMAIL_STATUS_QUEUED',
    email_recipients TEXT[] DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__emails PRIMARY KEY (email_id)
);

CREATE POLICY rls_emails ON "emails" USING (permission_check(resource_path, 'emails')) with check (permission_check(resource_path, 'emails'));
CREATE POLICY rls_emails_restrictive ON "emails" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'emails')) WITH CHECK (permission_check(resource_path, 'emails'));

ALTER TABLE "emails" ENABLE ROW LEVEL security;
ALTER TABLE "emails" FORCE ROW LEVEL security;

ALTER TABLE ONLY public.emails DROP CONSTRAINT IF EXISTS email_status_type_check;
ALTER TABLE public.emails ADD CONSTRAINT email_status_type_check CHECK (status = ANY (ARRAY[
		'EMAIL_STATUS_NONE',
		'EMAIL_STATUS_QUEUED',
		'EMAIL_STATUS_INTERNAL_FAILED',
		'EMAIL_STATUS_PROCESSED'
]::TEXT[]));

CREATE TABLE IF NOT EXISTS email_recipients (
    id TEXT NOT NULL,
    email_id TEXT NOT NULL,
    recipient_address TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__email_recipients PRIMARY KEY (id),
    CONSTRAINT fk__email_recipients__emails FOREIGN KEY (email_id) REFERENCES emails(email_id)
);

CREATE POLICY rls_email_recipients ON "email_recipients" USING (permission_check(resource_path, 'email_recipients')) with check (permission_check(resource_path, 'email_recipients'));
CREATE POLICY rls_email_recipients_restrictive ON "email_recipients" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'email_recipients')) WITH CHECK (permission_check(resource_path, 'email_recipients'));

ALTER TABLE "email_recipients" ENABLE ROW LEVEL security;
ALTER TABLE "email_recipients" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS email_recipient_events (
    id TEXT NOT NULL,
    email_recipient_id TEXT NOT NULL,
    sg_event_id TEXT NOT NULL,
    type TEXT NOT NULL,
    event TEXT NOT NULL,
    description JSONB DEFAULT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT pk__email_recipient_events PRIMARY KEY (id),
    CONSTRAINT fk__email_recipient_events__email_recipients FOREIGN KEY (email_recipient_id) REFERENCES email_recipients(id)
);

CREATE POLICY rls_email_recipient_events ON "email_recipient_events" USING (permission_check(resource_path, 'email_recipient_events')) with check (permission_check(resource_path, 'email_recipient_events'));
CREATE POLICY rls_email_recipient_events_restrictive ON "email_recipient_events" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'email_recipient_events')) WITH CHECK (permission_check(resource_path, 'email_recipient_events'));

ALTER TABLE "email_recipient_events" ENABLE ROW LEVEL security;
ALTER TABLE "email_recipient_events" FORCE ROW LEVEL security;

ALTER TABLE public.email_recipient_events ADD CONSTRAINT email_recipient_event_type_check CHECK (type = ANY (ARRAY[
		'EMAIL_EVENT_TYPE_NONE',
		'EMAIL_EVENT_TYPE_DELIVERY',
		'EMAIL_EVENT_TYPE_ENGAGEMENT'
]::TEXT[]));

ALTER TABLE public.email_recipient_events ADD CONSTRAINT email_recipient_event_check CHECK (event = ANY (ARRAY[
		'EMAIL_EVENT_NONE',
		'EMAIL_EVENT_PROCESSED',
		'EMAIL_EVENT_DROPPED',
		'EMAIL_EVENT_DELIVERED',
		'EMAIL_EVENT_DEFERRED',
		'EMAIL_EVENT_BOUNCE',
		'EMAIL_EVENT_BLOCKED',
		'EMAIL_EVENT_OPEN',
		'EMAIL_EVENT_CLICK',
		'EMAIL_EVENT_SPAM_REPORT',
		'EMAIL_EVENT_UNSUBSCRIBE',
		'EMAIL_EVENT_GROUP_UNSUBSCRIBE',
		'EMAIL_EVENT_GROUP_RESUBSCRIBE'
]::TEXT[]));


