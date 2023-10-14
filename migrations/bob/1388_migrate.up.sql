UPDATE users
SET is_system = TRUE,
    deleted_at = NULL
WHERE name = ANY(ARRAY[
    'Notification Schedule Job',
    'Payment Schedule Job'
])
