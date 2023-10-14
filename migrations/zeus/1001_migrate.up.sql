CREATE TABLE IF NOT EXISTS activity_logs (
    activity_log_id text PRIMARY KEY,
    user_id text,
    action_type text,
    status text,
    payload JSONB,
    resource_path text,
    request_at timestamptz,
    created_at timestamptz,
    updated_at timestamptz,
    deleted_at timestamptz
);
