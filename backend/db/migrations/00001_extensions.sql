-- +goose Up
-- +goose StatementBegin
create extension if not exists vectorscale cascade;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
drop extension if exists vectorscale cascade;
-- +goose StatementEnd
