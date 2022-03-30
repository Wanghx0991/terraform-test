CREATE TABLE `terraform_module_statistics` (
     `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
     `namespace` varchar(32) DEFAULT NULL COMMENT '命名空间',
     `module_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'module名称',
     `version` varchar(32) DEFAULT NULL COMMENT '版本',
     `verified` varchar(32) DEFAULT NULL COMMENT '是否回归',
     `tag` varchar(32) DEFAULT NULL COMMENT '标签',
     `source` varchar(256) DEFAULT NULL COMMENT '仓库地址',
     `resources` varchar(1024) DEFAULT NULL COMMENT '涉及资源',
     `examples` varchar(1024) DEFAULT NULL COMMENT '可用用例',
     `gmt_created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
     PRIMARY KEY (`id`),
     UNIQUE KEY `module_name` (`module_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='terraform-alicloud-module测试汇总';