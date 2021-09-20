CREATE DATABASE IF NOT EXISTS `fiskil`;
USE `fiskil`;
SET time_zone = '+00:00';

CREATE TABLE IF NOT EXISTS `service_logs` (
   `service_name` VARCHAR(100) NOT NULL,
   `payload` VARCHAR(2048) NOT NULL,
   `severity` ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
   `timestamp` DATETIME NOT NULL,
   `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    INDEX index_service_logs_service_name (`service_name`),
    INDEX index_service_logs_severity (`severity`),
    INDEX index_service_logs_timestamp (`timestamp`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8 ROW_FORMAT=FIXED;

CREATE TABLE IF NOT EXISTS `service_severity` (
   `service_name` VARCHAR(100) NOT NULL,
   `severity` ENUM("debug", "info", "warn", "error", "fatal") NOT NULL,
   `count` INT(4) NOT NULL,
   `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT pk_service_severity PRIMARY KEY (`service_name`, `severity`)
) ENGINE=MyISAM DEFAULT CHARSET=utf8 ROW_FORMAT=FIXED;

CREATE PROCEDURE count_severity()
BEGIN

    INSERT INTO `service_severity`
    SELECT `service_name`, `severity`, COUNT(`severity`) AS `count`, NOW() AS `created_at`
    FROM `service_logs`
    GROUP BY `service_name`, `severity`
    ON DUPLICATE KEY UPDATE `count`=VALUES(`count`), `created_at`= NOW();

END

CREATE EVENT schd_count_severity
    ON SCHEDULE EVERY 1 SECOND
    DO
      CALL count_severity();
