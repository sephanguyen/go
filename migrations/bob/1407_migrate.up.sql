-- insert permission payment.bill_item.read for Center Lead

INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES ('01GDM4F5RBBPCR9S8RPXEKW1F8','01G7XGB49W2PCQPHNBE6SAZ482', now(), now(),'-2147483648'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F7','01G7XGB49W2PCQPHNBE6SAZ472', now(), now(),'-2147483647'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F6','01G7XGB49W2PCQPHNBE6SAZ462', now(), now(),'-2147483646'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F5','01G7XGB49W2PCQPHNBE6SAZ452', now(), now(),'-2147483645'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F4','01G7XGB49W2PCQPHNBE6SAZ442', now(), now(),'-2147483644'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F3','01G7XGB49W2PCQPHNBE6SAZ432', now(), now(),'-2147483643'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F2','01G7XGB49W2PCQPHNBE6SAZ422', now(), now(),'-2147483642'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F1','01G7XGB49W2PCQPHNBE6SAZ412', now(), now(),'-2147483641'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F0','01G7XGB49W2PCQPHNBE6SAZ402', now(), now(),'-2147483640'),
       ('01GDM4F5RBBPCR9S8RPXEKW1E9','01G7XGB49W2PCQPHNBE6SAZ392', now(), now(),'-2147483639'),
       ('01GDM4F5RBBPCR9S8RPXEKW1E8','01G7XGB49W2PCQPHNBE6SAZ382', now(), now(),'-2147483638'),
       ('01GDM4F5RBBPCR9S8RPXEKW1E7','01G7XGB49W2PCQPHNBE6SAZ372', now(), now(),'-2147483637'),
       ('01GDM4F5RBBPCR9S8RPXEKW1E5','01G8T49ECX7SCTC723ZH6Q2MBE6', now(), now(),'-2147483635'),
       ('01GDM4F5RBBPCR9S8RPXEKW1E4','01G8T4FQ2CZ0X2YD88HJE4CYGQ6', now(), now(),'-2147483634')
    ON CONFLICT DO NOTHING;