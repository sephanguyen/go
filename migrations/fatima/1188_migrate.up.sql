create index if not exists order_item__order_id__idx on public.order_item(order_id);
create index if not exists order_item__product_id__idx on public.order_item(product_id);
