CREATE TABLE IF NOT EXISTS public.ts_transportation
(
    transportation_expense_id         TEXT PRIMARY KEY,
    timesheet_id                      TEXT       ,
    transportation_type               TEXT       ,
    transportation_from               TEXT       ,
    transportation_to                 TEXT       ,
    cost_amount                       INTEGER    ,
    round_trip                        BOOLEAN    ,
    transportation_expense_remarks    TEXT       ,
    transportation_expense_created_at timestamptz DEFAULT timezone('utc'::text, now()),
    transportation_expense_updated_at timestamptz DEFAULT timezone('utc'::text, now()),
    transportation_expense_deleted_at timestamptz,
    timesheet_status                  TEXT       ,
    timesheet_date                    timestamptz,
    timesheet_remark                  TEXT       ,
    timesheet_created_at              timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at              timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at              timestamptz,
    staff_id                          TEXT       
);

CREATE TABLE IF NOT EXISTS public.staff_transportation_expense
(
    staff_transportation_expense_id TEXT PRIMARY KEY,
    staff_id                        TEXT       ,
    location_id                     TEXT       ,
    transport_type                  TEXT       ,
    transportation_from             TEXT       ,
    transportation_to               TEXT       ,
    cost_amount                     INTEGER    ,
    round_trip                      BOOLEAN    ,
    remarks                         TEXT       ,
    created_at                      timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at                      timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at                      timestamptz
);

CREATE TABLE IF NOT EXISTS public.ts_lesson
(
    lesson_id                        TEXT PRIMARY KEY,
    timesheet_id                     TEXT       ,
    flag_on                          BOOLEAN              DEFAULT FALSE,
    timesheet_lesson_hour_created_at timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_lesson_hour_updated_at timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_lesson_hour_deleted_at timestamptz,
    staff_id                         TEXT       ,
    timesheet_status                 TEXT       ,
    timesheet_date                   timestamptz,
    timesheet_remark                 TEXT,
    location_id                      TEXT       ,
    timesheet_created_at             timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at             timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at             timestamptz
);

CREATE TABLE IF NOT EXISTS public.ts_other_working
(
    other_working_hours_id         TEXT PRIMARY KEY,
    timesheet_id                  TEXT       ,
    timesheet_config_id           TEXT       ,
    start_time                    timestamptz,
    end_time                      timestamptz,
    total_hour                    SMALLINT   ,
    other_working_hour_remarks    TEXT,
    other_working_hour_created_at timestamptz DEFAULT timezone('utc'::text, now()),
    other_working_hour_updated_at timestamptz DEFAULT timezone('utc'::text, now()),
    other_working_hour_deleted_at timestamptz,
    config_type                   TEXT,
    config_value                  TEXT,
    timesheet_config_created_at   timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_config_updated_at   timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_config_deleted_at   timestamptz,
    staff_id                      TEXT       ,
    timesheet_status              TEXT       ,
    timesheet_date                timestamptz,
    timesheet_remark              TEXT,
    location_id                   TEXT       ,
    timesheet_created_at          timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_updated_at          timestamptz DEFAULT timezone('utc'::text, now()),
    timesheet_deleted_at          timestamptz
);

CREATE TABLE IF NOT EXISTS public.auto_create_flag_activity_log
(
    auto_create_flag_activity_log_id TEXT PRIMARY KEY,
    staff_id                         TEXT       ,
    change_time                      timestamptz,
    flag_on                          BOOLEAN    ,
    created_at                       timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at                       timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at                       timestamptz
);

CREATE TABLE IF NOT EXISTS public.auto_create_timesheet_flag
(
    staff_id   TEXT PRIMARY KEY,
    flag_on    BOOLEAN    ,
    created_at timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.timesheet_confirmation_cut_off_date
(
    timesheet_confirmation_cut_off_date_id TEXT PRIMARY KEY,
    cut_off_date                           INTEGER,
    start_date                             timestamptz,
    end_date                               timestamptz,
    created_at                             timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at                             timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at                             timestamptz
);

CREATE TABLE IF NOT EXISTS public.timesheet_confirmation_info
(
    timesheet_confirmation_info_id TEXT PRIMARY KEY,
    period_id                      TEXT       ,
    location_id                    TEXT       ,
    created_at                     timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at                     timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at                     timestamptz
);

CREATE TABLE IF NOT EXISTS public.timesheet_confirmation_period
(
    timesheet_confirmation_period_id TEXT PRIMARY KEY,
    start_date                       timestamptz,
    end_date                         timestamptz,
    created_at                       timestamptz DEFAULT timezone('utc'::text, now()),
    updated_at                       timestamptz DEFAULT timezone('utc'::text, now()),
    deleted_at                       timestamptz
);

ALTER PUBLICATION kec_publication SET TABLE 
public.ts_transportation,
public.staff_transportation_expense,
public.ts_lesson,
public.ts_other_working,
public.auto_create_flag_activity_log,
public.auto_create_timesheet_flag,
public.timesheet_confirmation_cut_off_date,
public.timesheet_confirmation_info,
public.timesheet_confirmation_period
;
