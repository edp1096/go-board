-- migrations/sqlite/001_create_tables.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    full_name TEXT,
    role TEXT NOT NULL DEFAULT 'user',
    active TINYINT NOT NULL DEFAULT 1, -- SQLite는 BOOLEAN이 없음 (1=true, 0=false)
    approval_status TEXT NOT NULL DEFAULT 'pending', -- pending, approved, rejected
    approval_due TIMESTAMP NULL,
    external_id TEXT DEFAULT NULL,
    external_system TEXT DEFAULT NULL,
    token_invalidated_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_users_external ON users (external_id, external_system);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS boards (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    board_type TEXT NOT NULL,
    table_name TEXT NOT NULL UNIQUE,
    active TINYINT NOT NULL DEFAULT 1,
    comments_enabled TINYINT NOT NULL DEFAULT 1,
    allow_anonymous TINYINT NOT NULL DEFAULT 0,
    allow_private TINYINT NOT NULL DEFAULT 0,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS board_fields (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    board_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    column_name TEXT NOT NULL,
    display_name TEXT NOT NULL,
    field_type TEXT NOT NULL,
    required TINYINT NOT NULL DEFAULT 0,
    sortable TINYINT NOT NULL DEFAULT 0,
    searchable TINYINT NOT NULL DEFAULT 0,
    options TEXT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (board_id, name),
    UNIQUE (board_id, column_name),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS board_participants (
    board_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (board_id, user_id),
    FOREIGN KEY (board_id) REFERENCES boards(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS board_fields;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS boards;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd