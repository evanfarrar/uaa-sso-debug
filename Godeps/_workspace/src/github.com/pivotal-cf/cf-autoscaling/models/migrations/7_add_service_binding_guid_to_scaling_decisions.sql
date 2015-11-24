-- +goose Up
ALTER TABLE `scaling_decisions` ADD service_binding_guid varchar(255) NOT NULL DEFAULT "";

-- +goose Down
ALTER TABLE `scaling_decisions` DROP COLUMN scaling_decsions;
