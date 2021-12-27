-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS "submissions"
(
    id            UUID                   DEFAULT uuid_generate_v4() NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    user_id       UUID          NOT NULL,
    assessment_id UUID          NOT NULL,
    file_name     VARCHAR(255)  NOT NULL,
    file_url      VARCHAR(2048) NOT NULL,
    external_id   TEXT,
    result_date   TIMESTAMPTZ,
    result_pass   BOOLEAN,
    result_text   TEXT,
    PRIMARY KEY (id),
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id),
    CONSTRAINT fk_assessment
        FOREIGN KEY (assessment_id)
            REFERENCES assessments (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "submissions";
-- +goose StatementEnd

