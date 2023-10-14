#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

set role postgres;
do $$
declare 
	rec record; query text;
	rec_location record; rec_location_type record;
	org_location text; org_type text;
	rec_loc record; query_loc text; parent_id text; parent_access_path text;
begin
	create temporary table _internal_mapping (prod_rp text, internal_rp text)
	on commit drop;
	insert into _internal_mapping values
		('-2147483646','2147483646'),
		('-2147483631','2147483631'),
		('-2147483630','2147483630'),
		('-2147483629','2147483629'),
		('-2147483626','2147483626'),
		('-2147483625','2147483625'),
		('-2147483624','2147483624');
	
	query= 'select internal_rp, prod_rp from _internal_mapping;';

	for rec in execute query
	loop
		raise notice '
----------------------------------------------
*** Start clone location & location_type for %

[1] - Check location & location_type are empty', rec.internal_rp;
		--check location type is empty
		execute 'select array_agg(location_type_id) ids, count(*) c from location_types
				where level=0 and name =''org'' and deleted_at is null
				and resource_path='''||rec.internal_rp||'''' into rec_location_type;
		if rec_location_type.c!=1 then
			raise notice '% location_type is not empty to init (%)',rec.internal_rp, rec_location_type.c;
			continue when true;
		else
			org_type=rec_location_type.ids[1];
			raise notice 'org_type_id: %', org_type;
		end if;
	
		--check location type is empty
		execute 'select array_agg(location_id) ids, count(*) c, 
						count(*) filter (where location_type='''|| org_type ||''') co
				from locations where deleted_at is null
				and resource_path='''||rec.internal_rp||'''' into rec_location;
		if rec_location.c!=1 or rec_location.co!=1 then
			raise notice '% location is not empty to init (%, %)',rec.internal_rp, rec_location.c,rec_location.co;
			continue when true;
		else
			org_location=rec_location.ids[1];
			raise notice 'org_location_id: %', org_location;
		end if;
		raise notice 'Location & location_type are empty are empty: OK!
		';
	
		raise notice '
[2] - Cloning location_type...';
		insert into location_types (location_type_id,"name",display_name,updated_at,created_at,resource_path,"level")
		select generate_ulid(),lt."name",lt.display_name,now(),now(),rec.internal_rp,lt."level"
		from location_types lt where lt.resource_path = rec.prod_rp and lt.deleted_at is null and lt."level" <>0;
		
		raise notice '
[3] - Cloning location...';
		INSERT INTO locations (location_id,"name",created_at,updated_at,resource_path,partner_internal_id,partner_internal_parent_id,parent_location_id,access_path, location_type)
		select generate_ulid(), l."name", now(), l.updated_at, rec.internal_rp, l.partner_internal_id , l.partner_internal_parent_id, l.parent_location_id, l.access_path,
			-- replace location type by the new one created in previous step
			(select lt.location_type_id  
				from location_types lt 
				where lt.resource_path = rec.internal_rp and lt.deleted_at is null
				and (lt."name", lt."level") =
					(select lt2."name", lt2."level" from location_types lt2 where lt2.location_type_id = l.location_type and lt2.resource_path=rec.prod_rp and lt2.deleted_at is null)
			)
		from locations l where l.resource_path = rec.prod_rp and l.deleted_at is null and l.parent_location_id is not null;
		
		raise notice '
[4] - Replace location''s parent & access_path ...';
		query_loc= 'select * from locations where resource_path = ''' || rec.internal_rp 
					|| ''' and deleted_at is null and location_id <>''' || org_location ||''' order by access_path';
	
		for rec_loc in execute query_loc
		loop
			raise notice '';
			raise notice 'replace parent and access_path for locaiton: % - %', rec_loc.name, rec_loc.location_id;
			
			parent_id = org_location;
			parent_access_path = org_location;
			if rec_loc.partner_internal_parent_id is not null then
				raise notice 'find parent: %', rec_loc.partner_internal_parent_id;
				execute 'select * from locations where resource_path = ''' || rec.internal_rp 
					|| ''' and partner_internal_id='''||rec_loc.partner_internal_parent_id||'''' into rec_location;
				parent_id = rec_location.location_id;
				parent_access_path = rec_location.access_path;
			end if;
			raise notice 'replace by: % - %', parent_id, parent_access_path;
			update locations 
				set parent_location_id = parent_id, access_path = parent_access_path || '/'|| location_id 
				where location_id = rec_loc.location_id and resource_path = rec.internal_rp;
	   end loop;
   end loop;	
end; 

$$;	

EOF
