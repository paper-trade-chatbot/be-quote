
-- +migrate Up
CREATE TABLE `quote_source` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `code` VARCHAR(32) NOT NULL COMMENT '報價源代號',
    `api` VARCHAR(1024) NOT NULL COMMENT 'api',
    `example` VARCHAR(1024) NOT NULL COMMENT '範例',

    PRIMARY KEY (`id`),
    UNIQUE INDEX (`code`)
) AUTO_INCREMENT=1 CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='報價源';


-- +migrate Down
SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS `quote_source`;
