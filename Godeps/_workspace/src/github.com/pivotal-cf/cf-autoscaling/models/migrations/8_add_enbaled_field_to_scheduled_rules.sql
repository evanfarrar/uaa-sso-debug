-- +goose Up
ALTER TABLE `scheduled_rules` ADD enabled TINYINT(1) NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE `scheduled_rules` DROP COLUMN enabled;
