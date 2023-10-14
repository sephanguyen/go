ALTER TABLE public.student_product DROP COLUMN product_associations;
ALTER TABLE public.order_item DROP COLUMN product_associations;
ALTER TABLE public.order_item ADD COLUMN student_product_id text NOT NULL;
ALTER TABLE public.bill_item ADD COLUMN student_product_id text NOT NULL;
ALTER TABLE public.order_item ADD CONSTRAINT fk_order_item_student_product_id FOREIGN KEY(student_product_id) REFERENCES public.student_product(student_product_id);
ALTER TABLE public.bill_item ADD CONSTRAINT fk_bill_item_student_product_id FOREIGN KEY(student_product_id) REFERENCES public.student_product(student_product_id);

CREATE TABLE public.associated_product (
    student_product_id text NOT NULL,
    associated_product_id text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT pk_associated_product PRIMARY KEY (student_product_id, associated_product_id),
    CONSTRAINT fk_associated_product_student_product_id FOREIGN KEY (student_product_id) REFERENCES public.student_product(student_product_id),
    CONSTRAINT fk_associated_product_associated_product_id FOREIGN KEY (associated_product_id) REFERENCES public.student_product(student_product_id)
);

CREATE POLICY rls_associated_product ON "associated_product" using (permission_check(resource_path, 'associated_product')) with check (permission_check(resource_path, 'associated_product'));

ALTER TABLE "associated_product" ENABLE ROW LEVEL security;
ALTER TABLE "associated_product" FORCE ROW LEVEL security;