DROP TABLE IF EXISTS `friend`;

CREATE TABLE `friend` (
                          `id` int(11) NOT NULL AUTO_INCREMENT,
                          `from_id` int(11) NOT NULL,
                          `to_id` int(11) NOT NULL,
                          `created_at` datetime DEFAULT NULL,
                          PRIMARY KEY (`id`,`from_id`),
                          UNIQUE KEY `from_id` (`from_id`,`to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `log`;

CREATE TABLE `log` (
                       `id` int(11) NOT NULL AUTO_INCREMENT,
                       `user_id` int(11) NOT NULL,
                       `action` enum('add_friend','del_friend','add_subscription','del_subscription') DEFAULT NULL,
                       `to_id` int(11) NOT NULL,
                       `created_at` datetime DEFAULT NULL,
                       PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `subscription`;

CREATE TABLE `subscription` (
                                `id` int(11) NOT NULL AUTO_INCREMENT,
                                `from_id` int(11) NOT NULL,
                                `to_id` int(11) NOT NULL,
                                `created_at` datetime DEFAULT NULL,
                                PRIMARY KEY (`id`,`from_id`),
                                UNIQUE KEY `from_id` (`from_id`,`to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
                        `id` int(11) NOT NULL AUTO_INCREMENT,
                        `name` varchar(255) DEFAULT NULL,
                        `created_at` datetime DEFAULT NULL,
                        PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
