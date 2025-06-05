CREATE TABLE IF NOT EXISTS fund_categories (
    id bigserial primary key,
    name text not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null,
    created_by varchar(1000) null,
    modified_by varchar(1000) null
);