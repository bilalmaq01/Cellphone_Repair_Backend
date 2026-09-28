-- 0001_init.sql — initial schema for the cellphone repair shop.
-- Table order matters: employees before repairs (FK), repairs before
-- repair_authorizations and notifications (FK).

create table if not exists customers (
    phone_number   text primary key,          -- normalized E.164, the unique customer id
    name           text not null,
    email          text,
    created_at     timestamptz not null default now(),
    updated_at     timestamptz not null default now()
);

-- Repair status enum, created idempotently.
do $$
begin
    if not exists (select 1 from pg_type where typname = 'repair_status') then
        create type repair_status as enum (
            'RECEIVED',
            'DIAGNOSING',
            'AWAITING_APPROVAL',
            'IN_PROGRESS',
            'READY_FOR_PICKUP',
            'COMPLETED',
            'CANCELLED'
        );
    end if;
end$$;

create table if not exists employees (
    id             bigint generated always as identity primary key,
    email          text not null unique,
    password_hash  text not null,
    role           text not null default 'employee' check (role in ('employee','admin')),
    active         boolean not null default true,
    created_at     timestamptz not null default now()
);

create table if not exists repairs (
    id                    bigint generated always as identity primary key,
    customer_phone        text not null references customers(phone_number),
    device                text not null,
    issue_description     text not null,
    status                repair_status not null default 'RECEIVED',
    price                 numeric(10,2),
    notes                 text,               -- internal, employee-only
    parts_used            text,
    warranty              text,
    estimated_completion  timestamptz,        -- date + time
    intake_employee_id    bigint references employees(id),
    created_at            timestamptz not null default now(),
    updated_at            timestamptz not null default now()
);

create index if not exists repairs_customer_idx on repairs(customer_phone);
create index if not exists repairs_status_idx on repairs(status);

-- Signed repair authorization captured at intake (customer signs on store device).
create table if not exists repair_authorizations (
    id             bigint generated always as identity primary key,
    repair_id      bigint not null references repairs(id),
    terms_text     text not null,             -- snapshot of terms shown at signing
    signature_url  text not null,             -- points to PNG in Supabase Storage
    signed_by_name text not null,             -- name the customer signed as
    employee_id    bigint references employees(id),
    signed_at      timestamptz not null default now()
);

create index if not exists repair_auth_repair_idx on repair_authorizations(repair_id);

create table if not exists notifications (
    id             bigint generated always as identity primary key,
    customer_phone text not null,
    repair_id      bigint references repairs(id),
    type           text not null default 'STATUS_UPDATE',
    message        text not null,
    status         text not null default 'sent',   -- 'sent' | 'failed'
    sent_at        timestamptz not null default now()
);
