CREATE TABLE public.lms_study_plans (
    study_plan_id text NOT NULL,
    name text NOT NULL,
    course_id text NOT NULL,
    academic_year integer,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE',
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    PRIMARY KEY(study_plan_id)
);

ALTER TABLE lms_study_plans ADD CONSTRAINT study_plan_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])));

/* set RLS */
CREATE POLICY rls_lms_study_plans ON "lms_study_plans" using (
    permission_check(resource_path, 'lms_study_plans')
    ) with check (
    permission_check(resource_path, 'lms_study_plans')
    );

CREATE POLICY rls_lms_study_plans_restrictive ON "lms_study_plans" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lms_study_plans')
    ) with check (
    permission_check(resource_path, 'lms_study_plans')
    );

ALTER TABLE "lms_study_plans" ENABLE ROW LEVEL security;
ALTER TABLE "lms_study_plans" FORCE ROW LEVEL security;

CREATE TABLE public.lms_student_study_plans (
    student_id text NOT NULL,
    study_plan_id text NOT NULL,
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE',
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    PRIMARY KEY(student_id,study_plan_id),
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT fk_study_plan_id_lms_student_study_plans FOREIGN KEY(study_plan_id) REFERENCES lms_study_plans(study_plan_id)
);

ALTER TABLE lms_student_study_plans ADD CONSTRAINT student_study_plans_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])));

/* set RLS */
CREATE POLICY rls_lms_student_study_plans ON "lms_student_study_plans" using (
    permission_check(resource_path, 'lms_student_study_plans')
    ) with check (
    permission_check(resource_path, 'lms_student_study_plans')
    );

CREATE POLICY rls_lms_student_study_plans_restrictive ON "lms_student_study_plans" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lms_student_study_plans')
    ) with check (
    permission_check(resource_path, 'lms_student_study_plans')
    );

ALTER TABLE "lms_student_study_plans" ENABLE ROW LEVEL security;
ALTER TABLE "lms_student_study_plans" FORCE ROW LEVEL security;

CREATE TABLE public.lms_learning_material_list (
    lm_list_id text NOT NULL,
    lm_ids text[],
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    PRIMARY KEY(lm_list_id)
);

/* set RLS */
CREATE POLICY rls_lms_learning_material_list ON "lms_learning_material_list" using (
    permission_check(resource_path, 'lms_learning_material_list')
    ) with check (
    permission_check(resource_path, 'lms_learning_material_list')
    );

CREATE POLICY rls_lms_learning_material_list_restrictive ON "lms_learning_material_list" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lms_learning_material_list')
    ) with check (
    permission_check(resource_path, 'lms_learning_material_list')
    );

ALTER TABLE "lms_learning_material_list" ENABLE ROW LEVEL security;
ALTER TABLE "lms_learning_material_list" FORCE ROW LEVEL security;

CREATE TABLE public.lms_student_study_plan_item (
    student_id text NOT NULL,
    lm_list_id text NOT NULL,
    study_plan_id text NOT NULL,
    type text default 'STATIC',
    status text DEFAULT 'STUDY_PLAN_STATUS_ACTIVE',
    display_order integer DEFAULT 0 NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    PRIMARY KEY(student_id,lm_list_id),
    CONSTRAINT fk_study_plan_id_lms_student_study_plan_item FOREIGN KEY(study_plan_id) REFERENCES lms_study_plans(study_plan_id),
    CONSTRAINT fk_lm_list_id_lms_student_study_plan_item FOREIGN KEY(lm_list_id) REFERENCES lms_learning_material_list(lm_list_id)
);

ALTER TABLE lms_student_study_plan_item ADD CONSTRAINT lms_student_study_plan_item_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])));

/* set RLS */
CREATE POLICY rls_lms_student_study_plan_item ON "lms_student_study_plan_item" using (
    permission_check(resource_path, 'lms_student_study_plan_item')
    ) with check (
    permission_check(resource_path, 'lms_student_study_plan_item')
    );

CREATE POLICY rls_lms_student_study_plan_item_restrictive ON "lms_student_study_plan_item" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lms_student_study_plan_item')
    ) with check (
    permission_check(resource_path, 'lms_student_study_plan_item')
    );

ALTER TABLE "lms_student_study_plan_item" ENABLE ROW LEVEL security;
ALTER TABLE "lms_student_study_plan_item" FORCE ROW LEVEL security;

CREATE TABLE public.lms_study_plan_items (
    study_plan_item_id text NOT NULL,
    study_plan_id text NOT NULL,
    lm_list_id text NOT NULL,
    name text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    display_order integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'STUDY_PLAN_ITEM_STATUS_ACTIVE',
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    PRIMARY KEY(study_plan_item_id),
    CONSTRAINT fk_study_plan_id_lms_study_plan_items FOREIGN KEY(study_plan_id) REFERENCES lms_study_plans(study_plan_id),
    CONSTRAINT fk_lm_list_id_lms_study_plan_items FOREIGN KEY(lm_list_id) REFERENCES lms_learning_material_list(lm_list_id)
);

ALTER TABLE lms_study_plan_items ADD CONSTRAINT study_plan_item_status_check CHECK ((status = ANY (ARRAY['STUDY_PLAN_STATUS_NONE'::text, 'STUDY_PLAN_STATUS_ACTIVE'::text, 'STUDY_PLAN_STATUS_ARCHIVED'::text])));

/* set RLS */
CREATE POLICY rls_lms_study_plan_items ON "lms_study_plan_items" using (
    permission_check(resource_path, 'lms_study_plan_items')
    ) with check (
    permission_check(resource_path, 'lms_study_plan_items')
    );

CREATE POLICY rls_lms_study_plan_items_restrictive ON "lms_study_plan_items" AS RESTRICTIVE TO PUBLIC using (
    permission_check(resource_path, 'lms_study_plan_items')
    ) with check (
    permission_check(resource_path, 'lms_study_plan_items')
    );

ALTER TABLE "lms_study_plan_items" ENABLE ROW LEVEL security;
ALTER TABLE "lms_study_plan_items" FORCE ROW LEVEL security;