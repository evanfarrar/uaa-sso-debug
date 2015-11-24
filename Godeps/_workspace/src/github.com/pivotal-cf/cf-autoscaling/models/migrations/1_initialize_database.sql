-- +goose Up
CREATE TABLE IF NOT EXISTS `readings` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `service_binding_guid` varchar(255) NOT NULL DEFAULT "",
    `cpu_utilization` int(11) NOT NULL DEFAULT 0,
    `expected_instance_count` int(11) NOT NULL DEFAULT 0,
    `running_instance_count` int(11) NOT NULL DEFAULT 0,
    `time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `state` varchar(255) NOT NULL DEFAULT "",
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `scaling_decisions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `reading_id` int(11) NOT NULL DEFAULT 0,
  `scaling_factor` int(11) NOT NULL DEFAULT 0,
  `description` varchar(255) NOT NULL DEFAULT "",
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `service_bindings` (
  `guid` varchar(255) NOT NULL,
  `service_instance_guid` varchar(255) NOT NULL DEFAULT "",
  `app_guid` varchar(255) NOT NULL DEFAULT "",
  `app_name` varchar(255) NOT NULL DEFAULT "",
  `expected_instance_count` int(11) NOT NULL DEFAULT 0,
  `min_instances` int(11) NOT NULL DEFAULT 2,
  `max_instances` int(11) NOT NULL DEFAULT 5,
  `cpu_min_threshold` int(11) NOT NULL DEFAULT 20,
  `cpu_max_threshold` int(11) NOT NULL DEFAULT 80,
  `enabled` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `service_instances` (
  `guid` varchar(255) NOT NULL,
  `plan_id` varchar(255) NOT NULL DEFAULT "",
  PRIMARY KEY (`guid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
DROP TABLE readings;
DROP TABLE scaling_decisions;
DROP TABLE service_bindings;
DROP TABLE service_instances;
