create or replace procedure update_parent_activation(parentIDs text[])  
security invoker
as
$$
declare
	configValue text;
	count_active_student int;
	parentID text;
begin
	-- check config enable
	configValue = (select config_value from internal_configuration_value where config_key = 'user.student_management.deactivate_parent');
	if configValue is null or configValue != 'on' then
		return;
	end if;

	if parentIDs is null then
		return;
	end if;

	foreach parentID in array parentIDs
	loop
		count_active_student = (select count(*) 
			from student_parents sp
			inner join users u on sp.student_id = u.user_id
			where sp.parent_id = parentID
      			and u.deactivated_at is null
      			and sp.deleted_at is null);
      	-- if count_active_student means all of parent's children are deactivated, so deactivate parent
      	-- else means at least 1 parent's children are active, so update parent to active
		if count_active_student = 0 then 
			update users set deactivated_at = now()
			where user_id = parentID;
		else
			update users set deactivated_at = null
			where user_id = parentID;
		end if;
	end loop;
return;
end
$$
language 'plpgsql';
