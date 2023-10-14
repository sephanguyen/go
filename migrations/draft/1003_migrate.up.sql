ALTER TABLE "target_coverage"
DROP CONSTRAINT "target_coverage_repository_key";

ALTER TABLE "target_coverage"
ADD UNIQUE ("branch_name", "repository");

ALTER TABLE "history"
ADD COLUMN "target_branch_name" TEXT;
