INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','NOTIFICATION_EVENT_TEACHER_GIVE_ASSIGNMENT','USER_GROUP_STUDENT','Giáo viên của bạn vừa giao bài tập.','Giáo viên của bạn vừa giao bài tập.',now(),now())
,('COUNTRY_MASTER','NOTIFICATION_EVENT_TEACHER_GIVE_ASSIGNMENT','USER_GROUP_STUDENT','You''ve got a new assignment from your teacher','You''ve got a new assignment from your teacher',now(),now())
ON CONFLICT DO NOTHING;

INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT','USER_GROUP_STUDENT','Teacher returned your assignment. Let''s see the comments!','Teacher returned your assignment. Let''s see the comments!',now(),now())
,('COUNTRY_MASTER','NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT','USER_GROUP_STUDENT','Teacher returned your assignment. Let''s see the comments!','Teacher returned your assignment. Let''s see the comments!',now(),now())
ON CONFLICT DO NOTHING;

INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','NOTIFICATION_EVENT_STUDENT_SUBMIT_ASSIGNMENT','USER_GROUP_TEACHER','Học sinh của bạn vừa nộp bài.','Học sinh của bạn vừa nộp bài.',now(),now())
,('COUNTRY_MASTER','NOTIFICATION_EVENT_TEACHER_RETURN_ASSIGNMENT','USER_GROUP_TEACHER','Your student submitted an assignment','Your student submitted an assignment',now(),now())
ON CONFLICT DO NOTHING;

INSERT INTO public.notification_messages (country,"key",receiver_group,title,body,created_at,updated_at) VALUES
('COUNTRY_VN','NOTIFICATION_EVENT_ASSIGNMENT_UPDATED','USER_GROUP_STUDENT','Bài tập {{.AssignmentName}} vừa được thay đổi','Assignment {{.AssignmentName}} vừa được thay đổi',now(),now())
,('COUNTRY_MASTER','NOTIFICATION_EVENT_ASSIGNMENT_UPDATED','USER_GROUP_STUDENT','Assignment {{.AssignmentName}} is updated recently','Assignment {{.AssignmentName}} is updated recently',now(),now())
ON CONFLICT DO NOTHING;
