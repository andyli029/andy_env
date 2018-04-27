CREATE TABLE `mbk_orders` (
          `ORDERID` varchar(35) COLLATE utf8_bin NOT NULL COMMENT '订单ID',
          `USERID` varchar(35) COLLATE utf8_bin DEFAULT NULL,
          PRIMARY KEY (`ORDERID`),
          KEY `MBK_ORDER_USER_FK` (`USERID`),
          KEY `IX_STARTTIME` (`STARTTIME`)
);
CREATE TABLE `mbk_clothes` (
          `ORDERID` varchar(35) COLLATE utf8_bin NOT NULL COMMENT '订单ID',
          `USERID` varchar(35) COLLATE utf8_bin DEFAULT NULL,
          PRIMARY KEY (`ORDERID`),
          KEY `MBK_ORDER_USER_FK` (`USERID`),
          KEY `IX_STARTTIME` (`STARTTIME`)
);
CREATE TABLE `mbk_cars` (
          `ORDERID` varchar(35) COLLATE utf8_bin NOT NULL COMMENT '订单CAR ID',
          `USERID` varchar(35) COLLATE utf8_bin DEFAULT NULL,
          PRIMARY KEY (`ORDERID`),
          KEY `MBK_ORDER_USER_FK` (`USERID`)
);
CREATE TABLE `mbk_mobikes` (
          `ORDERID` varchar(35) COLLATE utf8_bin NOT NULL COMMENT '订单ID',
          `USERID` varchar(35) COLLATE utf8_bin DEFAULT NULL,
          PRIMARY KEY (`ORDERID`),
          KEY `MBK_ORDER_USER_FK` (`USERID`)
);
