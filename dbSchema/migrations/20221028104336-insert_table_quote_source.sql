
-- +migrate Up
INSERT INTO 
	`quote_source`(
        `code`,
        `api`,
        `example`
        )
VALUES
	(
        'TWSE',
        "https://mis.twse.com.tw/stock/api/getStockInfo.jsp?json=1&delay=0&ex_ch=",
        'https://mis.twse.com.tw/stock/api/getStockInfo.jsp?json=1&delay=0&ex_ch=tse_0050.tw|'
    );

INSERT INTO 
	`product_quote_source`(
        `product_id`,
        `type`,
        `quote_code`,
        `source_code`,
        `currency_code`,
        `interval`,
        `status`
        )
VALUES
	(
        '1',
        '1',
        'tse_2330.tw',
        'TWSE',
        'NTD',
        '5',
        '1'
    );


-- +migrate Down

DELETE FROM 
    `quote_source` 
WHERE 
    `code` = 'TWSE';

DELETE FROM 
    `product_quote_source` 
WHERE 
    `product_id` = '1';