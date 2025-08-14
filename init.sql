CREATE DATABASE IF NOT EXISTS `file_engine` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE `file_engine`;

CREATE TABLE IF NOT EXISTS `t_file` (
    `id` VARCHAR(40) NOT NULL,
    `name` VARCHAR(255) NOT NULL COMMENT '文件名',
    `content_type` VARCHAR(255) NOT NULL COMMENT '文件类型(application/octet-stream)',
    `bucket_id` VARCHAR(40) NOT NULL COMMENT '桶ID',
    `size` BIGINT(20) NOT NULL COMMENT '文件大小',
    `icon` VARCHAR(255) NOT NULL COMMENT '文件图标',
    `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_name` (`name`)
) ENGINE=InnoDB COMMENT='文件表';

