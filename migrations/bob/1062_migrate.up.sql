CREATE TABLE IF NOT EXISTS info_notification_msgs
(
	notification_msg_id TEXT,
	title TEXT NOT NULL,
	content jsonb,
	media_ids TEXT[],
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT pk__info_notification_msgs PRIMARY KEY(notification_msg_id)
);

CREATE TABLE IF NOT EXISTS info_notifications
(
	notification_id TEXT,
	notification_msg_id TEXT,
	type TEXT,
	data jsonb,
	editor_id TEXT,
	target_groups jsonb,
	receiver_ids TEXT[],
	event TEXT,
	status TEXT,
	scheduled_at TIMESTAMP WITH TIME ZONE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT pk__info_notifications PRIMARY KEY(notification_id),
	CONSTRAINT fk__info_notifications__notification_msg_id FOREIGN KEY(notification_msg_id) REFERENCES info_notification_msgs(notification_msg_id)
);

CREATE TABLE IF NOT EXISTS users_info_notifications 
(
	user_notification_id TEXT,
	notification_id TEXT,
	user_id TEXT,
	status TEXT,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL,
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
	deleted_at TIMESTAMP WITH TIME ZONE,
	CONSTRAINT pk__users_info_notifications PRIMARY KEY(user_notification_id),
	CONSTRAINT fk__users_info_notifications__notification_id FOREIGN KEY(notification_id) REFERENCES info_notifications(notification_id),
	CONSTRAINT fk__users_info_notifications__user_id FOREIGN KEY(user_id) REFERENCES public.users(user_id)
);