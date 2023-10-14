CREATE EXTENSION IF NOT EXISTS pg_trgm;

DROP INDEX IF EXISTS learning_material_name_idx_gin_trgm;

DROP INDEX IF EXISTS learning_material_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS learning_material_name_gist_trgm_idx ON learning_material USING gist (name gist_trgm_ops);

DROP INDEX IF EXISTS learning_objective_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS learning_objective_name_gist_trgm_idx ON learning_objective USING gist (name gist_trgm_ops);

DROP INDEX IF EXISTS exam_lo_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS exam_lo_name_gist_trgm_idx ON exam_lo USING gist (name gist_trgm_ops);

DROP INDEX IF EXISTS flash_card_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS flash_card_name_gist_trgm_idx ON flash_card USING gist (name gist_trgm_ops);

DROP INDEX IF EXISTS assignment_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS assignment_name_gist_trgm_idx ON assignment USING gist (name gist_trgm_ops);

DROP INDEX IF EXISTS task_assignment_name_gist_trgm_idx;
CREATE INDEX IF NOT EXISTS task_assignment_name_gist_trgm_idx ON task_assignment USING gist (name gist_trgm_ops);