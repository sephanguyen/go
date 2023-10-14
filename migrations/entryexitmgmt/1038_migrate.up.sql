Create TABLE entryexit_queue(
    entryexit_queue_id text NOT NULL,
    student_id text NOT NULL,
    resource_path text DEFAULT autofillresourcepath(),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT entryexit_queue_id__pk PRIMARY KEY (entryexit_queue_id),
    CONSTRAINT entryexit_queue_id__students__fk FOREIGN KEY (student_id) REFERENCES "students"(student_id)
);


CREATE POLICY rls_entryexit_queue ON "entryexit_queue" 
USING (permission_check(resource_path, 'entryexit_queue')) WITH CHECK (permission_check(resource_path, 'entryexit_queue'));

CREATE POLICY rls_entryexit_queue_restrictive ON "entryexit_queue" AS RESTRICTIVE TO PUBLIC
USING (permission_check(resource_path, 'entryexit_queue')) WITH CHECK (permission_check(resource_path, 'entryexit_queue'));


ALTER TABLE "entryexit_queue" ENABLE ROW LEVEL security;
ALTER TABLE "entryexit_queue" FORCE ROW LEVEL security;

CREATE OR REPLACE FUNCTION check_unique_entryexit_queue()
    RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM entryexit_queue
        WHERE student_id = NEW.student_id AND created_at >= NEW.created_at - interval '5 seconds'
    ) THEN
        RAISE EXCEPTION 'Cannot insert or update. Another record with id = % was created within the last 5 seconds.', NEW.student_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_check_unique_entryexit_queue
BEFORE INSERT OR UPDATE ON entryexit_queue
FOR EACH ROW
EXECUTE FUNCTION check_unique_entryexit_queue();