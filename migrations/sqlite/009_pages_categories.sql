-- migrations/sqlite/009_pages_categories.sql
-- +goose Up
-- 페이지 테이블 생성
CREATE TABLE pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    content TEXT,
    active TINYINT NOT NULL DEFAULT 1,
    show_in_menu TINYINT NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 카테고리 테이블 생성
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    parent_id INTEGER DEFAULT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    active TINYINT NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- 카테고리-아이템 관계 테이블 생성
CREATE TABLE category_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    item_type TEXT NOT NULL, -- 'board' 또는 'page'
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE (category_id, item_id, item_type)
);

-- 인덱스 생성
CREATE INDEX idx_pages_slug ON pages(slug);
CREATE INDEX idx_categories_slug ON categories(slug);
CREATE INDEX idx_categories_parent ON categories(parent_id);
CREATE INDEX idx_category_items_category ON category_items(category_id);
CREATE INDEX idx_category_items_item ON category_items(item_id, item_type);

-- +goose Down
DROP TABLE IF EXISTS category_items;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS pages;