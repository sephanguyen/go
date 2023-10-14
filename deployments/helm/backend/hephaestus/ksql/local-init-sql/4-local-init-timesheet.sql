\connect timesheet;

INSERT INTO public.timesheet_config
(timesheet_config_id, config_type, config_value, created_at, updated_at, deleted_at, resource_path, is_archived)
VALUES('01G6C0AM2FSCXPP7HE7KVQP1QS', 'OTHER_WORKING_HOURS', 'Office', '2022-06-28 23:44:11.940', '2022-06-28 23:44:11.940', NULL, '-2147483644', false) ON CONFLICT DO NOTHING;
INSERT INTO public.timesheet_config
(timesheet_config_id, config_type, config_value, created_at, updated_at, deleted_at, resource_path, is_archived)
VALUES('01G6C0AM2FSCXPP7HE7KVQP1QA', 'OTHER_WORKING_HOURS', 'TA', '2022-06-28 23:44:11.940', '2022-06-28 23:44:11.940', NULL, '-2147483644', false) ON CONFLICT DO NOTHING;
INSERT INTO public.timesheet_config
(timesheet_config_id, config_type, config_value, created_at, updated_at, deleted_at, resource_path, is_archived)
VALUES('01GA81PC0CDZ8V94K75BQ7R8M8', 'OTHER_WORKING_HOURS', 'Other', '2022-08-12 10:31:18.725', '2022-08-12 10:47:17.433', NULL, '-2147483644', false) ON CONFLICT DO NOTHING;



INSERT INTO public."timesheet_confirmation_cut_off_date" 
    (id, cut_off_date, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    ('01GTB0XKND13KM3GDZ49N025HQ', 0, '2021-12-31 15:00:00 +00:00' , '2099-12-30 14:59:59 +00:00', now(), now(), '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public."timesheet_confirmation_cut_off_date" 
    (id, cut_off_date, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    ('01GTB0T5XGFJST7MSVJ2F3AXFP', 0, '2021-12-31 15:00:00 +00:00' , '2099-12-30 14:59:59 +00:00', now(), now(), '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public."timesheet_confirmation_cut_off_date" 
    (id, cut_off_date, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    ('01GTB0Z1J6TCHXHPBRH6Z0WDN0', 0, '2021-12-31 15:00:00 +00:00' , '2099-12-30 14:59:59 +00:00', now(), now(), '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.users (user_id, email, country, name, user_group, updated_at, created_at, resource_path)
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVZF', 'staff@manabie.com', 'COUNTRY_VN', 'Staff name', 'USER_GROUP_TEACHER', '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.staff (staff_id, created_at, updated_at, resource_path)
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVZF', '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO locations (location_id, name, created_at, updated_at, resource_path, location_type)
VALUES ('01FR4M51XJY9E77GSN4QZ1Q9N5', 'End-to-end', '2022-08-01T02:12:53.021148+00:00', '2022-08-01T02:12:53.021148+00:00', '-2147483642', '01FR4M51XJY9E77GSN4QZ1Q9M5')
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet_config
VALUES ('01GA3GV0TRGGSMEMWZW5VMBzzz', 'OTHER_WORKING_HOURS', 'Office', '2022-06-28T16:44:11.94+00:00', '2022-06-28T16:44:11.94+00:00', null, '-2147483642', false)
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet (timesheet_id, staff_id, location_id, timesheet_status, timesheet_date, remark,
                       created_at, updated_at, resource_path)
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVka', '01GA3GV0TRGGSMEMWZW5VMBVZF', '01FR4M51XJY9E77GSN4QZ1Q9N5', 'TIMESHEET_STATUS_APPROVED',
        '2023-02-21 20:00:00', 'remark',
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.other_working_hours(other_working_hours_id, timesheet_id, timesheet_config_id, start_time,
                                end_time, total_hour, remarks, created_at, updated_at, deleted_at, resource_path)
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVka', '01GA3GV0TRGGSMEMWZW5VMBVka', '01GA3GV0TRGGSMEMWZW5VMBzzz', '2020-11-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', 3, 'aa', '2020-11-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', NULL, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.transportation_expense
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVKA', '01GA3GV0TRGGSMEMWZW5VMBVka', 'Train', 'Work', 'Home', 150, false, 'remark',
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', null, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.staff_transportation_expense
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVKA', '01GA3GV0TRGGSMEMWZW5VMBVZF', '01FR4M51XJY9E77GSN4QZ1Q9N5', 'TYPE_TRAIN', 'a',
        'b', 50, true, 'remark', '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', null,
        '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.lessons
(lesson_id, teacher_id, created_at, updated_at, deleted_at, lesson_type, status, stream_learner_counter,
 name, start_time, end_time, resource_path, teaching_model, center_id, teaching_method, teaching_medium,
 scheduling_status, is_locked, scheduler_id)
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVKA', '01GA3GV0TRGGSMEMWZW5VMBVZF', '2020-11-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', null, 'LESSON_TYPE_OFFLINE', 'LESSON_STATUS_DRAFT', 0, '',
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', '-2147483642',
        'LESSON_TEACHING_MODEL_INDIVIDUAL', null, 'LESSON_TEACHING_METHOD_INDIVIDUAL', 'LESSON_TEACHING_MEDIUM_OFFLINE',
        'LESSON_SCHEDULING_STATUS_DRAFT', false, null)
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet_lesson_hours
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVka', '01GA3GV0TRGGSMEMWZW5VMBVKA', '2020-11-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', null, '-2147483642', true)
ON CONFLICT DO NOTHING;

INSERT INTO public.auto_create_flag_activity_log
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVKA', '01GA3GV0TRGGSMEMWZW5VMBVZF', '2020-11-03T07:07:00.511459+00:00', true,
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', null, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.auto_create_timesheet_flag
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVZF', true, '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00',
        null, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet_confirmation_cut_off_date
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVKA', 17, '2020-11-03T07:07:00.511459+00:00', '2099-11-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', null, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet_confirmation_period
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVZF', '2022-11-03T07:07:00.511459+00:00', '2022-12-03T07:07:00.511459+00:00',
        '2020-11-03T07:07:00.511459+00:00', '2020-11-03T07:07:00.511459+00:00', null, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.timesheet_confirmation_info
VALUES ('01GA3GV0TRGGSMEMWZW5VMBVZF', '01GA3GV0TRGGSMEMWZW5VMBVZF', '01FR4M51XJY9E77GSN4QZ1Q9N5',
        '2022-12-03T07:07:00.511459+00:00', '2022-12-03T07:07:00.511459+00:00', null, '-2147483642')
ON CONFLICT DO NOTHING;
