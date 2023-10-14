INSERT INTO
    role (
        role_id,
        role_name,
        is_system,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ3PBYBPZM',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483629'
    ), (
        '01GSX7KMWVTMH0E8NZ3RS5PV7M',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483630'
    ), (
        '01GSX7KMWVTMH0E8NZ3SCJQEY1',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483631'
    ), (
        '01GSX7KMWVTMH0E8NZ3T6Q5038',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483634'
    ), (
        '01GSX7KMWVTMH0E8NZ3VSVJH4A',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483635'
    ), (
        '01GSX7KMWVTMH0E8NZ3XMG764J',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483637'
    ), (
        '01GSX7KMWVTMH0E8NZ3ZV9J5XG',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483638'
    ), (
        '01GSX7KMWVTMH0E8NZ41AGQE94',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483639'
    ), (
        '01GSX7KMWVTMH0E8NZ42AGGVSB',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483640'
    ), (
        '01GSX7KMWVTMH0E8NZ45ZE0ZAH',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483641'
    ), (
        '01GSX7KMWVTMH0E8NZ49247D1H',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483642'
    ), (
        '01GSX7KMWVTMH0E8NZ4BVCBB3Z',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483643'
    ), (
        '01GSX7KMWVTMH0E8NZ4DRAJ1R9',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483644'
    ), (
        '01GSX7KMWVTMH0E8NZ4GXDD0A9',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483645'
    ), (
        '01GSX7KMWVTMH0E8NZ4HTHA56Q',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483646'
    ), (
        '01GSX7KMWVTMH0E8NZ4JDVNX69',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483647'
    ), (
        '01GSX7KMWVTMH0E8NZ4NK34A2E',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    permission_role (
        permission_id,
        role_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GGVXWS1HVVKD8EZMYXASK39H',
        '01GSX7KMWVTMH0E8NZ3PBYBPZM',
        now(),
        now(),
        '-2147483629'
    ), (
        '01GGVXWS2AMVMCXWJWV1XEFHDR',
        '01GSX7KMWVTMH0E8NZ3PBYBPZM',
        now(),
        now(),
        '-2147483629'
    ), (
        '01GGVRH1NEDXT5V5Q37S2HE6TW',
        '01GSX7KMWVTMH0E8NZ3RS5PV7M',
        now(),
        now(),
        '-2147483630'
    ), (
        '01GGVRH1P9HZQ7T4EX424B1BCA',
        '01GSX7KMWVTMH0E8NZ3RS5PV7M',
        now(),
        now(),
        '-2147483630'
    ), (
        '01GGVJB2CAMY0TS1XXWA0MERH0',
        '01GSX7KMWVTMH0E8NZ3SCJQEY1',
        now(),
        now(),
        '-2147483631'
    ), (
        '01GGVJB2CJZ9JEN4P309WJ1PSA',
        '01GSX7KMWVTMH0E8NZ3SCJQEY1',
        now(),
        now(),
        '-2147483631'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YB2',
        '01GSX7KMWVTMH0E8NZ3T6Q5038',
        now(),
        now(),
        '-2147483634'
    ), (
        '01GG9EWWVNBHGDM01WQERA8MX0',
        '01GSX7KMWVTMH0E8NZ3T6Q5038',
        now(),
        now(),
        '-2147483634'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YB1',
        '01GSX7KMWVTMH0E8NZ3VSVJH4A',
        now(),
        now(),
        '-2147483635'
    ), (
        '01GG9EWWVNBHGDM01WQERA8MX2',
        '01GSX7KMWVTMH0E8NZ3VSVJH4A',
        now(),
        now(),
        '-2147483635'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA9',
        '01GSX7KMWVTMH0E8NZ3XMG764J',
        now(),
        now(),
        '-2147483637'
    ), (
        '01GG9EWWVNBHGDM01WQERA8MX4',
        '01GSX7KMWVTMH0E8NZ3XMG764J',
        now(),
        now(),
        '-2147483637'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA8',
        '01GSX7KMWVTMH0E8NZ3ZV9J5XG',
        now(),
        now(),
        '-2147483638'
    ), (
        '01GG9EWWVNBHGDM01WQERA8MX6',
        '01GSX7KMWVTMH0E8NZ3ZV9J5XG',
        now(),
        now(),
        '-2147483638'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA7',
        '01GSX7KMWVTMH0E8NZ41AGQE94',
        now(),
        now(),
        '-2147483639'
    ), (
        '01GG9EWWVNBHGDM01WQERA8MX8',
        '01GSX7KMWVTMH0E8NZ41AGQE94',
        now(),
        now(),
        '-2147483639'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA6',
        '01GSX7KMWVTMH0E8NZ42AGGVSB',
        now(),
        now(),
        '-2147483640'
    ), (
        '01GDMTJ3NB0TKRWS7ZRMHFTHH8',
        '01GSX7KMWVTMH0E8NZ42AGGVSB',
        now(),
        now(),
        '-2147483640'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA5',
        '01GSX7KMWVTMH0E8NZ45ZE0ZAH',
        now(),
        now(),
        '-2147483641'
    ), (
        '01GDMTJ3NB0TKRWS7ZRMHFTHH6',
        '01GSX7KMWVTMH0E8NZ45ZE0ZAH',
        now(),
        now(),
        '-2147483641'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA4',
        '01GSX7KMWVTMH0E8NZ49247D1H',
        now(),
        now(),
        '-2147483642'
    ), (
        '01GDMTJ3NB0TKRWS7ZRMHFTHH4',
        '01GSX7KMWVTMH0E8NZ49247D1H',
        now(),
        now(),
        '-2147483642'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA3',
        '01GSX7KMWVTMH0E8NZ4BVCBB3Z',
        now(),
        now(),
        '-2147483643'
    ), (
        '01GDMTJ3NB0TKRWS7ZRMHFTHH0',
        '01GSX7KMWVTMH0E8NZ4BVCBB3Z',
        now(),
        now(),
        '-2147483643'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA2',
        '01GSX7KMWVTMH0E8NZ4DRAJ1R9',
        now(),
        now(),
        '-2147483644'
    ), (
        '01GDMTJ3NB0TKRWS7ZRMHFTHH1',
        '01GSX7KMWVTMH0E8NZ4DRAJ1R9',
        now(),
        now(),
        '-2147483644'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YA1',
        '01GSX7KMWVTMH0E8NZ4GXDD0A9',
        now(),
        now(),
        '-2147483645'
    ), (
        '01GG9EWWVPFT4HTY1E7RF9F4B0',
        '01GSX7KMWVTMH0E8NZ4GXDD0A9',
        now(),
        now(),
        '-2147483645'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YK9',
        '01GSX7KMWVTMH0E8NZ4HTHA56Q',
        now(),
        now(),
        '-2147483646'
    ), (
        '01GG9EWWVPFT4HTY1E7RF9F4B2',
        '01GSX7KMWVTMH0E8NZ4HTHA56Q',
        now(),
        now(),
        '-2147483646'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YK8',
        '01GSX7KMWVTMH0E8NZ4JDVNX69',
        now(),
        now(),
        '-2147483647'
    ), (
        '01GG9EWWVPFT4HTY1E7RF9F4B4',
        '01GSX7KMWVTMH0E8NZ4JDVNX69',
        now(),
        now(),
        '-2147483647'
    ), (
        '01GCGRJG8WBF3J0CVZ4XDK2YK7',
        '01GSX7KMWVTMH0E8NZ4NK34A2E',
        now(),
        now(),
        '-2147483648'
    ), (
        '01GG9EWWVPFT4HTY1E7RF9F4B6',
        '01GSX7KMWVTMH0E8NZ4NK34A2E',
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.user_group (
        user_group_id,
        user_group_name,
        is_system,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ4R7E49T1',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483629'
    ), (
        '01GSX7KMWVTMH0E8NZ4SVXVTB7',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483630'
    ), (
        '01GSX7KMWVTMH0E8NZ4V0XFJND',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483631'
    ), (
        '01GSX7KMWVTMH0E8NZ4V5RSE9D',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483634'
    ), (
        '01GSX7KMWVTMH0E8NZ4WY8C5D3',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483635'
    ), (
        '01GSX7KMWVTMH0E8NZ502CF6HC',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483637'
    ), (
        '01GSX7KMWVTMH0E8NZ50MFQDZH',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483638'
    ), (
        '01GSX7KMWVTMH0E8NZ50N1KW1B',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483639'
    ), (
        '01GSX7KMWVTMH0E8NZ53949S4F',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483640'
    ), (
        '01GSX7KMWVTMH0E8NZ563YXW3W',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483641'
    ), (
        '01GSX7KMWVTMH0E8NZ578Y135Z',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483642'
    ), (
        '01GSX7KMWVTMH0E8NZ59NH878M',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483643'
    ), (
        '01GSX7KMWVTMH0E8NZ5BQF2596',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483644'
    ), (
        '01GSX7KMWVTMH0E8NZ5CV8JJ04',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483645'
    ), (
        '01GSX7KMWVTMH0E8NZ5EZJ0Q45',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483646'
    ), (
        '01GSX7KMWVTMH0E8NZ5GC42STC',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483647'
    ), (
        '01GSX7KMWVTMH0E8NZ5J6JX2A1',
        'LessonmgmtScheduleJob',
        true,
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.granted_role (
        granted_role_id,
        user_group_id,
        role_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ5M5P6C16',
        '01GSX7KMWVTMH0E8NZ4R7E49T1',
        '01GSX7KMWVTMH0E8NZ3PBYBPZM',
        now(),
        now(),
        '-2147483629'
    ), (
        '01GSX7KMWVTMH0E8NZ5NSQ4V6Z',
        '01GSX7KMWVTMH0E8NZ4SVXVTB7',
        '01GSX7KMWVTMH0E8NZ3RS5PV7M',
        now(),
        now(),
        '-2147483630'
    ), (
        '01GSX7KMWVTMH0E8NZ5SK9DWAT',
        '01GSX7KMWVTMH0E8NZ4V0XFJND',
        '01GSX7KMWVTMH0E8NZ3SCJQEY1',
        now(),
        now(),
        '-2147483631'
    ), (
        '01GSX7KMWVTMH0E8NZ5T0QEN0G',
        '01GSX7KMWVTMH0E8NZ4V5RSE9D',
        '01GSX7KMWVTMH0E8NZ3T6Q5038',
        now(),
        now(),
        '-2147483634'
    ), (
        '01GSX7KMWVTMH0E8NZ5VAEF5ST',
        '01GSX7KMWVTMH0E8NZ4WY8C5D3',
        '01GSX7KMWVTMH0E8NZ3VSVJH4A',
        now(),
        now(),
        '-2147483635'
    ), (
        '01GSX7KMWVTMH0E8NZ5XJWQFQY',
        '01GSX7KMWVTMH0E8NZ502CF6HC',
        '01GSX7KMWVTMH0E8NZ3XMG764J',
        now(),
        now(),
        '-2147483637'
    ), (
        '01GSX7KMWVTMH0E8NZ5ZHXSCYM',
        '01GSX7KMWVTMH0E8NZ50MFQDZH',
        '01GSX7KMWVTMH0E8NZ3ZV9J5XG',
        now(),
        now(),
        '-2147483638'
    ), (
        '01GSX7KMWVTMH0E8NZ6233CMXB',
        '01GSX7KMWVTMH0E8NZ50N1KW1B',
        '01GSX7KMWVTMH0E8NZ41AGQE94',
        now(),
        now(),
        '-2147483639'
    ), (
        '01GSX7KMWVTMH0E8NZ62JNEJPJ',
        '01GSX7KMWVTMH0E8NZ53949S4F',
        '01GSX7KMWVTMH0E8NZ42AGGVSB',
        now(),
        now(),
        '-2147483640'
    ), (
        '01GSX7KMWVTMH0E8NZ65QYJA0N',
        '01GSX7KMWVTMH0E8NZ563YXW3W',
        '01GSX7KMWVTMH0E8NZ45ZE0ZAH',
        now(),
        now(),
        '-2147483641'
    ), (
        '01GSX7KMWVTMH0E8NZ68D4ZFK6',
        '01GSX7KMWVTMH0E8NZ578Y135Z',
        '01GSX7KMWVTMH0E8NZ49247D1H',
        now(),
        now(),
        '-2147483642'
    ), (
        '01GSX7KMWVTMH0E8NZ68V72W87',
        '01GSX7KMWVTMH0E8NZ59NH878M',
        '01GSX7KMWVTMH0E8NZ4BVCBB3Z',
        now(),
        now(),
        '-2147483643'
    ), (
        '01GSX7KMWVTMH0E8NZ6BKQ86JE',
        '01GSX7KMWVTMH0E8NZ5BQF2596',
        '01GSX7KMWVTMH0E8NZ4DRAJ1R9',
        now(),
        now(),
        '-2147483644'
    ), (
        '01GSX7KMWVTMH0E8NZ6D2KXQSD',
        '01GSX7KMWVTMH0E8NZ5CV8JJ04',
        '01GSX7KMWVTMH0E8NZ4GXDD0A9',
        now(),
        now(),
        '-2147483645'
    ), (
        '01GSX7KMWVTMH0E8NZ6D8M344S',
        '01GSX7KMWVTMH0E8NZ5EZJ0Q45',
        '01GSX7KMWVTMH0E8NZ4HTHA56Q',
        now(),
        now(),
        '-2147483646'
    ), (
        '01GSX7KMWVTMH0E8NZ6ECXR08A',
        '01GSX7KMWVTMH0E8NZ5GC42STC',
        '01GSX7KMWVTMH0E8NZ4JDVNX69',
        now(),
        now(),
        '-2147483647'
    ), (
        '01GSX7KMWVTMH0E8NZ6HCDY3EE',
        '01GSX7KMWVTMH0E8NZ5J6JX2A1',
        '01GSX7KMWVTMH0E8NZ4NK34A2E',
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.granted_role_access_path (
        granted_role_id,
        location_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ5M5P6C16',
        '01GFMNHQ1WHGRC8AW6K913AM3G',
        now(),
        now(),
        '-2147483629'
    ), (
        '01GSX7KMWVTMH0E8NZ5NSQ4V6Z',
        '01GFMMFRXC6SKTTT44HWR3BRY8',
        now(),
        now(),
        '-2147483630'
    ), (
        '01GSX7KMWVTMH0E8NZ5SK9DWAT',
        '01GDWSMJS6APH4SX2NP5NFWHG5',
        now(),
        now(),
        '-2147483631'
    ), (
        '01GSX7KMWVTMH0E8NZ5T0QEN0G',
        '01FR4M51XJY9E77GSN4QZ1Q8N5',
        now(),
        now(),
        '-2147483634'
    ), (
        '01GSX7KMWVTMH0E8NZ5VAEF5ST',
        '01FR4M51XJY9E77GSN4QZ1Q8N4',
        now(),
        now(),
        '-2147483635'
    ), (
        '01GSX7KMWVTMH0E8NZ5XJWQFQY',
        '01FR4M51XJY9E77GSN4QZ1Q8N3',
        now(),
        now(),
        '-2147483637'
    ), (
        '01GSX7KMWVTMH0E8NZ5ZHXSCYM',
        '01FR4M51XJY9E77GSN4QZ1Q8N2',
        now(),
        now(),
        '-2147483638'
    ), (
        '01GSX7KMWVTMH0E8NZ6233CMXB',
        '01FR4M51XJY9E77GSN4QZ1Q8N1',
        now(),
        now(),
        '-2147483639'
    ), (
        '01GSX7KMWVTMH0E8NZ62JNEJPJ',
        '01FR4M51XJY9E77GSN4QZ1Q9N9',
        now(),
        now(),
        '-2147483640'
    ), (
        '01GSX7KMWVTMH0E8NZ65QYJA0N',
        '01FR4M51XJY9E77GSN4QZ1Q9N8',
        now(),
        now(),
        '-2147483641'
    ), (
        '01GSX7KMWVTMH0E8NZ68D4ZFK6',
        '01FR4M51XJY9E77GSN4QZ1Q9N7',
        now(),
        now(),
        '-2147483642'
    ), (
        '01GSX7KMWVTMH0E8NZ68V72W87',
        '01FR4M51XJY9E77GSN4QZ1Q9N6',
        now(),
        now(),
        '-2147483643'
    ), (
        '01GSX7KMWVTMH0E8NZ6BKQ86JE',
        '01FR4M51XJY9E77GSN4QZ1Q9N5',
        now(),
        now(),
        '-2147483644'
    ), (
        '01GSX7KMWVTMH0E8NZ6D2KXQSD',
        '01FR4M51XJY9E77GSN4QZ1Q9N4',
        now(),
        now(),
        '-2147483645'
    ), (
        '01GSX7KMWVTMH0E8NZ6D8M344S',
        '01FR4M51XJY9E77GSN4QZ1Q9N3',
        now(),
        now(),
        '-2147483646'
    ), (
        '01GSX7KMWVTMH0E8NZ6ECXR08A',
        '01FR4M51XJY9E77GSN4QZ1Q9N2',
        now(),
        now(),
        '-2147483647'
    ), (
        '01GSX7KMWVTMH0E8NZ6HCDY3EE',
        '01FR4M51XJY9E77GSN4QZ1Q9N1',
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ4R7E49T1') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ4SVXVTB7') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ4V0XFJND') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ4V5RSE9D') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ4WY8C5D3') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ502CF6HC') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ50MFQDZH') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ50N1KW1B') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ53949S4F') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ563YXW3W') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ578Y135Z') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ59NH878M') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ5BQF2596') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ5CV8JJ04') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ5EZJ0Q45') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ5GC42STC') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    granted_permission (
        user_group_id,
        user_group_name,
        role_id,
        role_name,
        permission_id,
        permission_name,
        location_id,
        resource_path
    )
