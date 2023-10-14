CREATE OR REPLACE PROCEDURE update_parent_activation(parentIDs text[], org_id text)  
SECURITY invoker
AS
$$
DECLARE
	configValue text;
	count_active_student int;
	parentID text;
BEGIN
	-- check config enable
	configValue = (
        SELECT icv.config_value FROM internal_configuration_value icv
        WHERE icv.config_key = 'user.student_management.deactivate_parent'
        AND icv.resource_path = org_id
    );
	IF configValue IS NULL OR configValue != 'on' THEN
		RETURN;
	END IF;

	IF parentIDs IS NULL THEN
		RETURN;
	END IF;

	FOREACH parentID IN ARRAY parentIDs
	LOOP
		count_active_student = (SELECT count(*) 
			FROM student_parents sp
			INNER JOIN users u ON sp.student_id = u.user_id
			WHERE sp.parent_id = parentID
      			AND u.deactivated_at IS NULL
      			AND sp.deleted_at IS NULL);
      	-- if count_active_student means all of parent's children are deactivated, so deactivate parent
      	-- else means at least 1 parent's children are active, so update parent to active
		IF count_active_student = 0 THEN 
			UPDATE users SET deactivated_at = NOW()
			WHERE user_id = parentID;
		ELSE
			UPDATE users SET deactivated_at = NULL
			WHERE user_id = parentID;
		END IF;
	END LOOP;
RETURN;
END
$$
LANGUAGE 'plpgsql';

CREATE OR REPLACE FUNCTION users__auto_update_parent_activation() RETURNS trigger  
LANGUAGE plpgsql 
AS
$$
DECLARE parentIDs TEXT[];
BEGIN
	parentIDs = (SELECT array_agg(parent_id::TEXT) FROM student_parents WHERE student_id = new.user_id);
	CALL update_parent_activation(parentIDs, new.resource_path);
RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS users__auto_update_parent_activation ON public.users;
CREATE trigger users__auto_update_parent_activation AFTER UPDATE of deactivated_at
ON public.users FOR EACH ROW
EXECUTE FUNCTION public.users__auto_update_parent_activation();

-- student_parents
CREATE OR REPLACE FUNCTION student_parents__auto_update_parent_activation() RETURNS trigger  
LANGUAGE plpgsql 
AS
$$
DECLARE parentIDs text[];
BEGIN
	parentIDs = (SELECT array_agg(new.parent_id::TEXT));
	CALL update_parent_activation(parentIDs, new.resource_path);
RETURN NULL;
END;
$$;

DROP TRIGGER IF EXISTS student_parents__auto_update_parent_activation ON public.student_parents;
CREATE trigger student_parents__auto_update_parent_activation AFTER INSERT OR UPDATE of deleted_at
ON public.student_parents FOR EACH ROW
EXECUTE FUNCTION public.student_parents__auto_update_parent_activation();
