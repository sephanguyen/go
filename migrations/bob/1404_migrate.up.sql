-- Index support AC for notification message table
DROP INDEX IF EXISTS info_notification_msgs_title_idx;
DROP INDEX IF EXISTS idx__info_notification_msgs__title;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx__info_notification_msgs__title ON public.info_notification_msgs USING GIN (title gin_trgm_ops);

-- Index support AC for notification access path table
DROP INDEX IF EXISTS idx__info_notifications_access_paths__resource_path;
CREATE INDEX idx__info_notifications_access_paths__resource_path ON public.info_notifications_access_paths USING btree (resource_path);

-- Index support AC for notification table
DROP INDEX IF EXISTS idx__info_notifications__resource_path;
CREATE INDEX idx__info_notifications__resource_path ON public.info_notifications USING btree (resource_path);

DROP INDEX IF EXISTS info_notifications_notification_msg_id_idx;
DROP INDEX IF EXISTS idx__info_notifications__notification_msg_id;
CREATE INDEX idx__info_notifications__notification_msg_id ON public.info_notifications USING btree (notification_msg_id);

DROP INDEX IF EXISTS idx__info_notifications__type_deleted_at_resource_path;
DROP INDEX IF EXISTS idx__info_notifications__status_type_deleted_at_resource_path;
CREATE INDEX idx__info_notifications__status_type_deleted_at_resource_path ON public.info_notifications USING btree (
    status,
    type,
    deleted_at,
    resource_path
);
