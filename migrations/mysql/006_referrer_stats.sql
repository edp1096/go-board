-- migrations/mysql/006_referrer_stats.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE referrer_stats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    referrer_url TEXT NOT NULL,
    target_url TEXT NOT NULL,
    visitor_ip VARCHAR(45) NOT NULL,
    user_id INT NULL,
    user_agent TEXT NULL,
    visit_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_referrer_stats_referrer ON referrer_stats(referrer_url(255));
CREATE INDEX idx_referrer_stats_target ON referrer_stats(target_url(255));
CREATE INDEX idx_referrer_stats_time ON referrer_stats(visit_time);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS referrer_stats;
-- +goose StatementEnd