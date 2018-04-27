#!/bin/bash
a="CREATE TABLE mbk_modou_total_v2_"
b='(`id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',`user_id` varchar(35) COLLATE utf8_bin DEFAULT NULL COMMENT '用户ID');'
for i in {0..1}; do echo -n $a; printf "%04d " ${i}; echo $b; done;

