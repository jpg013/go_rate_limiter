USE rate_limiter;

-- DROP ALL TABLES
DROP TABLE IF EXISTS rate_limit_request;
DROP TABLE IF EXISTS rate_limit;

-- CREATE RATE LIMIT
CREATE TABLE `rate_limit` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`type` VARCHAR(255) NOT NULL,
	`limit` INT NOT NULL,
	`reset_in_seconds` INT,
	`throttle_in_seconds` INT DEFAULT 0,
	`max_concurrent_connections` INT DEFAULT 0,
	`requests_per_window` INT DEFAULT 0,
	PRIMARY KEY (`id`)
) ENGINE=INNODB;

-- CREATE RATE LIMIT REQUEST
CREATE TABLE `rate_limit_request` (
	`token` VARCHAR(255) NOT NULL,
  `rate_limit_id` INT NOT NULL,
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`expires_at` DATETIME DEFAULT NULL,
	`in_use` BOOLEAN DEFAULT false,
	`total_usage` INT DEFAULT 0,
	PRIMARY KEY (`token`),
	INDEX (`expires_at`),
	INDEX (`in_use`),
  FOREIGN KEY (`rate_limit_id`)
		REFERENCES rate_limit(`id`)
    ON DELETE CASCADE
) ENGINE=INNODB;

INSERT INTO 
	`rate_limit` 
	(
		`name`,
		`limit`,
		`reset_in_seconds`
	)
VALUES
	(
		'twitter_friends',
		20,
		600
	);