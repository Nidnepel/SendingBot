-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chat
(
    id        BIGINT  NOT NULL default 0,
    chat_key  text not null default '',
    messenger text not null default ''
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chat;
-- +goose StatementEnd