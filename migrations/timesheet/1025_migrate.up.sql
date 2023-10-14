CREATE TABLE IF NOT EXISTS public.auto_create_flag_activity_log (
    id              TEXT NOT NULL CONSTRAINT pk__auto_create_flag_activity_log PRIMARY KEY,
    staff_id        TEXT NOT NULL,
    start_time      TIMESTAMP with time zone NOT NULL,
    end_time        TIMESTAMP with time zone,
    flag_on         BOOLEAN  NOT NULL,
    created_at      TIMESTAMP with time zone NOT NULL,
    updated_at      TIMESTAMP with time zone NOT NULL,
    deleted_at      TIMESTAMP with time zone,
    resource_path   TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT fk__auto_create_flag_activity_log_staff_id__staff_staff_id FOREIGN KEY (staff_id) REFERENCES public.staff(staff_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx__auto_create_flag_activity_log_time ON public.auto_create_flag_activity_log
USING btree(staff_id,start_time,end_time)
WHERE (deleted_at IS NULL);

CREATE POLICY rls_auto_create_flag_activity_log ON "auto_create_flag_activity_log"
    USING (permission_check (resource_path, 'auto_create_flag_activity_log'))
    WITH CHECK (permission_check (resource_path, 'auto_create_flag_activity_log'));

ALTER TABLE "auto_create_flag_activity_log" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "auto_create_flag_activity_log" FORCE ROW LEVEL SECURITY;