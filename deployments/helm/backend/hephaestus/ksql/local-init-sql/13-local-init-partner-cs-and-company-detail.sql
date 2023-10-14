\connect invoicemgmt;


INSERT INTO public.partner_convenience_store(
    partner_convenience_store_id, manufacturer_code, company_code, shop_code, company_name, company_tel_number, postal_code, address1, address2, message1, message2, message3, message4,
    message5, message6, message7, message8, message9, message10, message11, message12, message13, message14, message15, message16, message17, message18, message19, message20,
    message21, message22, message23, message24, remarks, is_archived, updated_at, created_at, resource_path
) VALUES (
    '01H79YTHM52TJ34RSECNHVCT1G', 
    123456,
    12345,
    '1234',
    'init company',
    '01234567',
    'init postal code',
    'init address 1',
    'init address 2',
    'init message 1',
    'init message 2',
    'init message 3',
    'init message 4',
    'init message 5',
    'init message 6',
    'init message 7',
    'init message 8',
    'init message 9',
    'init message 10',
    'init message 11',
    'init message 12',
    'init message 13',
    'init message 14',
    'init message 15',
    'init message 16',
    'init message 17',
    'init message 18',
    'init message 19',
    'init message 20',
    'init message 21',
    'init message 22',
    'init message 23',
    'init message 24',
    'init remarks',
    false,
    now(),
    now(), 
    '-2147483642'
);


INSERT INTO public.company_detail(
	company_detail_id, company_name, company_address, company_phone_number, company_logo_url, created_at, updated_at, resource_path
) VALUES ('01H7A5EEKB672CR1MAWQSZC2TN', 'init company name', 'init company address', '1234567', 'init company logo url', now(), now(), '-2147483642') ON CONFLICT DO NOTHING;
