CREATE TABLE IF NOT EXISTS `users`
(
    id bigint auto_increment,
    firstname varchar(255) NOT NULL,
    lastname varchar(255) NOT NULL,
    PRIMARY KEY (`id`)
);