CREATE SCHEMA IF NOT EXISTS timesheet;

CREATE TABLE IF NOT EXISTS timesheet.ts_transportation
(
    transportation_expense_id         TEXT PRIMARY KEY,
    timesheet_id                      TEXT        NOT NULL,
    transportation_type               TEXT        NOT NULL,
    transportation_from               TEXT        NOT NULL,
    transportation_to                 TEXT        NOT NULL,
    cost_amount                       INTEGER     NOT NULL,
    round_trip                        BOOLEAN     NOT NULL,
    transportation_expense_remarks    TEXT        NOT NULL,
    transportation_expense_created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    transportation_expense_updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    transportation_expense_deleted_at timestamptz,
    timesheet_status                  TEXT        NOT NULL,
    timesheet_date                    timestamptz NOT NULL,
    timesheet_remark                  TEXT        NOT NULL,
    timesheet_created_at              timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at              timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at              timestamptz,
    staff_id                          TEXT        NOT NULL
);

CREATE TABLE IF NOT EXISTS timesheet.staff_transportation_expense
(
    staff_transportation_expense_id TEXT PRIMARY KEY,
    staff_id                        TEXT        NOT NULL,
    location_id                     TEXT        NOT NULL,
    transport_type                  TEXT        NOT NULL,
    transportation_from             TEXT        NOT NULL,
    transportation_to               TEXT        NOT NULL,
    cost_amount                     INTEGER     NOT NULL,
    round_trip                      BOOLEAN     NOT NULL,
    remarks                         TEXT        NOT NULL,
    created_at                      timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at                      timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at                      timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.ts_lesson
(
    lesson_id                        TEXT PRIMARY KEY,
    timesheet_id                     TEXT        NOT NULL,
    flag_on                          BOOLEAN              DEFAULT FALSE,
    timesheet_lesson_hour_created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_lesson_hour_updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_lesson_hour_deleted_at timestamptz,
    staff_id                         TEXT        NOT NULL,
    timesheet_status                 TEXT        NOT NULL,
    timesheet_date                   timestamptz NOT NULL,
    timesheet_remark                 TEXT,
    location_id                      TEXT        NOT NULL,
    timesheet_created_at             timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at             timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at             timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.ts_other_working
(
    other_working_hours_id         TEXT PRIMARY KEY,
    timesheet_id                  TEXT        NOT NULL,
    timesheet_config_id           TEXT        NOT NULL,
    start_time                    timestamptz NOT NULL,
    end_time                      timestamptz NOT NULL,
    total_hour                    SMALLINT    NOT NULL,
    other_working_hour_remarks    TEXT,
    other_working_hour_created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    other_working_hour_updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    other_working_hour_deleted_at timestamptz,
    config_type                   TEXT,
    config_value                  TEXT,
    timesheet_config_created_at   timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_config_updated_at   timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_config_deleted_at   timestamptz,
    staff_id                      TEXT        NOT NULL,
    timesheet_status              TEXT        NOT NULL,
    timesheet_date                timestamptz NOT NULL,
    timesheet_remark              TEXT,
    location_id                   TEXT        NOT NULL,
    timesheet_created_at          timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at          timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at          timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.auto_create_flag_activity_log
(
    auto_create_flag_activity_log_id TEXT PRIMARY KEY,
    staff_id                         TEXT        NOT NULL,
    change_time                      timestamptz NOT NULL,
    flag_on                          BOOLEAN     NOT NULL,
    created_at                       timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at                       timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at                       timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.auto_create_timesheet_flag
(
    staff_id   TEXT PRIMARY KEY,
    flag_on    BOOLEAN     NOT NULL,
    created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.timesheet_confirmation_cut_off_date
(
    timesheet_confirmation_cut_off_date_id TEXT PRIMARY KEY,
    cut_off_date                           INTEGER NOT NULL,
    start_date                             timestamptz NOT NULL,
    end_date                               timestamptz NOT NULL,
    created_at                             timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at                             timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at                             timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.timesheet_confirmation_info
(
    timesheet_confirmation_info_id TEXT PRIMARY KEY,
    period_id                      TEXT        NOT NULL,
    location_id                    TEXT        NOT NULL,
    created_at                     timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at                     timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at                     timestamptz
);

CREATE TABLE IF NOT EXISTS timesheet.timesheet_confirmation_period
(
    timesheet_confirmation_period_id TEXT PRIMARY KEY,
    start_date                       timestamptz NOT NULL,
    end_date                         timestamptz NOT NULL,
    created_at                       timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at                       timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at                       timestamptz
);
