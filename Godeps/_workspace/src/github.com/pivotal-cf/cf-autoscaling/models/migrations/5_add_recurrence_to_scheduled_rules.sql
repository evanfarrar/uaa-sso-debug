-- +goose Up
ALTER TABLE `scheduled_rules` ADD recurrence int(3) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE `scheduled_rules` DROP COLUMN recurrence;
