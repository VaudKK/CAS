CREATE TABLE IF NOT EXISTS funds (
    id bigserial primary key,
    contributor text not null,
    total numeric(20, 2) not null,
    receipt_no varchar(200) not null,
    organization_id bigint not null,
    break_down jsonb not null,
    contribution_date date not null,
    created_at timestamp with time zone default now() not null,
    modified_at timestamp with time zone default now() not null,
    created_by varchar(1000) null,
    modified_by varchar(1000) null
);

CREATE INDEX IF NOT EXISTS funds_contributor_idx ON funds USING GIN (to_tsvector('simple', contributor));