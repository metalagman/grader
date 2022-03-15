-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS "assessments"
(
    id              UUID                  DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    part_id         VARCHAR(255) NOT NULL UNIQUE,
    container_image VARCHAR(255) NOT NULL,
    summary         TEXT         NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "assessments";
-- +goose StatementEnd
