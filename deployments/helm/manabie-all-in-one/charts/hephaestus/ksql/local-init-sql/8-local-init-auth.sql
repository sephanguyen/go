\connect auth;

-- CREATE TABLE public.users (
--                               user_id text NOT NULL,
--                               created_at timestamp with time zone NOT NULL,
--                               updated_at timestamp with time zone NOT NULL,
--                               deleted_at timestamp with time zone,
--                               resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
--                               deactivated_at timestamp with time zone
-- );

CREATE FUNCTION capture_user_change()
    RETURNS trigger AS
$$
DECLARE
    message text;
BEGIN
    message := row_to_json(NEW);
    RAISE NOTICE '%', format('sending message for %s, %s', TG_OP, message);
    EXECUTE FORMAT('NOTIFY synced_auth_user, ''%s''', message);
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER auth_user_crud_trigger AFTER INSERT OR UPDATE
    ON public.users
    FOR EACH ROW EXECUTE PROCEDURE capture_user_change();
