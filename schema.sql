CREATE TABLE `videos` (
	`id` char(11) NOT NULL,
	`title` varchar(255),
	`songConfirm` tinyint unsigned DEFAULT '0',
	`scheduled_start_time` datetime,
	`twitter_id` char(19),
	`created_at` datetime NOT NULL DEFAULT current_timestamp(),
	`updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	PRIMARY KEY (`id`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;

CREATE TABLE `vtubers` (
	`id` char(24) NOT NULL,
	`name` varchar(255) NOT NULL,
	`item_count` int unsigned NOT NULL DEFAULT '0',
	`created_at` datetime NOT NULL DEFAULT current_timestamp(),
	`updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
	PRIMARY KEY (`id`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;