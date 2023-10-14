ALTER TABLE public.student_orders DROP CONSTRAINT IF EXISTS student_orders_packages_fk;
ALTER TABLE public.package_items DROP CONSTRAINT IF EXISTS package_item__packages_fk;

-- =========================================================================================
-- =================================== public.packages table ===============================
-- =========================================================================================
DROP TABLE IF EXISTS public.packages;

CREATE TABLE IF NOT EXISTS public.packages (
    package_id TEXT NOT NULL,
    country TEXT,
    "name" TEXT NOT NULL,
    descriptions TEXT[],
    price INTEGER NOT NULL,
    discounted_price INTEGER,
    start_at TIMESTAMP WITH TIME ZONE,
    end_at TIMESTAMP WITH TIME ZONE,
    duration INTEGER, 
    prioritize_level INTEGER DEFAULT 0,
    properties JSONB NOT NULL, 
    is_recommended BOOLEAN NOT NULL,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT pk__packages PRIMARY KEY (package_id)
);

CREATE POLICY rls_packages ON "packages"
USING (permission_check(resource_path, 'packages')) WITH CHECK (permission_check(resource_path, 'packages'));
CREATE POLICY rls_packages_restrictive ON "packages" AS RESTRICTIVE
USING (permission_check(resource_path, 'packages')) WITH CHECK (permission_check(resource_path, 'packages'));

ALTER TABLE "packages" ENABLE ROW LEVEL security;
ALTER TABLE "packages" FORCE ROW LEVEL security;
