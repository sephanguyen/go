UPDATE
	assignments
SET
	settings = settings || jsonb '{"require_complete_date":false,"require_duration":false,"require_correctness":false,"require_understanding_level":false}'
WHERE
	settings IS NOT NULL AND settings ->> 'require_duration' IS NULL;
