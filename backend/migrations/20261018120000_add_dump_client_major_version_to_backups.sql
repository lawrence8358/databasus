-- +goose Up
-- +goose StatementBegin

ALTER TABLE logical_backups
    ADD COLUMN dump_client_major_version INT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE logical_backups
    DROP COLUMN IF EXISTS dump_client_major_version;

-- +goose StatementEnd
