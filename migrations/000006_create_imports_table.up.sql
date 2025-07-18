CREATE TABLE IF NOT EXISTS imports (
    id bigserial primary key,
    hash text unique not null,
    filename text not null,
    organization_id bigint not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null
);