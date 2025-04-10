CREATE TABLE IF NOT EXISTS organizations(
    id bigserial primary key,
    organization_name text not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null
);

INSERT INTO organizations (organization_name) VALUES ('Kitengela Central SDA');