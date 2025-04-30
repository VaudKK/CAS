CREATE TABLE IF NOT EXISTS otp (
    id bigserial primary key,
    subject text not null,
    verification_mode varchar(200) not null,
    otp varchar(200) not null,
    expiry timestamp with time zone not null,
    used boolean not null,
    session_id text unique not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null
);