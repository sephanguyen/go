ALTER TABLE bob.user_group_public_info 
DROP CONSTRAINT IF EXISTS pk__user_group;
ALTER TABLE bob.user_phone_number_public_info 
DROP CONSTRAINT IF EXISTS pk__user_phone_number;
ALTER TABLE bob.user_address_public_info 
DROP CONSTRAINT IF EXISTS pk__user_address;

CREATE TABLE IF NOT EXISTS bob.user_group (
    user_group_id text,
	user_group_name text,
	created_at timestamptz,
	updated_at timestamptz,
	deleted_at timestamptz,
	resource_path text,
	org_location_id text,
	is_system bool,

    CONSTRAINT pk__user_group PRIMARY KEY (user_group_id)
);

CREATE TABLE IF NOT EXISTS bob.user_phone_number (
    user_phone_number_id text,
	user_id text,
	phone_number text,
	"type" text,
	updated_at timestamptz,
	created_at timestamptz,
	deleted_at timestamptz,
	
	CONSTRAINT pk__user_phone_number PRIMARY KEY (user_phone_number_id)
);

CREATE TABLE IF NOT EXISTS bob.user_address (
    student_address_id text,
	student_id text,
	address_type text,
	postal_code text,
	prefecture_id text,
	city text,
	user_address_created_at timestamptz,
	user_address_updated_at timestamptz,
	user_address_deleted_at timestamptz,
	resource_path text,
	first_street text,
	second_street text,
	
	CONSTRAINT pk__user_address PRIMARY KEY (student_address_id)
);

ALTER PUBLICATION kec_publication SET TABLE 
bob.user_group,    
bob.user_phone_number,
bob.user_address
