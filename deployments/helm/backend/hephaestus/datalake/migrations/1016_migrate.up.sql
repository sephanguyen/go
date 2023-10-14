create schema if not exists timesheet;

create table if not exists timesheet.timesheet
(
    timesheet_id     text primary key,
    staff_id         text                     not null,
    location_id      text                     not null,
    timesheet_status text                     not null,
    timesheet_date   timestamp with time zone not null,
    remark           text,
    resource_path    text,
    created_at       timestamp with time zone not null,
    updated_at       timestamp with time zone not null,
    deleted_at       timestamp with time zone
);

create table if not exists timesheet.other_working_hours
(
    other_working_hours_id text primary key,
    timesheet_id           text                     not null,
    timesheet_config_id    text                     not null,
    start_time             timestamp with time zone not null,
    end_time               timestamp with time zone not null,
    total_hour             smallint                 not null,
    remarks                text,
    created_at             timestamp with time zone not null,
    updated_at             timestamp with time zone not null,
    deleted_at             timestamp with time zone,
    resource_path          text
);

create table if not exists timesheet.transportation_expense
(
    transportation_expense_id text primary key,
    timesheet_id              text                     not null,
    transportation_type       text                     not null,
    transportation_from       text                     not null,
    transportation_to         text                     not null,
    cost_amount               integer,
    round_trip                boolean                  not null,
    remarks                   text                     not null,
    created_at                timestamp with time zone not null,
    updated_at                timestamp with time zone not null,
    deleted_at                timestamp with time zone,
    resource_path             text
);

create table if not exists timesheet.staff_transportation_expense
(
    id                  text primary key,
    staff_id            text                     not null,
    location_id         text                     not null,
    transportation_type text                     not null,
    transportation_from text                     not null,
    transportation_to   text                     not null,
    cost_amount         integer default 0,
    round_trip          boolean                  not null,
    remarks             text                     not null,
    created_at          timestamp with time zone not null,
    updated_at          timestamp with time zone not null,
    deleted_at          timestamp with time zone,
    resource_path       text
);

create table if not exists timesheet.lessons
(
    lesson_id              text primary key,
    teacher_id             text,
    course_id              text,
    created_at             timestamp with time zone       not null,
    updated_at             timestamp with time zone       not null,
    deleted_at             timestamp with time zone,
    end_at                 timestamp with time zone,
    control_settings       jsonb,
    lesson_group_id        text,
    room_id                text,
    lesson_type            text,
    status                 text,
    stream_learner_counter integer default 0              not null,
    learner_ids            text[]  default '{}' :: text[] not null,
    name                   text,
    start_time             timestamp with time zone,
    end_time               timestamp with time zone,
    resource_path          text,
    room_state             jsonb,
    teaching_model         text,
    class_id               text,
    center_id              text,
    teaching_method        text,
    teaching_medium        text,
    scheduling_status      text    default 'LESSON_SCHEDULING_STATUS_PUBLISHED' :: text,
    is_locked              boolean default false          not null,
    scheduler_id           text
);

create table if not exists timesheet.timesheet_lesson_hours
(
    timesheet_id  text                                                            not null,
    lesson_id     text                                                            not null,
    created_at    timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at    timestamp with time zone                                        not null,
    deleted_at    timestamp with time zone,
    resource_path text,
    flag_on       boolean                  default false                          not null,
    constraint timesheet_lesson_hours_pk
        primary key (timesheet_id, lesson_id)
);

create table if not exists timesheet.auto_create_flag_activity_log
(
    id            text primary key,
    staff_id      text                     not null,
    change_time   timestamp with time zone not null,
    flag_on       boolean                  not null,
    created_at    timestamp with time zone not null,
    updated_at    timestamp with time zone not null,
    deleted_at    timestamp with time zone,
    resource_path text
);

create table if not exists timesheet.auto_create_timesheet_flag
(
    staff_id      text primary key,
    flag_on       boolean                  default false                          not null,
    created_at    timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at    timestamp with time zone                                        not null,
    deleted_at    timestamp with time zone,
    resource_path text
);

create table if not exists timesheet.timesheet_confirmation_cut_off_date
(
    id            text primary key,
    cut_off_date  integer                                                         not null,
    start_date    timestamp with time zone                                        not null,
    end_date      timestamp with time zone,
    created_at    timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at    timestamp with time zone                                        not null,
    deleted_at    timestamp with time zone,
    resource_path text
);

create table if not exists timesheet.timesheet_confirmation_info
(
    id            text primary key,
    period_id     text                                                            not null,
    location_id   text                                                            not null,
    created_at    timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at    timestamp with time zone                                        not null,
    deleted_at    timestamp with time zone,
    resource_path text
);

create table if not exists timesheet.timesheet_confirmation_period
(
    id            text primary key,
    start_date    timestamp with time zone                                        not null,
    end_date      timestamp with time zone                                        not null,
    created_at    timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at    timestamp with time zone                                        not null,
    deleted_at    timestamp with time zone,
    resource_path text
);

create table if not exists timesheet.timesheet_config
(
    timesheet_config_id text primary key,
    config_type         text                                                            not null,
    config_value        text                                                            not null,
    created_at          timestamp with time zone default timezone('utc' :: text, now()) not null,
    updated_at          timestamp with time zone default timezone('utc' :: text, now()) not null,
    deleted_at          timestamp with time zone,
    resource_path       text,
    is_archived         boolean                  default false                          not null
);

ALTER PUBLICATION publication_for_datawarehouse ADD TABLE
    timesheet.timesheet,
    timesheet.other_working_hours,
    timesheet.transportation_expense,
    timesheet.staff_transportation_expense,
    timesheet.lessons,
    timesheet.timesheet_lesson_hours,
    timesheet.auto_create_flag_activity_log,
    timesheet.auto_create_timesheet_flag,
    timesheet.timesheet_confirmation_cut_off_date,
    timesheet.timesheet_confirmation_info,
    timesheet.timesheet_confirmation_period,
    timesheet.timesheet_config;
