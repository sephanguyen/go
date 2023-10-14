CREATE OR REPLACE FUNCTION public.autofillresourcepath()
 RETURNS text
 LANGUAGE plpgsql
AS $function$
DECLARE
		resource_path text;
BEGIN
	resource_path := current_setting('permission.resource_path', 't');

	RETURN resource_path;
END $function$;


CREATE OR REPLACE FUNCTION public.permission_check(resource_path text, table_name text)
 RETURNS boolean
 LANGUAGE sql
 STABLE
AS $function$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$function$;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    user_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    deactivated_at timestamp with time zone
);

ALTER TABLE ONLY public.users FORCE ROW LEVEL SECURITY;


--
-- Name: locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.locations (
    location_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    location_type text,
    partner_internal_id text,
    partner_internal_parent_id text,
    parent_location_id text,
    is_archived boolean DEFAULT false NOT NULL,
    access_path text
);

ALTER TABLE ONLY public.locations FORCE ROW LEVEL SECURITY;


--
-- Name: user_group; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_group (
    user_group_id text NOT NULL,
    user_group_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    org_location_id text,
    is_system boolean
);

ALTER TABLE ONLY public.user_group FORCE ROW LEVEL SECURITY;


--
-- Name: api_keypair; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.api_keypair (
    public_key text NOT NULL,
    user_id text NOT NULL,
    private_key text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath()
);

ALTER TABLE ONLY public.api_keypair FORCE ROW LEVEL SECURITY;


