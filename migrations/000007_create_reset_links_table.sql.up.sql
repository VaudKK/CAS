CREATE TABLE IF NOT EXISTS reset_links (
    id bigserial primary key,
    email text not null,
    reset_token text unique not null,
    expiry timestamp with time zone not null,
    is_used boolean default false not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null,
    created_by varchar(1000) null,
    modified_by varchar(1000) null
);