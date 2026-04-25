-- Adminer 5.4.2 MariaDB 11.8.6-MariaDB-ubu2404 dump

SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `accounts`;
CREATE TABLE `accounts` (
  `login` varchar(45) NOT NULL DEFAULT '',
  `password` varchar(500) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `created_time` timestamp NOT NULL DEFAULT current_timestamp(),
  `lastactive` timestamp NULL DEFAULT NULL,
  `accessLevel` tinyint(4) NOT NULL DEFAULT 0,
  `lastIP` char(15) DEFAULT NULL,
  `lastServer` tinyint(4) DEFAULT 1,
  PRIMARY KEY (`login`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

INSERT INTO `accounts` (`login`, `password`, `email`, `created_time`, `lastactive`, `accessLevel`, `lastIP`, `lastServer`) VALUES
('admin',	'8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918',	NULL,	'2026-01-01 00:00:00',	NULL,	0,	NULL,	1);

DROP TABLE IF EXISTS `bbs`;
CREATE TABLE `bbs` (
  `id` int(10) unsigned NOT NULL DEFAULT 0,
  `receiver_id` int(10) unsigned NOT NULL DEFAULT 0,
  `sender_id` int(10) unsigned NOT NULL DEFAULT 0,
  `subject` varchar(128) DEFAULT NULL,
  `message` varchar(3000) DEFAULT NULL,
  `sent_date` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

INSERT INTO `bbs` (`id`, `receiver_id`, `sender_id`, `subject`, `message`, `sent_date`) VALUES
(1,	0,	0,	'Ni hao!',	'Hello my dear firend, welcome to Dark Ages!',	NULL);

DROP TABLE IF EXISTS `characters`;
CREATE TABLE `characters` (
  `account_name` varchar(45) DEFAULT NULL,
  `charId` int(10) unsigned NOT NULL DEFAULT 0,
  `char_name` varchar(35) NOT NULL,
  `level` tinyint(3) unsigned DEFAULT NULL,
  `maxHp` mediumint(8) unsigned DEFAULT NULL,
  `curHp` mediumint(8) unsigned DEFAULT NULL,
  `maxCp` mediumint(8) unsigned DEFAULT NULL,
  `curCp` mediumint(8) unsigned DEFAULT NULL,
  `maxMp` mediumint(8) unsigned DEFAULT NULL,
  `curMp` mediumint(8) unsigned DEFAULT NULL,
  `face` tinyint(3) unsigned DEFAULT NULL,
  `hairStyle` tinyint(3) unsigned DEFAULT NULL,
  `hairColor` tinyint(3) unsigned DEFAULT NULL,
  `sex` tinyint(3) unsigned DEFAULT NULL,
  `heading` mediumint(9) DEFAULT NULL,
  `x` mediumint(9) DEFAULT NULL,
  `y` mediumint(9) DEFAULT NULL,
  `z` mediumint(9) DEFAULT NULL,
  `exp` bigint(20) unsigned DEFAULT 0,
  `expBeforeDeath` bigint(20) unsigned DEFAULT 0,
  `sp` bigint(10) unsigned NOT NULL DEFAULT 0,
  `karma` int(10) unsigned DEFAULT NULL,
  `pvpkills` smallint(5) unsigned DEFAULT NULL,
  `pkkills` smallint(5) unsigned DEFAULT NULL,
  `clanid` int(10) unsigned DEFAULT NULL,
  `race` tinyint(3) unsigned DEFAULT NULL,
  `classid` tinyint(3) unsigned DEFAULT NULL,
  `base_class` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `deletetime` bigint(13) unsigned NOT NULL DEFAULT 0,
  `cancraft` tinyint(3) unsigned DEFAULT NULL,
  `title` varchar(21) DEFAULT NULL,
  `title_color` mediumint(8) unsigned NOT NULL DEFAULT 15530402,
  `accesslevel` mediumint(9) DEFAULT 0,
  `online` tinyint(3) unsigned DEFAULT NULL,
  `onlinetime` int(11) DEFAULT NULL,
  `char_slot` tinyint(3) unsigned DEFAULT NULL,
  `lastAccess` bigint(13) unsigned NOT NULL DEFAULT 0,
  `clan_privs` int(10) unsigned DEFAULT 0,
  `power_grade` tinyint(3) unsigned DEFAULT NULL,
  `nobless` tinyint(3) unsigned NOT NULL DEFAULT 0,
  `clan_join_expiry_time` bigint(13) unsigned NOT NULL DEFAULT 0,
  `clan_create_expiry_time` bigint(13) unsigned NOT NULL DEFAULT 0,
  `death_penalty_level` smallint(5) unsigned NOT NULL DEFAULT 0,
  `createDate` date NOT NULL DEFAULT '2015-01-01',
  `language` varchar(2) DEFAULT NULL,
  `newbie` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`charId`),
  KEY `account_name` (`account_name`),
  KEY `char_name` (`char_name`),
  KEY `clanid` (`clanid`),
  KEY `online` (`online`),
  KEY `idx_charId` (`charId`),
  KEY `idx_char_name` (`char_name`),
  KEY `idx_account_name` (`account_name`),
  KEY `idx_accountName_createDate` (`account_name`,`createDate`),
  KEY `idx_createDate` (`createDate`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;


SET NAMES utf8mb4;

DROP TABLE IF EXISTS `char_initial_items`;
CREATE TABLE `char_initial_items` (
  `classId` tinyint(3) unsigned DEFAULT NULL,
  `itemId` int(11) DEFAULT NULL,
  `amount` int(11) DEFAULT 1,
  `equipped` tinyint(1) DEFAULT 0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

INSERT INTO `char_initial_items` (`classId`, `itemId`, `amount`, `equipped`) VALUES
(0,	10,	1,	0),
(0,	1146,	1,	1),
(0,	1147,	1,	1),
(0,	2369,	1,	1),
(10,	6,	1,	1),
(10,	425,	1,	1),
(10,	461,	1,	1),
(18,	10,	1,	0),
(18,	1146,	1,	1),
(18,	1147,	1,	1),
(18,	2369,	1,	1),
(25,	6,	1,	1),
(25,	425,	1,	1),
(25,	461,	1,	1),
(31,	10,	1,	0),
(31,	1146,	1,	1),
(31,	1147,	1,	1),
(31,	2369,	1,	1),
(38,	6,	1,	1),
(38,	425,	1,	1),
(38,	461,	1,	1),
(44,	1146,	1,	1),
(44,	1147,	1,	1),
(44,	2369,	1,	0),
(44,	2368,	1,	1),
(49,	425,	1,	1),
(49,	461,	1,	1),
(49,	2368,	1,	1),
(53,	1146,	1,	1),
(53,	1147,	1,	1),
(53,	2369,	1,	0),
(53,	2370,	1,	1);

DROP TABLE IF EXISTS `char_templates`;
CREATE TABLE `char_templates` (
  `classId` tinyint(3) unsigned NOT NULL,
  `className` varchar(50) DEFAULT NULL,
  `raceId` tinyint(3) unsigned DEFAULT NULL,
  `startX` mediumint(9) DEFAULT NULL,
  `startY` mediumint(9) DEFAULT NULL,
  `startZ` mediumint(9) DEFAULT NULL,
  `str` tinyint(3) unsigned DEFAULT NULL,
  `dex` tinyint(3) unsigned DEFAULT NULL,
  `con` tinyint(3) unsigned DEFAULT NULL,
  `int` tinyint(3) unsigned DEFAULT NULL,
  `wit` tinyint(3) unsigned DEFAULT NULL,
  `men` tinyint(3) unsigned DEFAULT NULL,
  `hp` smallint(5) unsigned DEFAULT NULL,
  `mp` smallint(5) unsigned DEFAULT NULL,
  PRIMARY KEY (`classId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;

INSERT INTO `char_templates` (`classId`, `className`, `raceId`, `startX`, `startY`, `startZ`, `str`, `dex`, `con`, `int`, `wit`, `men`, `hp`, `mp`) VALUES
(0,	'Human Fighter',	0,	-70880,	257360,	-3080,	40,	30,	43,	21,	11,	25,	80,	30),
(10,	'Human Mystic',	0,	-90256,	249792,	-3570,	22,	21,	27,	41,	20,	39,	101,	40),
(18,	'Elven Fighter',	1,	46931,	51433,	-2977,	36,	35,	36,	23,	14,	26,	89,	30),
(25,	'Elven Mystic',	1,	46931,	51433,	-2977,	21,	24,	25,	37,	23,	40,	104,	40),
(31,	'Dark Elven Fighter',	2,	26960,	10496,	-4220,	41,	34,	32,	25,	12,	26,	94,	30),
(38,	'Dark Elven Mystic',	2,	26960,	10496,	-4220,	23,	23,	24,	44,	19,	37,	106,	40),
(44,	'Orc Fighter',	3,	-58192,	-113408,	-650,	40,	26,	47,	18,	12,	27,	80,	30),
(49,	'Orc Mystic',	3,	-58192,	-113408,	-650,	27,	24,	31,	31,	15,	42,	95,	40),
(53,	'Dwarven Fighter',	4,	115200,	-178000,	-950,	39,	29,	45,	20,	10,	27,	80,	30);

DROP TABLE IF EXISTS `gameservers`;
CREATE TABLE `gameservers` (
  `server_id` int(11) NOT NULL DEFAULT 0,
  `host` varchar(50) NOT NULL DEFAULT '',
  `port` varchar(50) NOT NULL DEFAULT '',
  PRIMARY KEY (`server_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;

INSERT INTO `gameservers` (`server_id`, `host`, `port`) VALUES
(1,	'127.0.0.1',	'7777');

DROP TABLE IF EXISTS `items`;
CREATE TABLE `items` (
  `owner_id` int(11) DEFAULT NULL,
  `object_id` int(11) NOT NULL DEFAULT 0,
  `item_id` int(11) DEFAULT NULL,
  `count` bigint(20) unsigned NOT NULL DEFAULT 0,
  `enchant_level` int(11) DEFAULT NULL,
  `loc` varchar(10) DEFAULT NULL,
  `loc_data` int(11) DEFAULT NULL,
  `time_of_use` int(11) DEFAULT NULL,
  `custom_type1` int(11) DEFAULT 0,
  `custom_type2` int(11) DEFAULT 0,
  `mana_left` decimal(5,0) NOT NULL DEFAULT -1,
  `time` decimal(13,0) NOT NULL DEFAULT 0,
  PRIMARY KEY (`object_id`),
  KEY `owner_id` (`owner_id`),
  KEY `item_id` (`item_id`),
  KEY `loc` (`loc`),
  KEY `time_of_use` (`time_of_use`),
  KEY `idx_item_id` (`item_id`),
  KEY `idx_object_id` (`object_id`),
  KEY `idx_owner_id` (`owner_id`),
  KEY `idx_owner_id_loc` (`owner_id`,`loc`),
  KEY `idx_owner_id_item_id` (`owner_id`,`item_id`),
  KEY `idx_owner_id_loc_locdata` (`owner_id`,`loc`,`loc_data`),
  KEY `idx_owner_id_loc_locdata_enchant` (`owner_id`,`loc`,`loc_data`,`enchant_level`,`item_id`,`object_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb3 COLLATE=utf8mb3_unicode_ci;


DROP TABLE IF EXISTS `object_id_registry`;
CREATE TABLE `object_id_registry` (
  `obj_id` int(11) NOT NULL AUTO_INCREMENT,
  PRIMARY KEY (`obj_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;


-- 2026-04-25 18:43:26 UTC
