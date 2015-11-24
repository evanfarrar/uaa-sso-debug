-- +goose Up
CREATE TABLE IF NOT EXISTS `scheduled_rules` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `min_instances` int(11) NOT NULL DEFAULT 0,
    `max_instances` int(11) NOT NULL DEFAULT 0,
    `executes_at` timestamp NULL,
    `service_binding_guid` varchar(255) NOT NULL DEFAULT "",
    `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
DROP TABLE scheduled_rules;
