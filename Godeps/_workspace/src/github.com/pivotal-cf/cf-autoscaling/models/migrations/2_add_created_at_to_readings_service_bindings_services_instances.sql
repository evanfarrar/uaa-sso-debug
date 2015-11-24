-- +goose Up
ALTER TABLE `readings` DROP COLUMN time;
ALTER TABLE `readings` ADD created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `service_bindings` ADD created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;
ALTER TABLE `service_instances` ADD created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- +goose Down
ALTER TABLE readings DROP COLUMN created_at;
ALTER TABLE service_bindings DROP COLUMN created_at;
ALTER TABLE service_instances DROP COLUMN created_at;
