CREATE TABLE IF NOT EXISTS public.permission_role (
    permission_id TEXT NOT NULL,
    role_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__permission_role PRIMARY KEY (permission_id, role_id, resource_path)
);
CREATE POLICY rls_permission_role ON public.permission_role
    USING (permission_check(resource_path, 'permission_role'))
    WITH CHECK (permission_check(resource_path, 'permission_role'));
CREATE POLICY rls_permission_role_restrictive ON "permission_role"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'permission_role'))
    WITH CHECK (permission_check(resource_path, 'permission_role'));
ALTER TABLE public.permission_role ENABLE ROW LEVEL security;
ALTER TABLE public.permission_role FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.granted_role_access_path (
    granted_role_id TEXT NOT NULL,
    location_id TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id)
);
CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path
    USING (permission_check(resource_path, 'granted_role_access_path'))
    WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
CREATE POLICY rls_granted_role_access_path_restrictive ON "granted_role_access_path" AS RESTRICTIVE TO PUBLIC
    USING (permission_check(resource_path,'granted_role_access_path'))
    WITH CHECK (permission_check(resource_path, 'granted_role_access_path'));
ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL security;
ALTER TABLE public.granted_role_access_path FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.permission (
    permission_id TEXT NOT NULL,
    permission_name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath() NOT NULL,

    CONSTRAINT pk__permission PRIMARY KEY (permission_id)
);
CREATE POLICY rls_permission ON public.permission
    USING (permission_check(resource_path, 'permission'))
    WITH CHECK (permission_check(resource_path, 'permission'));
CREATE POLICY rls_permission_restrictive ON "permission"
    AS RESTRICTIVE TO public
    USING (permission_check(resource_path, 'permission'))
    WITH CHECK (permission_check(resource_path, 'permission'));
ALTER TABLE public.permission ENABLE ROW LEVEL security;
ALTER TABLE public.permission FORCE ROW LEVEL security;

create or replace
view public.granted_permissions
as
select
	ugm.user_id,
	p.permission_name,
	l1.location_id,
	ugm.resource_path,
	p.permission_id
from
	user_group_member ugm
join user_group ug on
	ugm.user_group_id = ug.user_group_id
join granted_role gr on
	ug.user_group_id = gr.user_group_id
join role r on
	gr.role_id = r.role_id
join permission_role pr on
	r.role_id = pr.role_id
join permission p on
	p.permission_id = pr.permission_id
join granted_role_access_path grap on
	gr.granted_role_id = grap.granted_role_id
join locations l on
	l.location_id = grap.location_id
join locations l1 on
	l1.access_path ~~ (l.access_path || '%'::text)
where
	ugm.deleted_at is null
	and ug.deleted_at is null
	and gr.deleted_at is null
	and r.deleted_at is null
	and pr.deleted_at is null
	and p.deleted_at is null
	and grap.deleted_at is null
	and l.deleted_at is null
	and l1.deleted_at is null;

DROP POLICY IF EXISTS rls_allocate_marker_location on "allocate_marker";

CREATE POLICY rls_allocate_marker_location ON "allocate_marker" AS RESTRICTIVE FOR ALL TO PUBLIC
using (
1 <= (
	select			
		count(*)
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = ANY(
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'syllabus.allocate_marker.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	limit 1
	)
)
with check (
1 <= (
	select			
		count(*)
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = ANY(
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'syllabus.allocate_marker.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = allocate_marker.created_by
		and usp.deleted_at is null
	limit 1
	)
);
