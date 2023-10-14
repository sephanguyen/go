CREATE TABLE IF NOT EXISTS invoicemgmt.partner_convenience_store (
    partner_convenience_store_id text  NOT NULL,
    manufacturer_code integer NOT NULL,
    company_code integer NOT NULL,
    shop_code text,
    company_name text NOT NULL,
    company_tel_number text,
    postal_code text,
    address1 text,
    address2 text,
    message1 text,
    message2 text,
    message3 text,
    message4 text,
    message5 text,
    message6 text,
    message7 text,
    message8 text,
    message9 text,
    message10 text,
    message11 text,
    message12 text,
    message13 text,
    message14 text,
    message15 text,
    message16 text,
    message17 text,
    message18 text,
    message19 text,
    message20 text,
    message21 text,
    message22 text,
    message23 text,
    message24 text,
    remarks text,
    is_archived boolean DEFAULT false,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT partner_convenience_store___pk PRIMARY KEY (partner_convenience_store_id)
);

CREATE TABLE IF NOT EXISTS invoicemgmt.company_detail
(
    company_detail_id text NOT NULL,
    company_name text NOT NULL,
    company_address text,
    company_phone_number text,
    company_logo_url text,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at timestamp with time zone,
    resource_path text,
    CONSTRAINT pk__company_detail PRIMARY KEY (company_detail_id)
);


ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.partner_convenience_store;
ALTER PUBLICATION publication_for_datawarehouse ADD TABLE invoicemgmt.company_detail;
