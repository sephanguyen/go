ALTER TABLE assignments
ADD COLUMN IF NOT EXISTS is_required_grade bool;
