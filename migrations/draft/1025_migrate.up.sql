CREATE TABLE IF NOT EXISTS "skipped_bdd_test" (
    id SERIAL PRIMARY KEY,
    repository VARCHAR(255) NOT NULL,
    feature_path VARCHAR(1000) NOT NULL,
    scenario_name VARCHAR(1000),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
    CONSTRAINT feature_scenario_unique UNIQUE (feature_path, scenario_name)
);