--
-- Name: granted_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.granted_role (
    granted_role_id text NOT NULL,
    user_group_id text NOT NULL,
    role_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.granted_role FORCE ROW LEVEL SECURITY;


--
-- Name: granted_role_access_path; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.granted_role_access_path (
    granted_role_id text NOT NULL,
    location_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.granted_role_access_path FORCE ROW LEVEL SECURITY;


--
-- Name: permission; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission (
    permission_id text NOT NULL,
    permission_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.permission FORCE ROW LEVEL SECURITY;


--
-- Name: permission_role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permission_role (
    permission_id text NOT NULL,
    role_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.permission_role FORCE ROW LEVEL SECURITY;


--
-- Name: role; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role (
    role_id text NOT NULL,
    role_name text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL,
    is_system boolean
);

ALTER TABLE ONLY public.role FORCE ROW LEVEL SECURITY;


--
-- Name: user_group_member; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_group_member (
    user_id text NOT NULL,
    user_group_id text NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.user_group_member FORCE ROW LEVEL SECURITY;


--
-- Name: granted_permissions; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.granted_permissions AS
 SELECT ugm.user_id,
    p.permission_name,
    l1.location_id,
    ugm.resource_path,
    p.permission_id
   FROM ((((((((public.user_group_member ugm
     JOIN public.user_group ug ON ((ugm.user_group_id = ug.user_group_id)))
     JOIN public.granted_role gr ON ((ug.user_group_id = gr.user_group_id)))
     JOIN public.role r ON ((gr.role_id = r.role_id)))
     JOIN public.permission_role pr ON ((r.role_id = pr.role_id)))
     JOIN public.permission p ON ((p.permission_id = pr.permission_id)))
     JOIN public.granted_role_access_path grap ON ((gr.granted_role_id = grap.granted_role_id)))
     JOIN public.locations l ON ((l.location_id = grap.location_id)))
     JOIN public.locations l1 ON ((l1.access_path ~~ (l.access_path || '%'::text))))
  WHERE ((ugm.deleted_at IS NULL) AND (ug.deleted_at IS NULL) AND (gr.deleted_at IS NULL) AND (r.deleted_at IS NULL) AND (pr.deleted_at IS NULL) AND (p.deleted_at IS NULL) AND (grap.deleted_at IS NULL) AND (l.deleted_at IS NULL) AND (l1.deleted_at IS NULL));


--
-- Name: organization_auths; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organization_auths (
    organization_id integer NOT NULL,
    auth_project_id text NOT NULL,
    auth_tenant_id text NOT NULL
);


--
-- Name: organizations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.organizations (
    organization_id text NOT NULL,
    tenant_id text,
    domain_name text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);


--
-- Name: user_access_paths; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_access_paths (
    user_id text NOT NULL,
    location_id text NOT NULL,
    access_path text,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT public.autofillresourcepath() NOT NULL
);

ALTER TABLE ONLY public.user_access_paths FORCE ROW LEVEL SECURITY;


--
-- Name: api_keypair api_keypair__pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.api_keypair
    ADD CONSTRAINT api_keypair__pk PRIMARY KEY (public_key);


--
-- Name: granted_role granted_role_granted_role_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.granted_role
    ADD CONSTRAINT granted_role_granted_role_id_key UNIQUE (granted_role_id);


--
-- Name: locations locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.locations
    ADD CONSTRAINT locations_pkey PRIMARY KEY (location_id);


--
-- Name: organization_auths organization_auths__pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organization_auths
    ADD CONSTRAINT organization_auths__pk PRIMARY KEY (organization_id, auth_project_id, auth_tenant_id);


--
-- Name: organizations organizations__pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations__pk PRIMARY KEY (organization_id);


--
-- Name: permission_role permission_role__pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission_role
    ADD CONSTRAINT permission_role__pk PRIMARY KEY (permission_id, role_id, resource_path);


--
-- Name: granted_role pk__granted_role; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.granted_role
    ADD CONSTRAINT pk__granted_role PRIMARY KEY (granted_role_id);


--
-- Name: granted_role_access_path pk__granted_role_access_path; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.granted_role_access_path
    ADD CONSTRAINT pk__granted_role_access_path PRIMARY KEY (granted_role_id, location_id);


--
-- Name: permission pk__permission; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permission
    ADD CONSTRAINT pk__permission PRIMARY KEY (permission_id);


--
-- Name: user_group pk__user_group; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_group
    ADD CONSTRAINT pk__user_group PRIMARY KEY (user_group_id);


--
-- Name: user_group_member pk__user_group_member; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_group_member
    ADD CONSTRAINT pk__user_group_member PRIMARY KEY (user_id, user_group_id);


--
-- Name: role role__pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role
    ADD CONSTRAINT role__pk PRIMARY KEY (role_id, resource_path);


--
-- Name: user_access_paths user_access_paths_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_access_paths
    ADD CONSTRAINT user_access_paths_pk PRIMARY KEY (user_id, location_id);


--
-- Name: users users_pk; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pk PRIMARY KEY (user_id);


--
-- Name: granted_role_user_group_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX granted_role_user_group_id_idx ON public.granted_role USING btree (user_group_id);


--
-- Name: idx__permission__permssion_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx__permission__permssion_name ON public.permission USING btree (permission_name);


--
-- Name: idx_user_id_user_group_member; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_id_user_group_member ON public.user_group_member USING btree (user_id);


--
-- Name: locations_access_path_text_pattern_ops_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX locations_access_path_text_pattern_ops_idx ON public.locations USING btree (access_path text_pattern_ops);


--
-- Name: organization_auths__organization_id__idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX organization_auths__organization_id__idx ON public.organization_auths USING btree (((organization_id)::text));


--
-- Name: permission_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX permission_name_idx ON public.permission USING btree (permission_name);


--
-- Name: user_access_paths__location_id__idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_access_paths__location_id__idx ON public.user_access_paths USING btree (location_id);


--
-- Name: user_access_paths__user_id__idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_access_paths__user_id__idx ON public.user_access_paths USING btree (user_id);


--
-- Name: user_group__user_group_name_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_group__user_group_name_idx ON public.user_group USING btree (user_group_name);


--
-- Name: user_group_user_id_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX user_group_user_id_idx ON public.user_group_member USING btree (user_group_id, user_id);


--
-- Name: users__created_at__idx_desc; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users__created_at__idx_desc ON public.users USING btree (created_at DESC);


--
-- Name: users__created_at_desc__user_id_desc__idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users__created_at_desc__user_id_desc__idx ON public.users USING btree (created_at DESC, user_id DESC);


--
-- Name: users_resource_path_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX users_resource_path_idx ON public.users USING btree (resource_path);


--
-- Name: api_keypair; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.api_keypair ENABLE ROW LEVEL SECURITY;

--
-- Name: granted_role; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.granted_role ENABLE ROW LEVEL SECURITY;

--
-- Name: granted_role_access_path; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.granted_role_access_path ENABLE ROW LEVEL SECURITY;

--
-- Name: locations; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.locations ENABLE ROW LEVEL SECURITY;

--
-- Name: permission; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.permission ENABLE ROW LEVEL SECURITY;

--
-- Name: permission_role; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.permission_role ENABLE ROW LEVEL SECURITY;

--
-- Name: api_keypair rls_api_keypair; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_api_keypair ON public.api_keypair USING (public.permission_check(resource_path, 'api_keypair'::text)) WITH CHECK (public.permission_check(resource_path, 'api_keypair'::text));


--
-- Name: api_keypair rls_api_keypair_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_api_keypair_restrictive ON public.api_keypair AS RESTRICTIVE USING (public.permission_check(resource_path, 'api_keypair'::text)) WITH CHECK (public.permission_check(resource_path, 'api_keypair'::text));


--
-- Name: granted_role rls_granted_role; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_granted_role ON public.granted_role USING (public.permission_check(resource_path, 'granted_role'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role'::text));


--
-- Name: granted_role_access_path rls_granted_role_access_path; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_granted_role_access_path ON public.granted_role_access_path USING (public.permission_check(resource_path, 'granted_role_access_path'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role_access_path'::text));


--
-- Name: granted_role_access_path rls_granted_role_access_path_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_granted_role_access_path_restrictive ON public.granted_role_access_path AS RESTRICTIVE USING (public.permission_check(resource_path, 'granted_role_access_path'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role_access_path'::text));


--
-- Name: granted_role rls_granted_role_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_granted_role_restrictive ON public.granted_role AS RESTRICTIVE USING (public.permission_check(resource_path, 'granted_role'::text)) WITH CHECK (public.permission_check(resource_path, 'granted_role'::text));


--
-- Name: locations rls_locations_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_locations_location ON public.locations USING ((location_id IN ( SELECT p.location_id
   FROM public.granted_permissions p
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'master.location.read'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))))))) WITH CHECK ((EXISTS ( SELECT true AS bool
   FROM public.granted_permissions p
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'master.location.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text)))))))));


--
-- Name: locations rls_locations_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_locations_restrictive ON public.locations AS RESTRICTIVE USING (public.permission_check(resource_path, 'locations'::text)) WITH CHECK (public.permission_check(resource_path, 'locations'::text));


--
-- Name: permission rls_permission; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_permission ON public.permission USING (public.permission_check(resource_path, 'permission'::text)) WITH CHECK (public.permission_check(resource_path, 'permission'::text));


--
-- Name: permission rls_permission_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_permission_restrictive ON public.permission AS RESTRICTIVE USING (public.permission_check(resource_path, 'permission'::text)) WITH CHECK (public.permission_check(resource_path, 'permission'::text));


--
-- Name: permission_role rls_permission_role; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_permission_role ON public.permission_role USING (public.permission_check(resource_path, 'permission_role'::text)) WITH CHECK (public.permission_check(resource_path, 'permission_role'::text));


--
-- Name: permission_role rls_permission_role_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_permission_role_restrictive ON public.permission_role AS RESTRICTIVE USING (public.permission_check(resource_path, 'permission_role'::text)) WITH CHECK (public.permission_check(resource_path, 'permission_role'::text));


--
-- Name: role rls_role; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_role ON public.role USING (public.permission_check(resource_path, 'role'::text)) WITH CHECK (public.permission_check(resource_path, 'role'::text));


--
-- Name: role rls_role_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_role_restrictive ON public.role AS RESTRICTIVE USING (public.permission_check(resource_path, 'role'::text)) WITH CHECK (public.permission_check(resource_path, 'role'::text));


--
-- Name: user_access_paths rls_user_access_paths_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_access_paths_location ON public.user_access_paths USING ((location_id IN ( SELECT p.location_id
   FROM public.granted_permissions p
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.read'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))))))) WITH CHECK ((location_id IN ( SELECT p.location_id
   FROM public.granted_permissions p
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text)))))))));


--
-- Name: user_access_paths rls_user_access_paths_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_access_paths_restrictive ON public.user_access_paths AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_access_paths'::text)) WITH CHECK (public.permission_check(resource_path, 'user_access_paths'::text));


--
-- Name: user_group rls_user_group; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_group ON public.user_group USING (public.permission_check(resource_path, 'user_group'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group'::text));


--
-- Name: user_group_member rls_user_group_member; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_group_member ON public.user_group_member USING (public.permission_check(resource_path, 'user_group_member'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group_member'::text));


--
-- Name: user_group_member rls_user_group_member_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_group_member_restrictive ON public.user_group_member AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_group_member'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group_member'::text));


--
-- Name: user_group rls_user_group_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_user_group_restrictive ON public.user_group AS RESTRICTIVE USING (public.permission_check(resource_path, 'user_group'::text)) WITH CHECK (public.permission_check(resource_path, 'user_group'::text));


--
-- Name: users rls_users_delete_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_delete_location ON public.users FOR DELETE USING ((true <= ( SELECT true AS bool
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.user_id = users.user_id) AND (usp.deleted_at IS NULL))
 LIMIT 1)));


--
-- Name: users rls_users_insert_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_insert_location ON public.users FOR INSERT WITH CHECK ((1 = 1));


--
-- Name: users rls_users_permission_v4; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_permission_v4 ON public.users USING ((current_setting('app.user_id'::text) = user_id)) WITH CHECK ((current_setting('app.user_id'::text) = user_id));


--
-- Name: users rls_users_restrictive; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_restrictive ON public.users AS RESTRICTIVE USING (public.permission_check(resource_path, 'users'::text)) WITH CHECK (public.permission_check(resource_path, 'users'::text));


--
-- Name: users rls_users_select_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_select_location ON public.users FOR SELECT USING ((true <= ( SELECT true AS bool
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.read'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.user_id = users.user_id) AND (usp.deleted_at IS NULL))
 LIMIT 1)));


--
-- Name: users rls_users_update_location; Type: POLICY; Schema: public; Owner: -
--

CREATE POLICY rls_users_update_location ON public.users FOR UPDATE USING ((true <= ( SELECT true AS bool
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.user_id = users.user_id) AND (usp.deleted_at IS NULL))
 LIMIT 1))) WITH CHECK ((true <= ( SELECT true AS bool
   FROM (public.granted_permissions p
     JOIN public.user_access_paths usp ON ((usp.location_id = p.location_id)))
  WHERE ((p.user_id = current_setting('app.user_id'::text)) AND (p.permission_id = ( SELECT p2.permission_id
           FROM public.permission p2
          WHERE ((p2.permission_name = 'user.user.write'::text) AND (p2.resource_path = current_setting('permission.resource_path'::text))))) AND (usp.user_id = users.user_id) AND (usp.deleted_at IS NULL))
 LIMIT 1)));


--
-- Name: role; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.role ENABLE ROW LEVEL SECURITY;

--
-- Name: user_access_paths; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_access_paths ENABLE ROW LEVEL SECURITY;

--
-- Name: user_group; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_group ENABLE ROW LEVEL SECURITY;

--
-- Name: user_group_member; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.user_group_member ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: -
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
