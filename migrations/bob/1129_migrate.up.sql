update info_notifications set resource_path = owner where resource_path is null;

update info_notification_msgs as nm set resource_path = n.resource_path
    from info_notifications as n where nm.notification_msg_id = n.notification_msg_id ;

update users_info_notifications as nu set resource_path = n.resource_path
    from info_notifications as n where nu.notification_id = n.notification_id;
