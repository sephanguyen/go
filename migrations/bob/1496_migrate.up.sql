UPDATE public.role
SET
    role_name = 'VirtualClassroomScheduleJob'
WHERE
    role_id IN(
        '01GSX7KMWVTMH0E8NZ3PBYBPZM',
        '01GSX7KMWVTMH0E8NZ3RS5PV7M',
        '01GSX7KMWVTMH0E8NZ3SCJQEY1',
        '01GSX7KMWVTMH0E8NZ3T6Q5038',
        '01GSX7KMWVTMH0E8NZ3VSVJH4A',
        '01GSX7KMWVTMH0E8NZ3XMG764J',
        '01GSX7KMWVTMH0E8NZ3ZV9J5XG',
        '01GSX7KMWVTMH0E8NZ41AGQE94',
        '01GSX7KMWVTMH0E8NZ42AGGVSB',
        '01GSX7KMWVTMH0E8NZ45ZE0ZAH',
        '01GSX7KMWVTMH0E8NZ49247D1H',
        '01GSX7KMWVTMH0E8NZ4BVCBB3Z',
        '01GSX7KMWVTMH0E8NZ4DRAJ1R9',
        '01GSX7KMWVTMH0E8NZ4GXDD0A9',
        '01GSX7KMWVTMH0E8NZ4HTHA56Q',
        '01GSX7KMWVTMH0E8NZ4JDVNX69',
        '01GSX7KMWVTMH0E8NZ4NK34A2E'
    ) AND role_name='LessonmgmtScheduleJob';

UPDATE public.user_group
SET user_group_name='VirtualClassroomScheduleJob'
WHERE
    user_group_id IN(
        '01GSX7KMWVTMH0E8NZ4R7E49T1',
        '01GSX7KMWVTMH0E8NZ4SVXVTB7',
        '01GSX7KMWVTMH0E8NZ4V0XFJND',
        '01GSX7KMWVTMH0E8NZ4V5RSE9D',
        '01GSX7KMWVTMH0E8NZ4WY8C5D3',
        '01GSX7KMWVTMH0E8NZ502CF6HC',
        '01GSX7KMWVTMH0E8NZ50MFQDZH',
        '01GSX7KMWVTMH0E8NZ50N1KW1B',
        '01GSX7KMWVTMH0E8NZ53949S4F',
        '01GSX7KMWVTMH0E8NZ563YXW3W',
        '01GSX7KMWVTMH0E8NZ578Y135Z',
        '01GSX7KMWVTMH0E8NZ59NH878M',
        '01GSX7KMWVTMH0E8NZ5BQF2596',
        '01GSX7KMWVTMH0E8NZ5CV8JJ04',
        '01GSX7KMWVTMH0E8NZ5EZJ0Q45',
        '01GSX7KMWVTMH0E8NZ5GC42STC',
        '01GSX7KMWVTMH0E8NZ5J6JX2A1'
    )
    AND user_group_name = 'LessonmgmtScheduleJob';

UPDATE public.users
SET name='Virtual Classroom Schedule Job',email='schedule_job+virtualclassroom@manabie.com'
WHERE
    user_id IN(
        '01GSX7KMWVTMH0E8NZ6HHZRCS1',
        '01GSX7KMWVTMH0E8NZ6KDGYXQ8',
        '01GSX7KMWVTMH0E8NZ6KTD7TFM',
        '01GSX7KMWVTMH0E8NZ6MC88MHG',
        '01GSX7KMWWED9ZZ79GDK4C1XYP',
        '01GSX7KMWWED9ZZ79GDPAE0ME4',
        '01GSX7KMWWED9ZZ79GDQG1DRX3',
        '01GSX7KMWWED9ZZ79GDSYYGR4G',
        '01GSX7KMWWED9ZZ79GDVZA9GHP',
        '01GSX7KMWWED9ZZ79GDZ45GSHY',
        '01GSX7KMWWED9ZZ79GDZ7ZX525',
        '01GSX7KMWWED9ZZ79GDZ8AW30Z',
        '01GSX7KMWWED9ZZ79GE0JF1THV',
        '01GSX7KMWWED9ZZ79GE0ZNQTJ5',
        '01GSX7KMWWED9ZZ79GE1WC0CKA',
        '01GSX7KMWWED9ZZ79GE2DBHSRW',
        '01GSX7KMWWED9ZZ79GE4JTJ309'
    )
    AND name = 'Lesson Schedule Job';

DELETE FROM granted_permission WHERE
    user_group_id IN(
        '01GSX7KMWVTMH0E8NZ4R7E49T1',
        '01GSX7KMWVTMH0E8NZ4SVXVTB7',
        '01GSX7KMWVTMH0E8NZ4V0XFJND',
        '01GSX7KMWVTMH0E8NZ4V5RSE9D',
        '01GSX7KMWVTMH0E8NZ4WY8C5D3',
        '01GSX7KMWVTMH0E8NZ502CF6HC',
        '01GSX7KMWVTMH0E8NZ50MFQDZH',
        '01GSX7KMWVTMH0E8NZ50N1KW1B',
        '01GSX7KMWVTMH0E8NZ53949S4F',
        '01GSX7KMWVTMH0E8NZ563YXW3W',
        '01GSX7KMWVTMH0E8NZ578Y135Z',
        '01GSX7KMWVTMH0E8NZ59NH878M',
        '01GSX7KMWVTMH0E8NZ5BQF2596',
        '01GSX7KMWVTMH0E8NZ5CV8JJ04',
        '01GSX7KMWVTMH0E8NZ5EZJ0Q45',
        '01GSX7KMWVTMH0E8NZ5GC42STC',
        '01GSX7KMWVTMH0E8NZ5J6JX2A1'
    );

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
