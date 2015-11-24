-- +goose Up
CREATE TABLE IF NOT EXISTS `key_value_store` (
    `key` varchar(255) NOT NULL DEFAULT "",
    `value` text NOT NULL,
    PRIMARY KEY (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
DROP TABLE key_value_store;
