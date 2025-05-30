-- +goose Up
CREATE TABLE smtp (
    `id` INTEGER PRIMARY KEY AUTOINCREMENT,
    `account` INTEGER NOT NULL UNIQUE,
    `code` TEXT NOT NULL, -- Hashed
    `expires` TIMESTAMP NOT NULL, -- Expiration time
    `created` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (account) REFERENCES account(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE smtp;