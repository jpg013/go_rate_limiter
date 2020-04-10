USE rate_limiter;

-- DROP ALL TABLES
DROP TABLE IF EXISTS rate_limit_resource;
DROP TABLE IF EXISTS rate_limit;

-- CREATE RATE LIMIT
CREATE TABLE `rate_limit` (
	`id` INT NOT NULL AUTO_INCREMENT,
	`name` VARCHAR(255) NOT NULL,
	`limit` INT NOT NULL,
	`reset_in_seconds` INT,
	PRIMARY KEY (`id`)
) ENGINE=INNODB;

-- CREATE RATE LIMIT RESOURCE
CREATE TABLE `rate_limit_resource` (
	`id` INT NOT NULL AUTO_INCREMENT,
  `rate_limit_id` INT NOT NULL,
	`created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	`expires_at` DATETIME NOT NULL,
	`is_expired` BOOLEAN DEFAULT false,
  `count` INT NOT NULL,
	PRIMARY KEY (`id`),
	INDEX (`expires_at`),
	INDEX (`is_expired`),
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
		'crawlera',
		20,
		600
	);