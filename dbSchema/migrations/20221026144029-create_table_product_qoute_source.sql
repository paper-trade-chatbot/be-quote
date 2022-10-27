
-- +migrate Up
CREATE TABLE `product_quote_source` (
    `product_id` BIGINT UNSIGNED NOT NULL COMMENT '產品id',
    `type` TINYINT(4) NOT NULL COMMENT '產品種類 1:stock, 2:crypto, 3:forex, 4:futures',
    `quote_code` VARCHAR(32) NOT NULL COMMENT '報價代號',
    `source_code` VARCHAR(32) NOT NULL COMMENT '報價源代號',
    `currency_code` VARCHAR(32) NOT NULL COMMENT '貨幣代號',
    `interval` VARCHAR(6) NOT NULL COMMENT '報價間隔 HHMMSS',
    `status` TINYINT(4) NOT NULL COMMENT '狀態 1:enabled, 2:disabled',

    PRIMARY KEY (`product_id`),
    UNIQUE INDEX (`source_code`, `quote_code`),
    FOREIGN KEY (`source_code`) REFERENCES quote_source(`code`) ON DELETE CASCADE
) AUTO_INCREMENT=1 CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='產品報價源';


-- +migrate Down
SET FOREIGN_KEY_CHECKS=0;
DROP TABLE IF EXISTS `product_quote_source`;
