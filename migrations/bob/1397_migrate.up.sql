--- Add permission ---
INSERT INTO permission
(permission_id, permission_name, created_at, updated_at, resource_path)
VALUES
    ('01GJZGF0VZDYQSHG5P8GSEH53Y', 'user.user.read', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8HNSJSFX', 'user.user.write', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8A5ENZRR', 'user.user.read', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P8C7B3TGF', 'user.user.write', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P7KX1MHRV', 'user.student_enrollment_status_history.read', now(), now(), '-2147483629'),
    ('01GJZGF0VZDYQSHG5P7XD2NXCQ', 'user.student_enrollment_status_history.read', now(), now(), '-2147483630'),
    ('01GJZGF0VZDYQSHG5P8304Y5FM', 'user.student_enrollment_status_history.read', now(), now(), '-2147483631'),
    ('01GJZGF0VZDYQSHG5P89MT578D', 'user.student_enrollment_status_history.read', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P8ER0QPA9', 'user.student_enrollment_status_history.read', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8J2X450T', 'user.student_enrollment_status_history.read', now(), now(), '-2147483634'),
    ('01GJZGF0W0YCW6TWPQ5A51BV37', 'user.student_enrollment_status_history.read', now(), now(), '-2147483635'),
    ('01GJZGF0W0YCW6TWPQ5G71GK9S', 'user.student_enrollment_status_history.read', now(), now(), '-2147483637'),
    ('01GJZGF0W0YCW6TWPQ5JW624EN', 'user.student_enrollment_status_history.read', now(), now(), '-2147483638'),
    ('01GJZGF0W0YCW6TWPQ5VG96RWR', 'user.student_enrollment_status_history.read', now(), now(), '-2147483639'),
    ('01GJZGF0W0YCW6TWPQ626TYMR0', 'user.student_enrollment_status_history.read', now(), now(), '-2147483640'),
    ('01GJZGF0W0YCW6TWPQ6A05BZWT', 'user.student_enrollment_status_history.read', now(), now(), '-2147483641'),
    ('01GJZGF0W0YCW6TWPQ6JMZ6G95', 'user.student_enrollment_status_history.read', now(), now(), '-2147483642'),
    ('01GJZGF0W0YCW6TWPQ6S3XWBKS', 'user.student_enrollment_status_history.read', now(), now(), '-2147483643'),
    ('01GJZGF0W0YCW6TWPQ717QQCP1', 'user.student_enrollment_status_history.read', now(), now(), '-2147483644'),
    ('01GJZGF0W0YCW6TWPQ79E0H07T', 'user.student_enrollment_status_history.read', now(), now(), '-2147483645'),
    ('01GJZGF0W0YCW6TWPQ7KC1WR6E', 'user.student_enrollment_status_history.read', now(), now(), '-2147483646'),
    ('01GJZGF0W0YCW6TWPQ7W6D1FFV', 'user.student_enrollment_status_history.read', now(), now(), '-2147483647'),
    ('01GJZGF0W0YCW6TWPQ85FBBEMY', 'user.student_enrollment_status_history.read', now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Add role ---
INSERT INTO role
(role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01GJZNHK0NJSK9Z6ASGSPXRX2R', 'UsermgmtScheduleJob', true, now(), now(), '-2147483629'),
    ('01GJZNHK0NJSK9Z6ASGWXHXS7X', 'UsermgmtScheduleJob', true, now(), now(), '-2147483630'),
    ('01GJZNHK0NJSK9Z6ASGYW5J1XY', 'UsermgmtScheduleJob', true, now(), now(), '-2147483631'),
    ('01GJZNHK0NJSK9Z6ASH12X2GV7', 'UsermgmtScheduleJob', true, now(), now(), '-2147483632'),
    ('01GJZNHK0NJSK9Z6ASH24NT9RC', 'UsermgmtScheduleJob', true, now(), now(), '-2147483633'),
    ('01GJZNHK0NJSK9Z6ASH2KHZ8DW', 'UsermgmtScheduleJob', true, now(), now(), '-2147483634'),
    ('01GJZNHK0NJSK9Z6ASH2QDJBRH', 'UsermgmtScheduleJob', true, now(), now(), '-2147483635'),
    ('01GJZNHK0NJSK9Z6ASH32FJNCX', 'UsermgmtScheduleJob', true, now(), now(), '-2147483637'),
    ('01GJZNHK0NJSK9Z6ASH6YZA563', 'UsermgmtScheduleJob', true, now(), now(), '-2147483638'),
    ('01GJZNHK0NJSK9Z6ASHA4PDMAZ', 'UsermgmtScheduleJob', true, now(), now(), '-2147483639'),
    ('01GJZNHK0NJSK9Z6ASHBQA46N8', 'UsermgmtScheduleJob', true, now(), now(), '-2147483640'),
    ('01GJZNHK0NJSK9Z6ASHCRSPQDY', 'UsermgmtScheduleJob', true, now(), now(), '-2147483641'),
    ('01GJZNHK0NJSK9Z6ASHET0PTPK', 'UsermgmtScheduleJob', true, now(), now(), '-2147483642'),
    ('01GJZNHK0NJSK9Z6ASHFWF35Z1', 'UsermgmtScheduleJob', true, now(), now(), '-2147483643'),
    ('01GJZNHK0NJSK9Z6ASHH8F5WN8', 'UsermgmtScheduleJob', true, now(), now(), '-2147483644'),
    ('01GJZNHK0NJSK9Z6ASHJ5CMFN9', 'UsermgmtScheduleJob', true, now(), now(), '-2147483645'),
    ('01GJZNHK0NJSK9Z6ASHJMCMR1Q', 'UsermgmtScheduleJob', true, now(), now(), '-2147483646'),
    ('01GJZNHK0NJSK9Z6ASHPM24AAB', 'UsermgmtScheduleJob', true, now(), now(), '-2147483647'),
    ('01GJZNHK0NJSK9Z6ASHPWBS324', 'UsermgmtScheduleJob', true, now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Add permission_role ---
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GJZGF0VZDYQSHG5P7KX1MHRV', '01GJZNHK0NJSK9Z6ASGSPXRX2R', now(), now(), '-2147483629'),
    ('01GGVXWS2AMVMCXWJWV1XEFHDR', '01GJZNHK0NJSK9Z6ASGSPXRX2R', now(), now(), '-2147483629'),
    ('01GGVXWS2E95PSQ028HBDRSDXQ', '01GJZNHK0NJSK9Z6ASGSPXRX2R', now(), now(), '-2147483629'),
    ('01GJZGF0VZDYQSHG5P7XD2NXCQ', '01GJZNHK0NJSK9Z6ASGWXHXS7X', now(), now(), '-2147483630'),
    ('01GGVRH1PF6TV99ZNHH36D9W5N', '01GJZNHK0NJSK9Z6ASGWXHXS7X', now(), now(), '-2147483630'),
    ('01GGVRH1P9HZQ7T4EX424B1BCA', '01GJZNHK0NJSK9Z6ASGWXHXS7X', now(), now(), '-2147483630'),
    ('01GJZGF0VZDYQSHG5P8304Y5FM', '01GJZNHK0NJSK9Z6ASGYW5J1XY', now(), now(), '-2147483631'),
    ('01GGVJB2CKXCPM0W1SR5PJ0VMX', '01GJZNHK0NJSK9Z6ASGYW5J1XY', now(), now(), '-2147483631'),
    ('01GGVJB2CJZ9JEN4P309WJ1PSA', '01GJZNHK0NJSK9Z6ASGYW5J1XY', now(), now(), '-2147483631'),
    ('01GJZGF0VZDYQSHG5P89MT578D', '01GJZNHK0NJSK9Z6ASH12X2GV7', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P8A5ENZRR', '01GJZNHK0NJSK9Z6ASH12X2GV7', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P8C7B3TGF', '01GJZNHK0NJSK9Z6ASH12X2GV7', now(), now(), '-2147483632'),
    ('01GJZGF0VZDYQSHG5P8ER0QPA9', '01GJZNHK0NJSK9Z6ASH24NT9RC', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8GSEH53Y', '01GJZNHK0NJSK9Z6ASH24NT9RC', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8HNSJSFX', '01GJZNHK0NJSK9Z6ASH24NT9RC', now(), now(), '-2147483633'),
    ('01GJZGF0VZDYQSHG5P8J2X450T', '01GJZNHK0NJSK9Z6ASH2KHZ8DW', now(), now(), '-2147483634'),
    ('01GG9EWWVNBHGDM01WQERA8MX0', '01GJZNHK0NJSK9Z6ASH2KHZ8DW', now(), now(), '-2147483634'),
    ('01GG9EWWVNBHGDM01WQERA8MX1', '01GJZNHK0NJSK9Z6ASH2KHZ8DW', now(), now(), '-2147483634'),
    ('01GJZGF0W0YCW6TWPQ5A51BV37', '01GJZNHK0NJSK9Z6ASH2QDJBRH', now(), now(), '-2147483635'),
    ('01GG9EWWVNBHGDM01WQERA8MX2', '01GJZNHK0NJSK9Z6ASH2QDJBRH', now(), now(), '-2147483635'),
    ('01GG9EWWVNBHGDM01WQERA8MX3', '01GJZNHK0NJSK9Z6ASH2QDJBRH', now(), now(), '-2147483635'),
    ('01GJZGF0W0YCW6TWPQ5G71GK9S', '01GJZNHK0NJSK9Z6ASH32FJNCX', now(), now(), '-2147483637'),
    ('01GG9EWWVNBHGDM01WQERA8MX4', '01GJZNHK0NJSK9Z6ASH32FJNCX', now(), now(), '-2147483637'),
    ('01GG9EWWVNBHGDM01WQERA8MX5', '01GJZNHK0NJSK9Z6ASH32FJNCX', now(), now(), '-2147483637'),
    ('01GJZGF0W0YCW6TWPQ5JW624EN', '01GJZNHK0NJSK9Z6ASH6YZA563', now(), now(), '-2147483638'),
    ('01GG9EWWVNBHGDM01WQERA8MX6', '01GJZNHK0NJSK9Z6ASH6YZA563', now(), now(), '-2147483638'),
    ('01GG9EWWVNBHGDM01WQERA8MX7', '01GJZNHK0NJSK9Z6ASH6YZA563', now(), now(), '-2147483638'),
    ('01GJZGF0W0YCW6TWPQ5VG96RWR', '01GJZNHK0NJSK9Z6ASHA4PDMAZ', now(), now(), '-2147483639'),
    ('01GG9EWWVNBHGDM01WQERA8MX8', '01GJZNHK0NJSK9Z6ASHA4PDMAZ', now(), now(), '-2147483639'),
    ('01GG9EWWVNBHGDM01WQERA8MX9', '01GJZNHK0NJSK9Z6ASHA4PDMAZ', now(), now(), '-2147483639'),
    ('01GJZGF0W0YCW6TWPQ626TYMR0', '01GJZNHK0NJSK9Z6ASHBQA46N8', now(), now(), '-2147483640'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH8', '01GJZNHK0NJSK9Z6ASHBQA46N8', now(), now(), '-2147483640'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH9', '01GJZNHK0NJSK9Z6ASHBQA46N8', now(), now(), '-2147483640'),
    ('01GJZGF0W0YCW6TWPQ6A05BZWT', '01GJZNHK0NJSK9Z6ASHCRSPQDY', now(), now(), '-2147483641'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH6', '01GJZNHK0NJSK9Z6ASHCRSPQDY', now(), now(), '-2147483641'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH7', '01GJZNHK0NJSK9Z6ASHCRSPQDY', now(), now(), '-2147483641'),
    ('01GJZGF0W0YCW6TWPQ6JMZ6G95', '01GJZNHK0NJSK9Z6ASHET0PTPK', now(), now(), '-2147483642'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH4', '01GJZNHK0NJSK9Z6ASHET0PTPK', now(), now(), '-2147483642'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH5', '01GJZNHK0NJSK9Z6ASHET0PTPK', now(), now(), '-2147483642'),
    ('01GJZGF0W0YCW6TWPQ6S3XWBKS', '01GJZNHK0NJSK9Z6ASHFWF35Z1', now(), now(), '-2147483643'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH0', '01GJZNHK0NJSK9Z6ASHFWF35Z1', now(), now(), '-2147483643'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH3', '01GJZNHK0NJSK9Z6ASHFWF35Z1', now(), now(), '-2147483643'),
    ('01GJZGF0W0YCW6TWPQ717QQCP1', '01GJZNHK0NJSK9Z6ASHH8F5WN8', now(), now(), '-2147483644'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH1', '01GJZNHK0NJSK9Z6ASHH8F5WN8', now(), now(), '-2147483644'),
    ('01GDMTJ3NB0TKRWS7ZRMHFTHH2', '01GJZNHK0NJSK9Z6ASHH8F5WN8', now(), now(), '-2147483644'),
    ('01GJZGF0W0YCW6TWPQ79E0H07T', '01GJZNHK0NJSK9Z6ASHJ5CMFN9', now(), now(), '-2147483645'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B0', '01GJZNHK0NJSK9Z6ASHJ5CMFN9', now(), now(), '-2147483645'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B1', '01GJZNHK0NJSK9Z6ASHJ5CMFN9', now(), now(), '-2147483645'),
    ('01GJZGF0W0YCW6TWPQ7KC1WR6E', '01GJZNHK0NJSK9Z6ASHJMCMR1Q', now(), now(), '-2147483646'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B2', '01GJZNHK0NJSK9Z6ASHJMCMR1Q', now(), now(), '-2147483646'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B3', '01GJZNHK0NJSK9Z6ASHJMCMR1Q', now(), now(), '-2147483646'),
    ('01GJZGF0W0YCW6TWPQ7W6D1FFV', '01GJZNHK0NJSK9Z6ASHPM24AAB', now(), now(), '-2147483647'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B4', '01GJZNHK0NJSK9Z6ASHPM24AAB', now(), now(), '-2147483647'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B5', '01GJZNHK0NJSK9Z6ASHPM24AAB', now(), now(), '-2147483647'),
    ('01GJZGF0W0YCW6TWPQ85FBBEMY', '01GJZNHK0NJSK9Z6ASHPWBS324', now(), now(), '-2147483648'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B6', '01GJZNHK0NJSK9Z6ASHPWBS324', now(), now(), '-2147483648'),
    ('01GG9EWWVPFT4HTY1E7RF9F4B7', '01GJZNHK0NJSK9Z6ASHPWBS324', now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Add User Group ---
INSERT INTO public.user_group
(user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01GJZQ8C9TCT9H3VWAACGYE8YA', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483629'),
    ('01GJZQ8C9TCT9H3VWAAFC4B05V', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483630'),
    ('01GJZQ8C9TCT9H3VWAAFT4JYDK', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483631'),
    ('01GJZQ8C9TCT9H3VWAAK4VCS6K', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483632'),
    ('01GJZQ8C9TCT9H3VWAANCZJDWV', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483633'),
    ('01GJZQ8C9TCT9H3VWAAR5ZXEQ2', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483634'),
    ('01GJZQ8C9TCT9H3VWAAS3XCJ2F', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483635'),
    ('01GJZQ8C9TCT9H3VWAAWB9REXE', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483637'),
    ('01GJZQ8C9TCT9H3VWAB03X8JP1', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483638'),
    ('01GJZQ8C9TCT9H3VWAB346EBMR', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483639'),
    ('01GJZQ8C9TCT9H3VWAB42HE3FA', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483640'),
    ('01GJZQ8C9TCT9H3VWAB4PNWD3V', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483641'),
    ('01GJZQ8C9TCT9H3VWAB4YQG240', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483642'),
    ('01GJZQ8C9TCT9H3VWAB55C8QG3', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483643'),
    ('01GJZQ8C9TCT9H3VWAB5X6AJJT', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483644'),
    ('01GJZQ8C9TCT9H3VWAB9CW93S7', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483645'),
    ('01GJZQ8C9TCT9H3VWABAKDZG6Q', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483646'),
    ('01GJZQ8C9TCT9H3VWABDBRCRTD', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483647'),
    ('01GJZQ8C9TCT9H3VWABFGP75A1', 'UserGroup UsermgmtScheduleJob', true, now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Grant role to User group ---
INSERT INTO public.granted_role
(granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GJZQA9FV203R493A5ZX110Y7', '01GJZQ8C9TCT9H3VWAACGYE8YA', '01GJZNHK0NJSK9Z6ASGSPXRX2R', now(), now(), '-2147483629'),
    ('01GJZQA9FV203R493A62N76F2N', '01GJZQ8C9TCT9H3VWAAFC4B05V', '01GJZNHK0NJSK9Z6ASGWXHXS7X', now(), now(), '-2147483630'),
    ('01GJZQA9FV203R493A66BJ3PP0', '01GJZQ8C9TCT9H3VWAAFT4JYDK', '01GJZNHK0NJSK9Z6ASGYW5J1XY', now(), now(), '-2147483631'),
    ('01GJZQA9FV203R493A66S2FSES', '01GJZQ8C9TCT9H3VWAAK4VCS6K', '01GJZNHK0NJSK9Z6ASH12X2GV7', now(), now(), '-2147483632'),
    ('01GJZQA9FV203R493A68YY24P0', '01GJZQ8C9TCT9H3VWAANCZJDWV', '01GJZNHK0NJSK9Z6ASH24NT9RC', now(), now(), '-2147483633'),
    ('01GJZQA9FV203R493A6C83YEJC', '01GJZQ8C9TCT9H3VWAAR5ZXEQ2', '01GJZNHK0NJSK9Z6ASH2KHZ8DW', now(), now(), '-2147483634'),
    ('01GJZQA9FV203R493A6E0011RT', '01GJZQ8C9TCT9H3VWAAS3XCJ2F', '01GJZNHK0NJSK9Z6ASH2QDJBRH', now(), now(), '-2147483635'),
    ('01GJZQA9FV203R493A6GJPCXHJ', '01GJZQ8C9TCT9H3VWAAWB9REXE', '01GJZNHK0NJSK9Z6ASH32FJNCX', now(), now(), '-2147483637'),
    ('01GJZQA9FV203R493A6JZ905RS', '01GJZQ8C9TCT9H3VWAB03X8JP1', '01GJZNHK0NJSK9Z6ASH6YZA563', now(), now(), '-2147483638'),
    ('01GJZQA9FV203R493A6M7DZEDY', '01GJZQ8C9TCT9H3VWAB346EBMR', '01GJZNHK0NJSK9Z6ASHA4PDMAZ', now(), now(), '-2147483639'),
    ('01GJZQA9FV203R493A6QEJZ662', '01GJZQ8C9TCT9H3VWAB42HE3FA', '01GJZNHK0NJSK9Z6ASHBQA46N8', now(), now(), '-2147483640'),
    ('01GJZQA9FV203R493A6SY6T3ET', '01GJZQ8C9TCT9H3VWAB4PNWD3V', '01GJZNHK0NJSK9Z6ASHCRSPQDY', now(), now(), '-2147483641'),
    ('01GJZQA9FV203R493A6V2CQKTS', '01GJZQ8C9TCT9H3VWAB4YQG240', '01GJZNHK0NJSK9Z6ASHET0PTPK', now(), now(), '-2147483642'),
    ('01GJZQA9FV203R493A6W183GDR', '01GJZQ8C9TCT9H3VWAB55C8QG3', '01GJZNHK0NJSK9Z6ASHFWF35Z1', now(), now(), '-2147483643'),
    ('01GJZQA9FWDF96J4HY8C2PD24T', '01GJZQ8C9TCT9H3VWAB5X6AJJT', '01GJZNHK0NJSK9Z6ASHH8F5WN8', now(), now(), '-2147483644'),
    ('01GJZQA9FWDF96J4HY8F11PCYT', '01GJZQ8C9TCT9H3VWAB9CW93S7', '01GJZNHK0NJSK9Z6ASHJ5CMFN9', now(), now(), '-2147483645'),
    ('01GJZQA9FWDF96J4HY8JXRFYSQ', '01GJZQ8C9TCT9H3VWABAKDZG6Q', '01GJZNHK0NJSK9Z6ASHJMCMR1Q', now(), now(), '-2147483646'),
    ('01GJZQA9FWDF96J4HY8M4ABGSB', '01GJZQ8C9TCT9H3VWABDBRCRTD', '01GJZNHK0NJSK9Z6ASHPM24AAB', now(), now(), '-2147483647'),
    ('01GJZQA9FWDF96J4HY8Q3RGKYM', '01GJZQ8C9TCT9H3VWABFGP75A1', '01GJZNHK0NJSK9Z6ASHPWBS324', now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Grant location to a role ---
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01GJZQA9FV203R493A5ZX110Y7', '01GFMNHQ1WHGRC8AW6K913AM3G', now(), now(), '-2147483629'),
    ('01GJZQA9FV203R493A62N76F2N', '01GFMMFRXC6SKTTT44HWR3BRY8', now(), now(), '-2147483630'),
    ('01GJZQA9FV203R493A66BJ3PP0', '01GDWSMJS6APH4SX2NP5NFWHG5', now(), now(), '-2147483631'),
    ('01GJZQA9FV203R493A6C83YEJC', '01FR4M51XJY9E77GSN4QZ1Q8N5', now(), now(), '-2147483634'),
    ('01GJZQA9FV203R493A6E0011RT', '01FR4M51XJY9E77GSN4QZ1Q8N4', now(), now(), '-2147483635'),
    ('01GJZQA9FV203R493A6GJPCXHJ', '01FR4M51XJY9E77GSN4QZ1Q8N3', now(), now(), '-2147483637'),
    ('01GJZQA9FV203R493A6JZ905RS', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638'),
    ('01GJZQA9FV203R493A6M7DZEDY', '01FR4M51XJY9E77GSN4QZ1Q8N1', now(), now(), '-2147483639'),
    ('01GJZQA9FV203R493A6QEJZ662', '01FR4M51XJY9E77GSN4QZ1Q9N9', now(), now(), '-2147483640'),
    ('01GJZQA9FV203R493A6SY6T3ET', '01FR4M51XJY9E77GSN4QZ1Q9N8', now(), now(), '-2147483641'),
    ('01GJZQA9FV203R493A6V2CQKTS', '01FR4M51XJY9E77GSN4QZ1Q9N7', now(), now(), '-2147483642'),
    ('01GJZQA9FV203R493A6W183GDR', '01FR4M51XJY9E77GSN4QZ1Q9N6', now(), now(), '-2147483643'),
    ('01GJZQA9FWDF96J4HY8C2PD24T', '01FR4M51XJY9E77GSN4QZ1Q9N5', now(), now(), '-2147483644'),
    ('01GJZQA9FWDF96J4HY8F11PCYT', '01FR4M51XJY9E77GSN4QZ1Q9N4', now(), now(), '-2147483645'),
    ('01GJZQA9FWDF96J4HY8JXRFYSQ', '01FR4M51XJY9E77GSN4QZ1Q9N3', now(), now(), '-2147483646'),
    ('01GJZQA9FWDF96J4HY8M4ABGSB', '01FR4M51XJY9E77GSN4QZ1Q9N2', now(), now(), '-2147483647'),
    ('01GJZQA9FWDF96J4HY8Q3RGKYM', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

--- Upsert granted_permission ---
INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAACGYE8YA')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAFC4B05V')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAFT4JYDK')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAK4VCS6K')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAANCZJDWV')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAR5ZXEQ2')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAS3XCJ2F')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAWB9REXE')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB03X8JP1')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB346EBMR')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB42HE3FA')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB4PNWD3V')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB4YQG240')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB55C8QG3')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB5X6AJJT')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAB9CW93S7')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWABAKDZG6Q')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWABDBRCRTD')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWABFGP75A1')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;
