-- create japanese_collation 
CREATE COLLATION IF NOT EXISTS japanese_collation (provider = icu, locale = 'en-u-kn-true-kr-digit-en-ja_JP');

-- create index for user_group table
DROP INDEX IF EXISTS user_group_user_group_name_idx;
CREATE INDEX IF NOT EXISTS user_group__user_group_name_idx ON public.user_group (user_group_name ASC);

-- create function get user_group and order by user_group_name with jp
DROP FUNCTION IF EXISTS get_sorted_user_groups;

CREATE OR REPLACE FUNCTION public.get_sorted_user_groups() RETURNS SETOF public.user_group
    LANGUAGE SQL STABLE
    AS $$
        SELECT * FROM user_group
        ORDER BY user_group_name COLLATE japanese_collation ASC
    $$