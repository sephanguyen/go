ALTER TABLE "history"
ADD COLUMN "integration" BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "target_coverage"
ADD COLUMN "integration" BOOLEAN NOT NULL DEFAULT false;

ALTER TABLE "target_coverage"
DROP CONSTRAINT "target_coverage_branch_name_repository_key";

ALTER TABLE "target_coverage"
ADD UNIQUE ("branch_name", "repository", "integration");
