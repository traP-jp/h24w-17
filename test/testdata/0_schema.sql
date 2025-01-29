CREATE TABLE `users` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `age` INT NOT NULL,
    `group_id` INT,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`)
);
