-- +goose Up
ALTER TABLE `scaling_decisions` ADD decision_type int(11) NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE `scaling_decisions` DROP COLUMN decision_type;
