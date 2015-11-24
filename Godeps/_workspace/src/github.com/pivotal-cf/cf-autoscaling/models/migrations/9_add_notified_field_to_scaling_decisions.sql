-- +goose Up
ALTER TABLE `scaling_decisions` ADD notified TINYINT(1) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE `scaling_decisions` DROP COLUMN notified;
