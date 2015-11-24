-- +goose Up
ALTER TABLE `service_instances` CHANGE plan_id plan_guid VARCHAR(255) NOT NULL DEFAULT "";

-- +goose Down
ALTER TABLE `service_instances` CHANGE plan_guid plan_id VARCHAR(255) NOT NULL DEFAULT "";
