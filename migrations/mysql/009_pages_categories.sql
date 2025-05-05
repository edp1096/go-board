-- migrations/mysql/009_pages_categories.sql
-- +goose Up
-- +goose StatementBegin
-- 페이지 테이블 생성
CREATE TABLE pages (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    content MEDIUMTEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    show_in_menu BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 카테고리 테이블 생성
CREATE TABLE categories (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    parent_id INT DEFAULT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 카테고리-아이템 관계 테이블 생성
CREATE TABLE category_items (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    category_id INT NOT NULL,
    item_id INT NOT NULL,
    item_type VARCHAR(20) NOT NULL, -- 'board' 또는 'page'
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE KEY unique_category_item (category_id, item_id, item_type)
);
-- +goose StatementEnd

-- +goose StatementBegin
-- 인덱스 생성
CREATE INDEX idx_pages_slug ON pages(slug);
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE INDEX idx_category_items_category ON category_items(category_id);
CREATE INDEX idx_category_items_item ON category_items(item_id, item_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS category_items;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS categories;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS pages;
-- +goose StatementEnd