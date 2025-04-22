CREATE TABLE IF NOT EXISTS users (
    id bigserial primary key,
    username text not null,
    email text unique not null,
    password varchar(200) not null,
    organization_id bigint not null,
    verified boolean null,
    active boolean null,
    last_login timestamp with time zone null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null
);