CREATE TABLE `terraform_test_statistics` (
     `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
     `resource_name` varchar(128) NOT NULL DEFAULT '' COMMENT '资源名称',
     `cloud_product` varchar(2048) DEFAULT NULL COMMENT '扩展，备用',
     `code_coverage` int(3) NOT NULL COMMENT '代码覆盖率',
     `tag` varchar(32) DEFAULT NULL COMMENT '标签',
     `extension` varchar(2048) DEFAULT NULL COMMENT '扩展，备用',
     `gmt_created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
     `gmt_modified` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新日期',
     PRIMARY KEY (`id`),
     UNIQUE KEY `resource_name` (`resource_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='terraform 测试统计';
