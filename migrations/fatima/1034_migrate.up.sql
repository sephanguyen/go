ALTER TABLE public.billing_schedule_period ADD CONSTRAINT fk_billing_schedule_id FOREIGN KEY(billing_schedule_id) REFERENCES billing_schedule(id);
ALTER TABLE public.product_location ADD CONSTRAINT fk_location_id FOREIGN KEY(location_id) REFERENCES location(id);
