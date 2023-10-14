CREATE TABLE IF NOT EXISTS public.groups (
    group_id text NOT NULL,
    name text NOT NULL,
    description text,
    privileges JSONB,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__groups PRIMARY KEY (group_id)
);

CREATE TABLE IF NOT EXISTS public.users_groups (
    user_id text NOT NULL,
    group_id text NOT NULL,
    is_origin bool NOT NULL,
    status TEXT NOT NULL, -- USER_GROUP_STATUS_ACTIVE, USER_GROUP_STATUS_INACTIVE
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT pk__users_groups PRIMARY KEY (user_id, group_id),
    CONSTRAINT fk__users_groups__user_id FOREIGN KEY (user_id) REFERENCES public.users (user_id),
    CONSTRAINT fk__users_groups__group_id FOREIGN KEY (group_id) REFERENCES public.groups (group_id)
);

INSERT INTO public.groups (group_id,name,created_at,updated_at) VALUES
('USER_GROUP_STUDENT','Student', NOW(), NOW()),
('USER_GROUP_COACH','Coach', NOW(), NOW()),
('USER_GROUP_TUTOR','Tutor', NOW(), NOW()),
('USER_GROUP_STAFF','Staff', NOW(), NOW()),
('USER_GROUP_ADMIN','Admin', NOW(), NOW()),
('USER_GROUP_TEACHER','Teacher', NOW(), NOW()),
('USER_GROUP_PARENT','Parent', NOW(), NOW()),
('USER_GROUP_CONTENT_ADMIN','Content Admin', NOW(), NOW()),
('USER_GROUP_CONTENT_STAFF','Content Staff', NOW(), NOW()),
('USER_GROUP_SALES_ADMIN','Sales Admin', NOW(), NOW()),
('USER_GROUP_SALES_STAFF','Sales Staff', NOW(), NOW()),
('USER_GROUP_CS_ADMIN','CS Admin', NOW(), NOW()),
('USER_GROUP_CS_STAFF','CS Staff', NOW(), NOW()),
('USER_GROUP_SCHOOL_ADMIN','School Admin', NOW(), NOW()),
('USER_GROUP_SCHOOL_STAFF','Schook Staff', NOW(), NOW())
ON CONFLICT DO NOTHING;

INSERT INTO users_groups
SELECT user_id, user_group, TRUE, 'USER_GROUP_STATUS_ACTIVE', NOW(), NOW() FROM users
ON CONFLICT DO NOTHING;
