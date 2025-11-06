-- +goose Up
-- +goose StatementBegin
create extension if not exists vectorscale cascade;

create table users (
    id bigserial primary key,
    email varchar(255) unique not null,
    password_hash text not null
);

create table refresh_tokens (
    id bigserial primary key,
    user_id bigint not null references users(id) on delete cascade,
    token_hash text unique not null,
    expires_at timestamptz not null
);
create index if not exists refresh_tokens_user_id_idx on refresh_tokens (user_id);

create table documents (
    id bigserial primary key,
    user_id bigint not null references users(id) on delete cascade,
    filename text not null
);
create index if not exists documents_user_id_idx on documents (user_id);

create table chunks (
    id bigserial primary key,
    user_id bigint not null references users(id) on delete cascade,
    document_id bigint not null references documents(id) on delete cascade,
    text text not null,
    embedding vector(768)
);
create index if not exists chunks_user_id_idx on chunks (user_id);
create index if not exists chunks_document_id_idx on chunks (document_id);
create index if not exists chunks_embedding_idx on chunks using diskann (embedding vector_cosine_ops) with (storage_layout='memory_optimized');
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop table if exists chunks;

drop table if exists documents;

drop table if exists refresh_tokens;

drop table if exists users;

drop extension if exists vectorscale cascade;

-- +goose StatementEnd
