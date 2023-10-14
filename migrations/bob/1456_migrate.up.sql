-- insert permission payment.order.write, payment.bill_item.read for Centre Staff at resource_path = '-2147483648'

INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES ('01GDM4F5RAWE8ZG9AGHWW64AX8','01G7XGB49W2PCQPHNBE6SAZ485', now(), now(),'-2147483648'),
       ('01GDM4F5RBBPCR9S8RPXEKW1F8','01G7XGB49W2PCQPHNBE6SAZ485', now(), now(),'-2147483648')
    ON CONFLICT DO NOTHING;

-- insert permission payment.bill_item.read for Centre Manager at resource_path = '-2147483648'

INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES ('01GDM4F5RBBPCR9S8RPXEKW1F8','01G7XGB49W2PCQPHNBE6SAZ484', now(), now(),'-2147483648')
    ON CONFLICT DO NOTHING;
