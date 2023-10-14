CREATE OR REPLACE FUNCTION users__auto_update_parent_activation() RETURNS trigger  
LANGUAGE plpgsql 
AS
$$
DECLARE parentIDs TEXT[];
BEGIN
	parentIDs = (SELECT array_agg(parent_id::TEXT) FROM student_parents WHERE student_id = new.user_id);
	CALL update_parent_activation(parentIDs);
RETURN NULL;
END;
$$;

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
	CALL update_parent_activation(parentIDs);
RETURN NULL;
END;
$$;

CREATE trigger student_parents__auto_update_parent_activation AFTER INSERT OR UPDATE of deleted_at
ON public.student_parents FOR EACH ROW
EXECUTE FUNCTION public.student_parents__auto_update_parent_activation();