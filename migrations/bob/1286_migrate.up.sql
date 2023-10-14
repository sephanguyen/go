ALTER TABLE "lesson_members"
  ADD COLUMN IF NOT EXISTS user_first_name TEXT,
  ADD COLUMN IF NOT EXISTS user_last_name TEXT;

ALTER TABLE "lesson_student_subscriptions"
  ADD COLUMN IF NOT EXISTS student_first_name TEXT,
  ADD COLUMN IF NOT EXISTS student_last_name TEXT;

ALTER TABLE "lessons_teachers"
  ADD COLUMN IF NOT EXISTS teacher_name TEXT;