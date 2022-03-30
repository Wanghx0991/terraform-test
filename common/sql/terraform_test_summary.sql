CREATE TABLE `terraform_test_summary` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `success_rate` int(3) NOT NULL COMMENT '测试成功率',
    `code_coverage` int(3) NOT NULL COMMENT '代码覆盖率',
    `valid_issues` int(3) NOT NULL COMMENT '有效问题数',
    `solved_issues` int(3) NOT NULL COMMENT '已解决问题数',
    `solved_rate` int(4) NOT NULL COMMENT '解决率',
    `cases` int(4) NOT NULL COMMENT 'case总数',
    `success_cases` int(4) NOT NULL COMMENT '成功case数',
    `failed_cases` int(4) NOT NULL COMMENT '失败case数',
    `skipped_cases` int(4) NOT NULL COMMENT '跳过case数',
    `verified_modules` int(3) NOT NULL COMMENT '已验证module数',
    `extension` varchar(2048) DEFAULT NULL COMMENT '扩展，备用',
    `gmt_created` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建日期',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='terraform 整体统计';

