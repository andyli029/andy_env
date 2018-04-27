CREATE TABLE `mbk_a` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` varchar(35) COLLATE utf8_bin DEFAULT NULL COMMENT '用户ID',
  `user_name` varchar(50) COLLATE utf8_bin DEFAULT NULL COMMENT '用户名称',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_userid` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `mbk_b` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` varchar(35) COLLATE utf8_bin DEFAULT NULL COMMENT '用户ID',
    `user_name` varchar(50) COLLATE utf8_bin DEFAULT NULL COMMENT '用户名称',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_userid` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE `mbk_c` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` varchar(35) COLLATE utf8_bin DEFAULT NULL COMMENT '用户ID',
    `user_name` varchar(50) COLLATE utf8_bin DEFAULT NULL COMMENT '用户名称',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_userid` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