SELECT *
FROM
    retrieve_src_granted_permission('01GSX7KMWVTMH0E8NZ5J6JX2A1') ON CONFLICT ON CONSTRAINT granted_permission__pk
DO
UPDATE
SET
    user_group_name = excluded.user_group_name;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWVTMH0E8NZ6HHZRCS1',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483629'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ6HHZRCS1',
        '01GSX7KMWVTMH0E8NZ4R7E49T1',
        now(),
        now(),
        '-2147483629'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWVTMH0E8NZ6KDGYXQ8',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483630'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ6KDGYXQ8',
        '01GSX7KMWVTMH0E8NZ4SVXVTB7',
        now(),
        now(),
        '-2147483630'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWVTMH0E8NZ6KTD7TFM',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483631'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ6KTD7TFM',
        '01GSX7KMWVTMH0E8NZ4V0XFJND',
        now(),
        now(),
        '-2147483631'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWVTMH0E8NZ6MC88MHG',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483634'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWVTMH0E8NZ6MC88MHG',
        '01GSX7KMWVTMH0E8NZ4V5RSE9D',
        now(),
        now(),
        '-2147483634'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDK4C1XYP',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483635'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDK4C1XYP',
        '01GSX7KMWVTMH0E8NZ4WY8C5D3',
        now(),
        now(),
        '-2147483635'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDPAE0ME4',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483637'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDPAE0ME4',
        '01GSX7KMWVTMH0E8NZ502CF6HC',
        now(),
        now(),
        '-2147483637'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDQG1DRX3',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483638'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDQG1DRX3',
        '01GSX7KMWVTMH0E8NZ50MFQDZH',
        now(),
        now(),
        '-2147483638'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDSYYGR4G',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483639'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDSYYGR4G',
        '01GSX7KMWVTMH0E8NZ50N1KW1B',
        now(),
        now(),
        '-2147483639'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDVZA9GHP',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483640'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDVZA9GHP',
        '01GSX7KMWVTMH0E8NZ53949S4F',
        now(),
        now(),
        '-2147483640'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDZ45GSHY',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483641'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDZ45GSHY',
        '01GSX7KMWVTMH0E8NZ563YXW3W',
        now(),
        now(),
        '-2147483641'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDZ7ZX525',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483642'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDZ7ZX525',
        '01GSX7KMWVTMH0E8NZ578Y135Z',
        now(),
        now(),
        '-2147483642'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GDZ8AW30Z',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483643'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GDZ8AW30Z',
        '01GSX7KMWVTMH0E8NZ59NH878M',
        now(),
        now(),
        '-2147483643'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GE0JF1THV',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483644'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GE0JF1THV',
        '01GSX7KMWVTMH0E8NZ5BQF2596',
        now(),
        now(),
        '-2147483644'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GE0ZNQTJ5',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483645'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GE0ZNQTJ5',
        '01GSX7KMWVTMH0E8NZ5CV8JJ04',
        now(),
        now(),
        '-2147483645'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GE1WC0CKA',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483646'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GE1WC0CKA',
        '01GSX7KMWVTMH0E8NZ5EZJ0Q45',
        now(),
        now(),
        '-2147483646'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GE2DBHSRW',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483647'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GE2DBHSRW',
        '01GSX7KMWVTMH0E8NZ5GC42STC',
        now(),
        now(),
        '-2147483647'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    public.users (
        user_id,
        country,
        name,
        avatar,
        phone_number,
        email,
        device_token,
        allow_notification,
        user_group,
        updated_at,
        created_at,
        is_tester,
        facebook_id,
        platform,
        phone_verified,
        email_verified,
        deleted_at,
        given_name,
        is_system,
        resource_path
    )
VALUES
(
        '01GSX7KMWWED9ZZ79GE4JTJ309',
        'COUNTRY_JP',
        'Lesson Schedule Job',
        '',
        NULL,
        'schedule_job+lessonmgmt@manabie.com',
        NULL,
        NULL,
        'USER_GROUP_SCHOOL_ADMIN',
        now(),
        now(),
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        true,
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;

INSERT INTO
    user_group_member(
        user_id,
        user_group_id,
        created_at,
        updated_at,
        resource_path
    )
VALUES (
        '01GSX7KMWWED9ZZ79GE4JTJ309',
        '01GSX7KMWVTMH0E8NZ5J6JX2A1',
        now(),
        now(),
        '-2147483648'
    ) ON CONFLICT
DO NOTHING;
