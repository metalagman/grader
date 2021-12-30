-- +goose Up
-- +goose StatementBegin
INSERT INTO users (name, password, is_admin)
VALUES ("graderadmin", crypt("graderpass", gen_salt('bf')), true)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE name="graderadmin"
-- +goose StatementEnd
