CREATE TABLE IF NOT EXISTS fund_categories (
    id bigserial primary key,
    name text not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null
);