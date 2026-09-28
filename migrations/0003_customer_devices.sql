alter table customers add column if not exists first_name text;

update customers
set first_name = coalesce(nullif(split_part(trim(name), ' ', 1), ''), 'Customer')
where first_name is null;

alter table customers alter column first_name set not null;

create table if not exists customer_devices (
    id bigint generated always as identity primary key,
    customer_phone text not null references customers(phone_number) on delete cascade,
    model_name text not null,
    model_key text not null,
    created_at timestamptz not null default now(),
    unique (customer_phone, model_key)
);

insert into customer_devices (customer_phone, model_name, model_key)
select distinct on (customer_phone, lower(regexp_replace(trim(device), '\s+', ' ', 'g')))
    customer_phone,
    trim(device),
    lower(regexp_replace(trim(device), '\s+', ' ', 'g'))
from repairs
where trim(device) <> ''
order by customer_phone, lower(regexp_replace(trim(device), '\s+', ' ', 'g')), created_at;

alter table repairs add column if not exists customer_device_id bigint references customer_devices(id);

update repairs r
set customer_device_id = d.id
from customer_devices d
where d.customer_phone = r.customer_phone
  and d.model_key = lower(regexp_replace(trim(r.device), '\s+', ' ', 'g'));

create index if not exists customer_devices_phone_idx on customer_devices(customer_phone);
create index if not exists repairs_customer_device_idx on repairs(customer_device_id);
