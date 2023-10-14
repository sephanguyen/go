ALTER TABLE public.course DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_course ON course ;
ALTER TABLE public.course RENAME COLUMN id TO course_id;
ALTER TABLE public.course RENAME CONSTRAINT course_pk TO courses_pk;

ALTER TABLE public.course RENAME TO courses;

ALTER TABLE public.courses ADD COLUMN grade smallint NOT NULL;
ALTER TABLE public.courses ADD COLUMN deleted_at timestamp with time zone;
ALTER TABLE public.courses DROP COLUMN IF EXISTS final_price;

ALTER TABLE public.package_course ADD CONSTRAINT fk_package_course_course_id FOREIGN KEY(course_id) REFERENCES courses(course_id);

CREATE POLICY rls_courses ON "courses" using (permission_check(resource_path, 'courses')) with check (permission_check(resource_path, 'courses'));

ALTER TABLE "courses" ENABLE ROW LEVEL security;
ALTER TABLE "courses" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.order_item_course (
    order_id text NOT NULL,
    package_id integer NOT NULL,
    course_id text NOT NULL,
    course_name text NOT NULL,
    course_slot integer,
    course_slot_per_week integer,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT order_item_course_pk PRIMARY KEY (order_id, package_id, course_id),
    CONSTRAINT order_item_course_order_item_fk FOREIGN KEY (order_id, package_id) REFERENCES "order_item"(order_id, product_id),
    CONSTRAINT order_item_course_course_id_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id)
);

CREATE POLICY rls_order_item_course ON "order_item_course" using (permission_check(resource_path, 'order_item_course')) with check (permission_check(resource_path, 'order_item_course'));

ALTER TABLE "order_item_course" ENABLE ROW LEVEL security;
ALTER TABLE "order_item_course" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.bill_item_course (
    bill_item_sequence_number integer NOT NULL,
    course_id text NOT NULL,
    course_name text NOT NULL,
    course_weight integer,
    course_slot integer,
    course_slot_per_week integer,
    created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT bill_item_course_pk PRIMARY KEY (bill_item_sequence_number, resource_path, course_id),
    CONSTRAINT bill_item_course_bill_item_fk FOREIGN KEY (bill_item_sequence_number, resource_path) REFERENCES "bill_item"(bill_item_sequence_number, resource_path),
    CONSTRAINT bill_item_course_course_id_fk FOREIGN KEY (course_id) REFERENCES "courses"(course_id)
);

CREATE POLICY rls_bill_item_course ON "bill_item_course" using (permission_check(resource_path, 'bill_item_course')) with check (permission_check(resource_path, 'bill_item_course'));

ALTER TABLE "bill_item_course" ENABLE ROW LEVEL security;
ALTER TABLE "bill_item_course" FORCE ROW LEVEL security;