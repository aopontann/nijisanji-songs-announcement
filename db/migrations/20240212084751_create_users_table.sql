-- Create "users" table
CREATE TABLE `users` (`token` varchar(200) NOT NULL, `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP, `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, PRIMARY KEY (`token`)) CHARSET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
