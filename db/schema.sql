CREATE TABLE `vtubers` (
    `id` varchar(24) NOT NULL, 
    `name` VARCHAR(255) NOT NULL, 
    `item_count` integer DEFAULT 0, 
    `created_at` DATETIME NOT NULL DEFAULT current_timestamp(),
    `updated_at` DATETIME NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(), PRIMARY KEY (`id`)
);

CREATE TABLE `videos` (
    `id` varchar(11) NOT NULL, 
    `title` VARCHAR(255) NOT NULL, 
    `duration` VARCHAR(255) NOT NULL, 
    `viewers` integer NOT NULL,
    `content` VARCHAR(255) NOT NULL,
    `announced` boolean DEFAULT false,
    `scheduled_start_time` timestamp, 
    `created_at` DATETIME NOT NULL DEFAULT current_timestamp(), 
    `updated_at` DATETIME NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(), PRIMARY KEY (`id`)
);