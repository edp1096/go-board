-- migrations/sqlite/006_referrer_stats.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE referrer_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    referrer_url TEXT NOT NULL,
    referrer_domain TEXT NOT NULL,
    referrer_type TEXT NOT NULL,
    target_url TEXT NOT NULL,
    visitor_ip TEXT NOT NULL,
    user_id INTEGER NULL,
    user_agent TEXT NULL,
    visit_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url);
CREATE INDEX idx_referrer_stats_domain ON referrer_stats(referrer_domain);
CREATE INDEX idx_referrer_stats_type ON referrer_stats(referrer_type);
CREATE INDEX idx_referrer_stats_target ON referrer_stats(target_url);
CREATE INDEX idx_referrer_stats_time ON referrer_stats(visit_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS referrer_stats;
-- +goose StatementEnd