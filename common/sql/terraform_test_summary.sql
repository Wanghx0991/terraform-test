CREATE TABLE `terraform_test_summary` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `success_rate` int(3) NOT NULL COMMENT '测试成功率',
    `code_coverage` int(3) NOT NULL COMMENT '代码覆盖率',
    `extension` varchar(2048) DEFAULT NULL COMMENT '扩展，备用',
    `gmt_created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='terraform 整体统计';

