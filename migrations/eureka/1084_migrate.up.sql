CREATE TABLE IF NOT EXISTS public.master_study_plan (
    study_plan_id text NOT NULL,
    learning_material_id text NOT NULL,
    status text NOT NULL,
    start_date timestamp with time zone,
    end_date timestamp with time zone,
    available_from timestamp with time zone,
    available_to timestamp with time zone,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    school_date timestamp with time zone,    
    resource_path TEXT,
    CONSTRAINT learning_material_id_fk FOREIGN KEY (learning_material_id) REFERENCES public.learning_material (learning_material_id),
    CONSTRAINT study_plan_id_fk FOREIGN KEY (study_plan_id) REFERENCES public.study_plans (study_plan_id),
    CONSTRAINT learning_material_id_study_plan_id_pk PRIMARY KEY (learning_material_id,study_plan_id)
);

/* set RLS */
CREATE POLICY rls_master_study_plan ON "master_study_plan" using (permission_check(resource_path, 'master_study_plan')) with check (permission_check(resource_path, 'master_study_plan'));
ALTER TABLE "master_study_plan" ENABLE ROW LEVEL security;
ALTER TABLE "master_study_plan" FORCE ROW LEVEL security;
