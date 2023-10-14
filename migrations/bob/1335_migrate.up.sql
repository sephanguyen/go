-- (A.1) Remove permission_role is read for Centre Lead role
delete from permission_role pr
where (pr.resource_path = '-2147483640' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ402' and pr.permission_id='01GEGJS281QPNKRSNKZWBR58K0')
   or (pr.resource_path = '-2147483641' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ412' and pr.permission_id='01GEGJRRF8J218N1CQJ18FAV22')
   or (pr.resource_path = '-2147483642' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ422' and pr.permission_id='01GEGJREPBFHTG1FYJRDB11Q0W')
   or (pr.resource_path = '-2147483643' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ432' and pr.permission_id='01GEGJR4XKQJ71DJNB9VECYWQG')
   or (pr.resource_path = '-2147483644' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ442' and pr.permission_id='01GEGJQV4T8NWAT0BVVD2VRRRB')
   or (pr.resource_path = '-2147483645' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ452' and pr.permission_id='01GEGJQHC1F8Q1YA76YT9FGR76')
   or (pr.resource_path = '-2147483646' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ462' and pr.permission_id='01GEGJQ7K89J1K0REAVBJVAK3G')
   or (pr.resource_path = '-2147483647' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ472' and pr.permission_id='01GEGJPXTGCSZBASY3TTQ8JEFM')
   or (pr.resource_path = '-2147483648' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ482' and pr.permission_id='01GEGJPM1QGC9P8275RE5AB79C')
;

-- (A.2) Remove permission_role is write for Centre Lead role
delete from permission_role pr
where (pr.resource_path = '-2147483640' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ402' and pr.permission_id='01GEGJVT6XJ4K31ZDWDA72HFJ7')
   or (pr.resource_path = '-2147483641' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ412' and pr.permission_id='01GEGJVGE4K9NW32A1H7QMCXRN')
   or (pr.resource_path = '-2147483642' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ422' and pr.permission_id='01GEGJV6NCVJD0VGZ2731N8AE4')
   or (pr.resource_path = '-2147483643' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ432' and pr.permission_id='01GEGJTWWKGPVH0BZ4PH8MAGQQ')
   or (pr.resource_path = '-2147483644' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ442' and pr.permission_id='01GEGJTK3W6QEC1QAMC8XDSZ2V')
   or (pr.resource_path = '-2147483645' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ452' and pr.permission_id='01GEGJT9B3F4WS4A18F4TVX67Q')
   or (pr.resource_path = '-2147483646' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ462' and pr.permission_id='01GEGJSZJBQND2WE73D6BV39X1')
   or (pr.resource_path = '-2147483647' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ472' and pr.permission_id='01GEGJSNSJ36VQMYJNMXJHTBT1')
   or (pr.resource_path = '-2147483648' and  pr.role_id ='01G7XGB49W2PCQPHNBE6SAZ482' and pr.permission_id='01GEGJSC0SP1BCYW7Y09BDX4EV')
;

-- (B.1) Associate timesheet.timesheet.read with Centre Manager role
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GEGJPM1QGC9P8275RE5AB79C', '01G7XGB49W2PCQPHNBE6SAZ484', now(), now(), '-2147483648'),
    ('01GEGJPXTGCSZBASY3TTQ8JEFM', '01G7XGB49W2PCQPHNBE6SAZ474', now(), now(), '-2147483647'),
    ('01GEGJQ7K89J1K0REAVBJVAK3G', '01G7XGB49W2PCQPHNBE6SAZ464', now(), now(), '-2147483646'),
    ('01GEGJQHC1F8Q1YA76YT9FGR76', '01G7XGB49W2PCQPHNBE6SAZ454', now(), now(), '-2147483645'),
    ('01GEGJQV4T8NWAT0BVVD2VRRRB', '01G7XGB49W2PCQPHNBE6SAZ444', now(), now(), '-2147483644'),
    ('01GEGJR4XKQJ71DJNB9VECYWQG', '01G7XGB49W2PCQPHNBE6SAZ434', now(), now(), '-2147483643'),
    ('01GEGJREPBFHTG1FYJRDB11Q0W', '01G7XGB49W2PCQPHNBE6SAZ424', now(), now(), '-2147483642'),
    ('01GEGJRRF8J218N1CQJ18FAV22', '01G7XGB49W2PCQPHNBE6SAZ414', now(), now(), '-2147483641'),
    ('01GEGJS281QPNKRSNKZWBR58K0', '01G7XGB49W2PCQPHNBE6SAZ404', now(), now(), '-2147483640')
    ON CONFLICT DO NOTHING;

-- (B.2) Associate timesheet.timesheet.write with Centre Manager role
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GEGJSC0SP1BCYW7Y09BDX4EV', '01G7XGB49W2PCQPHNBE6SAZ484', now(), now(), '-2147483648'),
    ('01GEGJSNSJ36VQMYJNMXJHTBT1', '01G7XGB49W2PCQPHNBE6SAZ474', now(), now(), '-2147483647'),
    ('01GEGJSZJBQND2WE73D6BV39X1', '01G7XGB49W2PCQPHNBE6SAZ464', now(), now(), '-2147483646'),
    ('01GEGJT9B3F4WS4A18F4TVX67Q', '01G7XGB49W2PCQPHNBE6SAZ454', now(), now(), '-2147483645'),
    ('01GEGJTK3W6QEC1QAMC8XDSZ2V', '01G7XGB49W2PCQPHNBE6SAZ444', now(), now(), '-2147483644'),
    ('01GEGJTWWKGPVH0BZ4PH8MAGQQ', '01G7XGB49W2PCQPHNBE6SAZ434', now(), now(), '-2147483643'),
    ('01GEGJV6NCVJD0VGZ2731N8AE4', '01G7XGB49W2PCQPHNBE6SAZ424', now(), now(), '-2147483642'),
    ('01GEGJVGE4K9NW32A1H7QMCXRN', '01G7XGB49W2PCQPHNBE6SAZ414', now(), now(), '-2147483641'),
    ('01GEGJVT6XJ4K31ZDWDA72HFJ7', '01G7XGB49W2PCQPHNBE6SAZ404', now(), now(), '-2147483640')
    ON CONFLICT DO NOTHING;
