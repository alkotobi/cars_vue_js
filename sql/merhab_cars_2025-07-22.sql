# ************************************************************
# Sequel Ace SQL dump
# Version 20095
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# Host: 173.214.163.18 (MySQL 11.4.7-MariaDB)
# Database: merhab_cars
# Generation Time: 2025-07-22 02:41:34 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


# Dump of table adv_sql
# ------------------------------------------------------------

CREATE TABLE `adv_sql` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `stmt` text DEFAULT NULL,
  `desc` varchar(255) DEFAULT NULL,
  `params` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`params`)),
  `param_values` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`param_values`)),
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `adv_sql` WRITE;
/*!40000 ALTER TABLE `adv_sql` DISABLE KEYS */;

INSERT INTO `adv_sql` (`id`, `stmt`, `desc`, `params`, `param_values`, `name`)
VALUES
	(1,'select * from users','test',X'5B5D',X'7B7D','test'),
	(2,'SELECT \n    cs.id as stock_id,\n    cs.vin,\n    cs.notes as stock_notes,\n    cs.date_sell,\n    cs.price_cell as sell_price,\n    cs.freight,\n    cs.date_loding as loading_date,\n    cs.date_send_documents,\n    cs.in_wharhouse_date,\n    cs.date_get_documents_from_supp,\n    cs.date_get_keys_from_supp,\n    cs.date_get_bl,\n    cs.date_pay_freight,\n    cs.container_ref,\n    cs.export_lisence_ref,\n    cs.is_used_car,\n    cs.is_big_car,\n    cs.is_tmp_client,\n    \n    -- Buy Bill Info\n    bb.id as buy_bill_id,\n    bb.bill_ref,\n    bb.date_buy,\n    bb.amount as buy_amount,\n    bb.payed as buy_paid,\n    bb.is_stock_updated,\n    bb.is_ordered,\n    \n    -- Buy Details Info\n    bd.id as buy_detail_id,\n    bd.amount as buy_detail_amount,\n    bd.QTY,\n    bd.year,\n    bd.month,\n    bd.price_sell as buy_price_sell,\n    bd.notes as buy_detail_notes,\n    \n    -- Car Info\n    cn.car_name,\n    c.color,\n    b.brand,\n    \n    -- Supplier Info\n    s.name as supplier_name,\n    s.contact_info as supplier_address,\n    \n    -- Ports Info\n    lp.loading_port,\n    dp.discharge_port,\n    \n    -- Warehouse Info\n\n\n    \n    -- Client Info (if sold)\n    cl.name as client_name,\n    cl.mobiles as client_mobile,\n    cl.email as client_email,\n    \n    -- Rate Info\n    cs.rate\n    \nFROM cars_stock cs\nLEFT JOIN buy_details bd ON cs.id_buy_details = bd.id\nLEFT JOIN buy_bill bb ON bd.id_buy_bill = bb.id\nLEFT JOIN cars_names cn ON bd.id_car_name = cn.id\nLEFT JOIN colors c ON bd.id_color = c.id\nLEFT JOIN brands b ON cn.id_brand = b.id\nLEFT JOIN suppliers s ON bb.id_supplier = s.id\nLEFT JOIN loading_ports lp ON cs.id_port_loading = lp.id\nLEFT JOIN discharge_ports dp ON cs.id_port_discharge = dp.id\nLEFT JOIN warehouses w ON cs.id_warehouse = w.id\nLEFT JOIN clients cl ON cs.id_client = cl.id\nWHERE bb.bill_ref = :bill_ref\nORDER BY cs.id DESC','get stock elements of a givven buy ref',X'5B2262696C6C5F726566225D',X'7B2262696C6C5F726566223A2252572D4D4D32303235303331382D3130227D','stock_buy_ref'),
	(3,'-- Update cars_stock table with calculated dates based on sell_bill date_sell\n-- This script updates records that have id_sell (sold cars)\n\nUPDATE cars_stock cs\nJOIN sell_bill sb ON cs.id_sell = sb.id\nSET \n    cs.in_wharhouse_date = DATE_SUB(sb.date_sell, INTERVAL 7 DAY),\n    cs.date_get_documents_from_supp = DATE_ADD(sb.date_sell, INTERVAL 10 DAY),\n    cs.date_pay_freight = DATE_ADD(sb.date_sell, INTERVAL 30 DAY),\n    cs.date_get_keys_from_supp = DATE_SUB(sb.date_sell, INTERVAL 7 DAY),\n    cs.date_get_bl = DATE_ADD(sb.date_sell, INTERVAL 40 DAY)\nWHERE cs.id_sell IS NOT NULL;','Update cars_stock table with calculated dates based on sell_bill date_sell',X'5B5D',X'7B7D','fake cars dates '),
	(4,'UPDATE cars_stock \nSET vin = null, container_ref =null\nWHERE vin = :vin','',X'5B2276696E225D',X'7B2276696E223A224C4C563243334232355330303031353034227D','clear vin and conteiner ref'),
	(5,'-- SQL Script to Remove Purchase Bill and Associated Cars\n-- This script will:\n-- 1. Delete cars from cars_stock that are associated with the purchase bill\n-- 2. Delete buy_details records associated with the purchase bill\n-- 3. Delete buy_payments records associated with the purchase bill\n-- 4. Delete the purchase bill itself\n\n-- Replace @BUY_BILL_ID with the actual purchase bill ID you want to remove\n-- Example: SET @BUY_BILL_ID = 123;\n\nSET @BUY_BILL_ID = :BUY_BILL_ID; -- Replace with actual purchase bill ID\n\n-- Start transaction for data integrity\nSTART TRANSACTION;\n\n-- Step 1: Delete cars from cars_stock that are associated with this purchase bill\n-- This deletes cars that have buy_details linked to this purchase bill\nDELETE cs FROM cars_stock cs\nINNER JOIN buy_details bd ON cs.id_buy_details = bd.id\nWHERE bd.id_buy_bill = @BUY_BILL_ID;\n\n-- Step 2: Delete buy_details records associated with this purchase bill\nDELETE FROM buy_details \nWHERE id_buy_bill = @BUY_BILL_ID;\n\n-- Step 3: Delete buy_payments records associated with this purchase bill\nDELETE FROM buy_payments \nWHERE id_buy_bill = @BUY_BILL_ID;\n\n-- Step 4: Delete the purchase bill itself\nDELETE FROM buy_bill \nWHERE id = @BUY_BILL_ID;\n\n-- Commit the transaction\nCOMMIT;\n\n-- Verification queries (optional - run these to verify the deletion)\n-- Check if any cars remain associated with this purchase bill\nSELECT COUNT(*) as remaining_cars \nFROM cars_stock cs\nINNER JOIN buy_details bd ON cs.id_buy_details = bd.id\nWHERE bd.id_buy_bill = @BUY_BILL_ID;\n\n-- Check if any buy_details remain for this purchase bill\nSELECT COUNT(*) as remaining_buy_details \nFROM buy_details \nWHERE id_buy_bill = @BUY_BILL_ID;\n\n-- Check if any payments remain for this purchase bill\nSELECT COUNT(*) as remaining_payments \nFROM buy_payments \nWHERE id_buy_bill = @BUY_BILL_ID;\n\n-- Check if the purchase bill still exists\nSELECT COUNT(*) as purchase_bill_exists \nFROM buy_bill \nWHERE id = @BUY_BILL_ID;','',X'5B224255595F42494C4C5F4944225D',X'7B224255595F42494C4C5F4944223A223437227D','delete purchase');

/*!40000 ALTER TABLE `adv_sql` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table banks
# ------------------------------------------------------------

CREATE TABLE `banks` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `company_name` varchar(255) DEFAULT NULL,
  `bank_name` varchar(255) DEFAULT NULL,
  `swift_code` varchar(255) DEFAULT NULL,
  `bank_account` varchar(255) DEFAULT NULL,
  `bank_address` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

LOCK TABLES `banks` WRITE;
/*!40000 ALTER TABLE `banks` DISABLE KEYS */;

INSERT INTO `banks` (`id`, `company_name`, `bank_name`, `swift_code`, `bank_account`, `bank_address`, `notes`)
VALUES
	(1,'GROUP MERHAB LIMITED','DBS Bank (Hong Kong) Limited','DÐÐ’ÐšÐÐšÐÐ','79969000747875','11th Floor, The Center, 99 Queen\'s Road Central, Central, Hong Kong Hong Kong',NULL),
	(2,'YIWU DASO TRADING COMPANY','DBS Bank (Hong Kong) Limited','DHBKHKHH','79969485592','11th Floor, The Center, 99 Queen\'s Road Central, Central, Hong Kong',''),
	(3,'GROUP MERHAB LIMITED','CASH','CASH','CASH','CASH',''),
	(4,'HK OCEAN VISION INTERNATIONAL TRADING LIMITED','DEUTSCHE BANK AG','DEUTHKHHXXX','1501240910118','hong kong','walle');

/*!40000 ALTER TABLE `banks` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table brands
# ------------------------------------------------------------

CREATE TABLE `brands` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `brand` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `brand` (`brand`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `brands` WRITE;
/*!40000 ALTER TABLE `brands` DISABLE KEYS */;

INSERT INTO `brands` (`id`, `brand`)
VALUES
	(9,'AUDI'),
	(7,'CHANGAN'),
	(8,'CHERRY'),
	(10,'FREIGHT'),
	(2,'GEELY'),
	(11,'JETA'),
	(4,'KIA'),
	(3,'MG'),
	(5,'PEUGEOT'),
	(6,'SKODA'),
	(1,'VW');

/*!40000 ALTER TABLE `brands` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table buy_bill
# ------------------------------------------------------------

CREATE TABLE `buy_bill` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_supplier` int(11) DEFAULT NULL,
  `date_buy` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `payed` decimal(10,2) DEFAULT NULL,
  `pi_path` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  `is_stock_updated` tinyint(1) DEFAULT NULL,
  `is_ordered` tinyint(1) DEFAULT 1,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `bill_ref` (`bill_ref`),
  KEY `id_supplier` (`id_supplier`),
  CONSTRAINT `buy_bill_ibfk_1` FOREIGN KEY (`id_supplier`) REFERENCES `suppliers` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=56 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `buy_bill` WRITE;
/*!40000 ALTER TABLE `buy_bill` DISABLE KEYS */;

INSERT INTO `buy_bill` (`id`, `id_supplier`, `date_buy`, `amount`, `payed`, `pi_path`, `bill_ref`, `is_stock_updated`, `is_ordered`, `notes`)
VALUES
	(2,2,'2025-03-18 07:32:00',152000.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-05-27_07-38-34_1748331514370.pdf','RW-MM20250318-10',1,1,NULL),
	(3,3,'2025-03-15 07:48:00',70000.00,NULL,'purchase_pi/pi_purchase_MESSOUD_KX1_2025-05-27_07-55-22_1748332522440.png','ATW2503DAS001',1,1,NULL),
	(4,2,'2025-03-18 08:00:00',129000.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-05-27_08-01-10_1748332870399.pdf','RW-MM20250318-11',1,1,NULL),
	(5,3,'2025-03-09 08:07:00',70000.00,NULL,'purchase_pi/pi_purchase_MESSOUD_KX1_2025-05-27_08-10-25_1748333425114.png','ATW2503DAS001-1',1,1,NULL),
	(6,2,'2025-03-18 08:14:00',109200.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-05-27_08-14-24_1748333664359.pdf','RW-MM20250318-12',1,1,NULL),
	(7,2,'2025-03-18 08:16:00',19667.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-05-27_08-17-07_1748333827642.pdf','RW-MM20250318-14',1,1,NULL),
	(8,4,'2025-04-03 10:03:00',441000.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-05-27_10-04-17_1748340257681.pdf','HSXGML20250403CQ001I',1,1,NULL),
	(9,4,'2025-03-22 10:23:00',21000.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-05-27_10-23-58_1748341438888.pdf','HSXGML250325CQ001',1,1,NULL),
	(10,4,'2025-04-09 10:27:00',10450.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-05-27_10-28-04_1748341684095.pdf','HSXGML250409BC001I',1,1,NULL),
	(11,1,'2025-03-27 10:31:00',104100.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_10-32-09_1748341929754.pdf','XGYY20250327002',1,1,NULL),
	(12,1,'2025-04-21 10:59:00',49840.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_11-00-37_1748343637273.pdf','XGYY20250401003',1,1,NULL),
	(13,4,'2025-03-20 11:02:00',104400.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-05-27_11-03-08_1748343788489.pdf','HSX20250320CQ01',1,1,NULL),
	(14,1,'2025-04-09 11:05:00',320000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_11-05-27_1748343927371.pdf','XGYY20250401004',1,1,NULL),
	(15,1,'2025-04-29 11:12:00',119000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_11-12-56_1748344376993.pdf','XGYY20250409003',1,1,NULL),
	(16,1,'2025-03-07 11:14:00',100150.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_11-14-54_1748344494840.pdf','XGYY20250307003',1,1,NULL),
	(17,5,'2025-04-01 11:21:00',131400.00,NULL,'purchase_pi/pi_purchase_LEO_2025-05-27_11-22-19_1748344939096.pdf','PIXSL20250401L001',1,1,NULL),
	(18,5,'2025-04-06 11:30:00',14600.00,NULL,'purchase_pi/pi_purchase_LEO_2025-05-27_11-31-12_1748345472329.pdf','PIXSL20250406L002',1,1,NULL),
	(19,4,'2025-05-27 12:23:00',292600.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-05-27_12-24-55_1748348695962.pdf','No.HSXGML2503BC001',1,1,NULL),
	(20,1,'2025-05-23 15:18:00',24000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_15-22-54_1748359374686.pdf','XGYYWL20250523009',1,1,NULL),
	(21,1,'2025-05-23 15:24:00',258000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_15-25-24_1748359524703.pdf','XGYYWL20250523008',1,1,NULL),
	(22,1,'2025-05-04 15:30:00',1427700.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-05-27_15-30-42_1748359842334.pdf','XGYY20250504001',1,1,NULL),
	(23,1,'2025-06-06 19:22:00',86000.00,86000.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-06-06_19-23-55_1749237835827.pdf','YYDK25060603',1,1,NULL),
	(24,4,'2025-05-01 11:38:00',169400.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-06-08_11-42-02_1749382922898.pdf','no pi yet',1,1,NULL),
	(25,6,'2025-05-12 20:51:00',13080.00,NULL,'purchase_pi/pi_purchase_ALI_CHINESE_2025-06-13_20-52-33_1749847953316.pdf','no bill',1,1,NULL),
	(26,4,'2025-05-08 21:11:00',122000.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-06-13_21-12-42_1749849162984.pdf','HSXMML20250508CQ03',1,1,NULL),
	(27,6,'2025-04-22 07:11:00',15864.00,NULL,'purchase_pi/pi_purchase_ALI_CHINESE_2025-06-14_07-12-48_1749885168690.jpg','no ref',1,1,NULL),
	(28,4,'2025-04-12 13:58:00',62700.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-06-14_13-59-05_1749909545004.pdf','HSXGML250412BC001I',1,1,NULL),
	(29,1,'2025-06-08 15:38:00',36900.00,36900.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-06-14_15-39-38_1749915578922.pdf','YYDK25060803',1,1,NULL),
	(30,1,'2025-06-08 12:32:00',88000.00,88000.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-06-15_12-32-44_1749990764323.pdf','YYDK25060801',1,1,NULL),
	(31,1,'2025-06-08 12:16:00',100000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-06-21_12-16-56_1750508216165.pdf','YYDK25060802',1,1,NULL),
	(32,4,'2025-06-02 12:21:00',58800.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-06-21_12-22-04_1750508524156.pdf','HSXGML20250602BC001C',1,1,NULL),
	(34,7,'2025-05-21 14:28:00',52500.00,NULL,'purchase_pi/pi_purchase_bella_2025-06-21_14-28-45_1750516125210.jpg','no bill for bella',1,1,NULL),
	(35,7,'2025-03-21 14:45:00',7500.00,NULL,'purchase_pi/pi_purchase_bella_2025-06-21_14-45-44_1750517144930.jpg','no bill liv ',1,1,NULL),
	(36,2,'2025-04-21 14:46:00',19800.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-06-21_14-47-03_1750517223576.jpg','no bill for golf',1,1,NULL),
	(37,4,'2025-04-21 14:48:00',26500.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-06-21_14-48-36_1750517316507.jpg','no bill for a3',1,1,NULL),
	(38,2,'2025-06-09 11:52:00',444000.00,444000.00,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-06-29_11-53-41_1751198021154.pdf','RW-MM202506-9-01',1,1,NULL),
	(39,6,'2025-05-29 18:49:00',8350.00,NULL,'purchase_pi/pi_purchase_ALI_CHINESE_2025-06-29_18-50-08_1751223008820.pdf','20-05-2025',1,1,NULL),
	(40,1,'2025-03-01 12:23:00',18000.00,NULL,'purchase_pi/pi_purchase_zhengzhou_walle__2025-07-01_12-23-51_1751372631194.png','walli t-roc from begening',1,1,NULL),
	(41,1,'2025-06-24 14:10:00',99600.00,99600.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-07-01_14-12-04_1751379124227.pdf','YYDK25062403',1,1,NULL),
	(42,1,'2025-07-01 14:19:00',36600.00,36600.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-07-01_14-19-37_1751379577251.pdf','YYDK25070103',1,1,NULL),
	(43,6,'2025-07-06 11:20:00',0.00,NULL,'purchase_pi/pi_purchase_ALI_CHINESE_2025-07-06_11-20-49_1751800849898.pdf','any',1,1,NULL),
	(44,2,'2025-07-01 05:40:00',26400.00,NULL,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-07-10_05-41-59_1752126119805.pdf','RW-MM202507-1-01',1,1,NULL),
	(45,2,'2025-07-02 05:44:00',84700.00,82900.00,'purchase_pi/pi_purchase_MUMU_Henan_rongwei_international_trade_co.,ltd_2025-07-10_05-45-17_1752126317883.pdf','RW-MM20250702-01',1,1,NULL),
	(46,6,'2025-07-10 16:50:00',0.00,NULL,'purchase_pi/pi_purchase_ALI_CHINESE_2025-07-10_16-51-02_1752166262945.png','just. for loading ',1,0,NULL),
	(48,6,'2025-05-12 21:50:00',0.00,7600.00,'purchase_pi/pi_purchase_ALI_CHINESE_2025-07-14_21-51-55_1752529915358.pdf','HYXS-250512-001',1,1,NULL),
	(49,1,'2025-07-16 19:55:00',107200.00,107200.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-07-17_19-56-27_1752782187553.pdf','YYDK25071608',1,1,NULL),
	(52,8,'2025-04-17 23:12:00',9862.00,NULL,'purchase_pi/pi_purchase_ChongQin_huanyu_2025-07-17_23-15-59_1752794159757.pdf','HYXS-250417-001',1,1,NULL),
	(53,1,'2025-07-04 14:31:00',145000.00,145000.00,'purchase_pi/pi_purchase_zhengzhou_walle__2025-07-20_14-31-43_1753021903063.pdf','XGYYZZ20250704001',1,1,NULL),
	(54,4,'2025-07-21 08:12:00',1740000.00,NULL,'purchase_pi/pi_purchase_Foshan_Hui_Shunxing_Automobile_Trading_Co.,_Ltd_2025-07-21_08-13-46_1753085626259.pdf','HSXMER20250719CQ01',1,1,NULL),
	(55,9,'2025-05-17 11:43:00',366000.00,NULL,'purchase_pi/pi_purchase_Weifang_Century_Sovereign_Automobile_Sales_Co.,_Ltd_2025-07-21_11-44-34_1753098274856.pdf','SJJY20250517JM',1,1,NULL);

/*!40000 ALTER TABLE `buy_bill` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table buy_details
# ------------------------------------------------------------

CREATE TABLE `buy_details` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_car_name` int(11) DEFAULT NULL,
  `id_color` int(11) DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `QTY` int(11) DEFAULT NULL,
  `year` int(11) DEFAULT NULL,
  `month` int(11) DEFAULT NULL,
  `is_used_car` tinyint(1) DEFAULT 0,
  `id_buy_bill` int(11) DEFAULT NULL,
  `price_sell` decimal(10,2) DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  KEY `id_car_name` (`id_car_name`),
  KEY `id_color` (`id_color`),
  KEY `id_buy_bill` (`id_buy_bill`),
  CONSTRAINT `buy_details_ibfk_1` FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`),
  CONSTRAINT `buy_details_ibfk_2` FOREIGN KEY (`id_color`) REFERENCES `colors` (`id`),
  CONSTRAINT `buy_details_ibfk_3` FOREIGN KEY (`id_buy_bill`) REFERENCES `buy_bill` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=78 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `buy_details` WRITE;
/*!40000 ALTER TABLE `buy_details` DISABLE KEYS */;

INSERT INTO `buy_details` (`id`, `id_car_name`, `id_color`, `amount`, `notes`, `QTY`, `year`, `month`, `is_used_car`, `id_buy_bill`, `price_sell`, `is_big_car`)
VALUES
	(2,9,1,7600.00,'please check if we add change km counter cost or not',20,2025,5,0,2,8600.00,0),
	(3,16,1,8750.00,'',8,2025,5,0,3,9750.00,0),
	(4,10,1,6450.00,'',20,2025,5,0,4,7450.00,0),
	(5,16,1,8750.00,'',8,2025,5,0,5,9750.00,0),
	(6,7,2,18200.00,'',6,2025,5,0,6,19200.00,0),
	(7,17,1,19667.00,'we changed the KM to 6000 \nmust the doc to Algeria be used car',1,2025,5,1,7,20667.00,0),
	(8,9,1,7350.00,'',60,2025,5,0,8,8350.00,0),
	(9,18,3,21000.00,'',1,2025,5,0,9,22000.00,0),
	(10,19,1,10450.00,'docs to Algeria must be used car ',1,2023,5,1,10,11450.00,0),
	(11,20,1,11100.00,'',3,2025,5,0,11,12100.00,0),
	(12,7,1,17700.00,'',4,2025,5,0,11,18700.00,0),
	(13,11,1,12460.00,'we add english system 360usd each',4,2025,5,0,12,13460.00,0),
	(14,16,1,8700.00,'',12,2025,5,0,13,9700.00,0),
	(15,15,1,6400.00,'',35,2025,5,0,14,7400.00,0),
	(16,15,4,6400.00,'',15,2025,5,0,14,7400.00,0),
	(17,20,1,11900.00,'',10,2025,5,0,15,12900.00,0),
	(18,17,1,21100.00,'',1,2025,5,0,16,22100.00,0),
	(19,16,1,8650.00,'',3,2025,5,0,16,9650.00,0),
	(20,7,1,17700.00,'',3,2025,5,0,16,18700.00,0),
	(21,9,1,7300.00,'',18,2025,5,0,17,8300.00,0),
	(22,9,1,7300.00,'',2,2025,5,0,18,8300.00,0),
	(23,9,1,7700.00,'',38,2025,5,0,19,8700.00,0),
	(24,11,5,12300.00,'',2,2025,5,0,20,13000.00,0),
	(25,13,1,8600.00,'',22,2025,5,0,21,9600.00,0),
	(26,13,4,8600.00,'',8,2025,5,0,21,9600.00,0),
	(27,12,1,12000.00,'',10,2025,5,0,22,13000.00,0),
	(28,21,1,13300.00,'',5,2025,5,0,22,14300.00,0),
	(29,22,1,12300.00,'',6,2025,5,0,22,13300.00,0),
	(30,9,1,7500.00,'',8,2025,5,0,22,8500.00,0),
	(31,23,1,7500.00,'',30,2025,5,0,22,8500.00,0),
	(32,7,6,17900.00,'',2,2025,5,0,22,18900.00,0),
	(33,24,1,10200.00,'',20,2022,12,1,22,11200.00,0),
	(36,26,4,12500.00,'disc',5,2025,5,0,22,13800.00,0),
	(37,14,1,7820.00,'i get discount ',5,2025,5,0,22,9100.00,0),
	(38,13,1,8700.00,'disc',4,2025,5,0,22,10000.00,0),
	(39,13,1,8600.00,'',10,2025,6,0,23,9600.00,0),
	(40,9,1,7700.00,'',22,2025,6,0,24,8700.00,0),
	(41,28,4,13080.00,'',1,2023,6,1,25,13080.00,0),
	(42,11,4,12200.00,'',10,2025,6,0,26,13200.00,0),
	(43,29,1,15864.00,'',1,2023,6,1,27,15864.00,1),
	(44,19,1,10450.00,'I think the price is not exact\nno pi ',6,2023,6,1,28,11450.00,0),
	(45,30,2,18450.00,'',2,2025,6,0,29,19450.00,0),
	(46,22,2,12500.00,'',2,2025,6,0,30,13500.00,0),
	(47,22,4,12600.00,'',5,2025,6,0,30,13600.00,0),
	(48,24,1,10000.00,'',10,2023,12,1,31,11000.00,0),
	(49,25,4,14700.00,'',4,2025,6,0,32,15700.00,0),
	(50,7,1,17500.00,'',3,2025,6,0,34,18500.00,0),
	(52,9,1,7500.00,'',1,2025,6,0,35,8500.00,0),
	(53,18,2,19800.00,'',1,2025,6,0,36,20800.00,0),
	(54,31,4,26500.00,'',1,2025,6,0,37,27500.00,0),
	(55,9,1,7400.00,'',60,2025,6,0,38,8400.00,0),
	(56,32,1,8350.00,'',1,2025,6,1,39,9350.00,0),
	(57,7,1,18000.00,'prices are not correct ',1,2025,7,0,40,19000.00,0),
	(58,7,1,17500.00,'',3,2025,7,0,41,18500.00,0),
	(59,33,1,10500.00,'',1,2025,7,0,41,11500.00,0),
	(60,30,2,18300.00,'',2,2025,7,0,41,19300.00,0),
	(61,15,1,7000.00,'',4,2025,7,0,42,8000.00,0),
	(62,34,1,8600.00,'',1,2025,7,0,42,9600.00,0),
	(63,33,1,0.00,'we only doing freight',1,2025,7,0,43,0.00,0),
	(64,10,1,6600.00,'',4,2025,7,0,44,7600.00,0),
	(65,11,5,12100.00,'',7,2025,7,0,45,13100.00,0),
	(66,22,1,0.00,'just for loading',1,2025,7,0,46,0.00,0),
	(68,38,4,13400.00,'',4,2025,7,0,49,14400.00,0),
	(69,38,2,13400.00,'',4,2025,7,0,49,14400.00,0),
	(70,32,1,9862.00,'',1,2023,7,1,48,10865.00,1),
	(73,39,1,9862.00,'',1,2025,7,1,52,10862.00,0),
	(74,25,4,14500.00,NULL,10,2025,7,0,53,15500.00,0),
	(75,40,4,85000.00,NULL,10,2025,7,0,54,9500.00,0),
	(76,13,4,89000.00,NULL,10,2025,7,0,54,9900.00,0),
	(77,7,3,18300.00,NULL,20,2025,7,0,55,19300.00,0);

/*!40000 ALTER TABLE `buy_details` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table buy_payments
# ------------------------------------------------------------

CREATE TABLE `buy_payments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_buy_bill` int(11) DEFAULT NULL,
  `date_payment` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `swift_path` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_buy_bill` (`id_buy_bill`),
  CONSTRAINT `buy_payments_ibfk_1` FOREIGN KEY (`id_buy_bill`) REFERENCES `buy_bill` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=18 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `buy_payments` WRITE;
/*!40000 ALTER TABLE `buy_payments` DISABLE KEYS */;

INSERT INTO `buy_payments` (`id`, `id_buy_bill`, `date_payment`, `amount`, `swift_path`, `notes`, `id_user`)
VALUES
	(1,23,'2025-06-06 19:29:00',86000.00,'/api/upload.php?path=buy_payment_swifts/bill_23_1749238166727.pdf','',1),
	(2,29,'2025-06-08 15:41:00',36900.00,'/api/upload.php?path=buy_payment_swifts/bill_29_1749915732189.jpg','together with others',1),
	(3,30,'2025-06-13 12:34:00',88000.00,'/api/upload.php?path=buy_payment_swifts/bill_30_1749990907538.jpg','',1),
	(4,38,'2025-06-29 11:54:00',444000.00,'/api/upload.php?path=buy_payment_swifts/bill_38_1751198274399.pdf','',1),
	(5,41,'2025-06-26 14:17:00',99600.00,'/api/upload.php?path=buy_payment_swifts/bill_41_1751379490604.jpg','',1),
	(6,42,'2025-07-01 14:22:00',36600.00,'/api/upload.php?path=buy_payment_swifts/bill_42_1751379774558.JPG','',1),
	(8,45,'2025-07-14 18:43:00',51800.00,'/api/upload.php?path=buy_payment_swifts/bill_45_1752518618600.JPG','',1),
	(10,48,'2025-05-28 21:59:00',7600.00,'/api/upload.php?path=buy_payment_swifts/bill_48_1752530411654.jpg','i give client fob cost 7999',2),
	(12,49,'2025-07-16 20:30:00',107200.00,'/api/upload.php?path=buy_payment_swifts/bill_49_1752784313802.jpg','',1),
	(16,53,'2025-07-05 14:33:00',145000.00,'/api/upload.php?path=buy_payment_swifts/bill_53_1753022036661.JPG','',1),
	(17,45,'2025-07-20 14:39:00',31100.00,'/api/upload.php?path=buy_payment_swifts/bill_45_1753022392864.jpg','this is all no money left. ',1);

/*!40000 ALTER TABLE `buy_payments` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table cars_names
# ------------------------------------------------------------

CREATE TABLE `cars_names` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_name` varchar(255) DEFAULT NULL,
  `id_brand` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT 0,
  `cbm` decimal(10,0) DEFAULT NULL,
  `gw` decimal(10,0) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `car_name` (`car_name`),
  KEY `id_brand` (`id_brand`),
  CONSTRAINT `cars_names_ibfk_1` FOREIGN KEY (`id_brand`) REFERENCES `brands` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `cars_names` WRITE;
/*!40000 ALTER TABLE `cars_names` DISABLE KEYS */;

INSERT INTO `cars_names` (`id`, `car_name`, `id_brand`, `notes`, `is_big_car`, `cbm`, `gw`)
VALUES
	(7,'T-ROC STARLIGHT',1,'',0,NULL,NULL),
	(9,'LIVAN AUTO',2,'',0,NULL,NULL),
	(10,'LIVAN MAN',2,'',0,NULL,NULL),
	(11,'COOLRAY SPORT',2,'',0,NULL,NULL),
	(12,'COOLRAY FULL NOT SPORT',2,'',0,NULL,NULL),
	(13,'COOLRAY SUPER AUTO no sunroof',2,'',0,NULL,NULL),
	(14,'COOLRAY SUPER MANUAL',2,'',0,NULL,NULL),
	(15,'MG5 MAN',3,'',0,NULL,NULL),
	(16,'KX1 20251.4LCVT SunroofEdition',4,'',0,NULL,NULL),
	(17,'TIGUAN L',1,'',0,NULL,NULL),
	(18,'Golf 300TSI R-Line',1,'',0,NULL,NULL),
	(19,'2008',5,'',0,NULL,NULL),
	(20,'SONET BLACK ROOF',4,'',0,NULL,NULL),
	(21,'KAMIQ GT',6,'',0,NULL,NULL),
	(22,'SELTOS LUXERY BLACK ROOF',4,'',0,NULL,NULL),
	(23,'EMGRAND MAN',2,'',0,NULL,NULL),
	(24,'TACOUA',1,'',0,NULL,NULL),
	(25,'THARU TOP',1,'',0,NULL,NULL),
	(26,'THARU MID',1,'',0,NULL,NULL),
	(27,'THARU BASIC',1,'',0,NULL,NULL),
	(28,'CHANGAN CS75 PLUS',7,'',0,NULL,NULL),
	(29,'CHERRY TIGO 7',8,'',1,NULL,NULL),
	(30,'T-ROC WITH PACKAKE',1,'',0,NULL,NULL),
	(31,'A3',9,'',0,NULL,NULL),
	(32,'TIGO 3',8,'',1,NULL,NULL),
	(33,'K3',4,'',0,NULL,NULL),
	(34,'MG5 BASE AUTO',3,'',0,NULL,NULL),
	(35,'FREIGHT',NULL,'',0,NULL,NULL),
	(36,'Chery Tiggo 3x',8,'',0,NULL,NULL),
	(38,'JETTA VS5',11,'',0,NULL,NULL),
	(39,'EMGRAND USED CAR',2,'',1,NULL,NULL),
	(40,'COOLRAY SUPER AUTO WITH SUNROOF',NULL,'',0,NULL,NULL);

/*!40000 ALTER TABLE `cars_names` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table cars_stock
# ------------------------------------------------------------

CREATE TABLE `cars_stock` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `notes` text DEFAULT NULL,
  `id_buy_details` int(11) DEFAULT NULL,
  `date_sell` datetime DEFAULT NULL,
  `id_client` int(11) DEFAULT NULL,
  `price_cell` decimal(10,2) DEFAULT NULL,
  `freight` decimal(10,2) DEFAULT NULL,
  `id_port_loading` int(11) DEFAULT NULL,
  `id_port_discharge` int(11) DEFAULT NULL,
  `vin` varchar(255) DEFAULT NULL,
  `path_documents` varchar(255) DEFAULT NULL,
  `date_loding` datetime DEFAULT NULL,
  `date_send_documents` datetime DEFAULT NULL,
  `hidden` tinyint(1) DEFAULT 0,
  `id_sell_pi` varchar(255) DEFAULT NULL,
  `sell_pi_path` varchar(255) DEFAULT NULL,
  `buy_pi_path` varchar(255) DEFAULT NULL,
  `id_sell` int(11) DEFAULT NULL,
  `export_lisence_ref` varchar(255) DEFAULT NULL,
  `id_warehouse` int(11) DEFAULT NULL,
  `in_wharhouse_date` date DEFAULT NULL,
  `date_get_documents_from_supp` date DEFAULT NULL,
  `date_get_keys_from_supp` date DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `date_get_bl` date DEFAULT NULL,
  `date_pay_freight` date DEFAULT NULL,
  `is_used_car` tinyint(1) DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT 0,
  `container_ref` varchar(255) DEFAULT NULL,
  `is_tmp_client` tinyint(1) DEFAULT 0,
  `id_loaded_container` int(11) DEFAULT NULL,
  `is_batch` tinyint(1) DEFAULT 0,
  `is_loading_inquiry_sent` tinyint(1) DEFAULT 0,
  `date_assigned` timestamp NULL DEFAULT NULL,
  `id_color` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_client` (`id_client`),
  KEY `id_port_loading` (`id_port_loading`),
  KEY `id_port_discharge` (`id_port_discharge`),
  KEY `id_buy_details` (`id_buy_details`),
  CONSTRAINT `cars_stock_ibfk_1` FOREIGN KEY (`id_client`) REFERENCES `clients` (`id`),
  CONSTRAINT `cars_stock_ibfk_2` FOREIGN KEY (`id_port_loading`) REFERENCES `loading_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_3` FOREIGN KEY (`id_port_discharge`) REFERENCES `discharge_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_4` FOREIGN KEY (`id_buy_details`) REFERENCES `buy_details` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=673 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `cars_stock` WRITE;
/*!40000 ALTER TABLE `cars_stock` DISABLE KEYS */;

INSERT INTO `cars_stock` (`id`, `notes`, `id_buy_details`, `date_sell`, `id_client`, `price_cell`, `freight`, `id_port_loading`, `id_port_discharge`, `vin`, `path_documents`, `date_loding`, `date_send_documents`, `hidden`, `id_sell_pi`, `sell_pi_path`, `buy_pi_path`, `id_sell`, `export_lisence_ref`, `id_warehouse`, `in_wharhouse_date`, `date_get_documents_from_supp`, `date_get_keys_from_supp`, `rate`, `date_get_bl`, `date_pay_freight`, `is_used_car`, `is_big_car`, `container_ref`, `is_tmp_client`, `id_loaded_container`, `is_batch`, `is_loading_inquiry_sent`, `date_assigned`, `id_color`)
VALUES
	(2,NULL,2,NULL,NULL,8600.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,NULL,1),
	(3,NULL,2,NULL,NULL,8600.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,NULL,1),
	(4,NULL,2,NULL,NULL,8600.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,NULL,1),
	(5,'READY FOR LOADING CAR',2,'2025-07-16 00:00:00',350,7600.00,1600.00,NULL,3,'LLV2C3B28S0005305',NULL,NULL,NULL,0,'MOH006160725173',NULL,NULL,173,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'MSDU7188538',0,21,0,0,'2025-07-16 06:03:57',1),
	(6,NULL,2,'2025-07-10 00:00:00',54,8800.00,1600.00,NULL,3,'LLV2C3B20S0002995',NULL,'2025-06-16 00:00:00',NULL,0,'MOH006100725153',NULL,NULL,153,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'TGBU7905296',0,NULL,0,0,NULL,1),
	(7,NULL,2,'2025-07-10 00:00:00',316,8800.00,1600.00,NULL,3,'LLV2C3B21S0003055',NULL,'2025-06-15 00:00:00',NULL,0,'MOH006100725152',NULL,NULL,152,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'CSNU6998621',0,NULL,0,0,NULL,1),
	(8,'buy car on sea',2,'2025-07-10 00:00:00',310,8800.00,1600.00,NULL,3,'LLV2C3B26S0001415',NULL,'2025-06-15 00:00:00',NULL,0,'MOH006100725150',NULL,NULL,150,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'FSCU9349902',0,NULL,0,0,NULL,1),
	(9,NULL,2,'2025-07-10 00:00:00',309,8800.00,1600.00,NULL,3,'LLV2C3B22S0003100',NULL,'2025-06-15 00:00:00',NULL,0,'MOH006100725149',NULL,NULL,149,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'CSNU6998621',0,NULL,0,0,NULL,1),
	(10,NULL,2,'2025-07-10 00:00:00',308,8800.00,1600.00,NULL,3,'LLV2C3B21S0005324',NULL,NULL,NULL,0,'MOH006100725148',NULL,NULL,148,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'MSDU7188538',0,21,0,0,NULL,1),
	(11,NULL,2,'2025-07-10 00:00:00',307,8800.00,1600.00,NULL,3,'LLV2C3B20S0003077',NULL,'2025-06-15 00:00:00',NULL,0,'MOH006100725147',NULL,NULL,147,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'CSNU6998621',0,NULL,0,0,NULL,1),
	(12,NULL,2,'2025-07-10 00:00:00',306,8600.00,1600.00,NULL,3,'LLV2C3B2XS0002969',NULL,'2025-06-15 00:00:00',NULL,0,'MOH006100725146',NULL,NULL,146,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'CSNU6998621',0,NULL,0,0,NULL,1),
	(13,NULL,2,'2025-07-10 00:00:00',302,8800.00,1600.00,NULL,4,'LLV2C3B24S0001395',NULL,'2025-06-18 00:00:00',NULL,0,'MOH006100725144',NULL,NULL,144,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'FSCU9349902',0,NULL,0,0,NULL,1),
	(14,NULL,2,'2025-06-24 00:00:00',276,8600.00,1600.00,NULL,3,'LLV2C3B20S0005329',NULL,NULL,NULL,0,'ISH003240625118',NULL,NULL,118,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'FFAU3202754',0,15,0,0,NULL,1),
	(15,NULL,2,'2025-06-24 00:00:00',275,8800.00,1600.00,NULL,3,'LLV2C3B26S0001334','cars/15/documents/documents_1752789761736.pdf',NULL,NULL,0,'ISH003240625117',NULL,NULL,117,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'TCLU5400775',0,7,0,0,NULL,1),
	(16,NULL,2,'2025-06-24 00:00:00',254,9000.00,1600.00,NULL,3,'LLV2C3B29S0000470',NULL,'2025-06-29 00:00:00',NULL,0,'ADM001120625116',NULL,NULL,116,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'CSNU7709554',0,NULL,0,0,NULL,1),
	(17,NULL,2,'2025-06-24 00:00:00',253,9000.00,1600.00,NULL,3,'LLV2C3B23S0000450',NULL,'2025-06-29 00:00:00',NULL,0,'ADM001120625116',NULL,NULL,116,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,'OOCU8632437',0,NULL,0,0,NULL,1),
	(18,NULL,2,'2025-06-22 00:00:00',263,7700.00,1600.00,NULL,3,'LLV2C3B24S0001199','cars/18/documents/documents_1752789734528.pdf',NULL,NULL,0,'ISH003220625112',NULL,NULL,112,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',NULL,0,'TCLU5400775',0,7,0,0,NULL,1),
	(19,NULL,2,'2025-06-22 00:00:00',261,7700.00,1600.00,NULL,3,'LLV2C3B29S0000338',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003220625112','cars/19/sell_pi/sell_pi_1752825884262.pdf',NULL,112,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',NULL,0,'TRHU6435993',0,NULL,0,0,NULL,1),
	(20,NULL,2,'2025-06-22 00:00:00',262,7700.00,1600.00,NULL,3,'LLV2C3B21S0000463',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003220625112',NULL,NULL,112,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',NULL,0,'TGBU9255184',0,NULL,0,0,NULL,1),
	(21,NULL,2,'2025-06-22 00:00:00',265,7700.00,1600.00,NULL,3,'LLV2C3B29S0000307',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003220625112',NULL,NULL,112,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',NULL,0,'TGBU9255184',0,NULL,0,0,NULL,1),
	(22,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:29:26',1),
	(23,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:29:19',1),
	(24,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:29:12',1),
	(25,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:29:04',1),
	(26,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:28:54',1),
	(27,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:28:13',1),
	(28,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:59',1),
	(29,NULL,3,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:48',1),
	(30,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(31,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(32,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(33,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(34,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(35,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(36,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(37,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(38,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(39,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(40,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(41,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(42,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(43,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(44,NULL,4,NULL,NULL,7450.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,105,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(45,NULL,4,'2025-06-17 00:00:00',244,7700.00,1500.00,NULL,2,'LLV2C3A28S0000395',NULL,'2025-05-21 00:00:00',NULL,0,'ADM001140525082','cars/45/sell_pi/sell_pi_1752825961068.pdf',NULL,82,NULL,NULL,'2025-05-07','2025-05-24','2025-05-07',250.00,'2025-06-13','2025-06-03',NULL,0,'COSU6417469830',0,NULL,0,0,NULL,1),
	(46,NULL,4,'2025-06-14 00:00:00',74,7780.00,1500.00,NULL,2,'LLV2C3A24S0000393',NULL,'2025-05-21 00:00:00',NULL,0,'ATE005140425058',NULL,NULL,58,NULL,NULL,'2025-04-07','2025-04-24','2025-04-07',250.00,'2025-05-14','2025-05-04',NULL,0,'COSU6417469830',0,NULL,0,0,NULL,1),
	(47,NULL,4,'2025-06-13 00:00:00',176,7700.00,1500.00,NULL,2,'LLV2C3A29S0000454',NULL,'2025-05-21 00:00:00',NULL,0,'MOH006140525035','cars/47/sell_pi/sell_pi_1752826263219.pdf',NULL,35,NULL,NULL,'2025-05-07','2025-05-24','2025-05-07',250.00,'2025-06-13','2025-06-03',NULL,0,'COSU6417469830',0,NULL,0,0,NULL,1),
	(48,NULL,4,'2025-06-13 00:00:00',150,7600.00,1600.00,NULL,3,'LLV2C3B23S0000478',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003140525021',NULL,NULL,21,NULL,NULL,'2025-05-07','2025-05-24','2025-05-07',250.00,'2025-06-13','2025-06-03',NULL,0,'COSU6453375060',0,NULL,0,0,NULL,1),
	(49,NULL,4,'2025-06-13 00:00:00',53,7700.00,1500.00,NULL,3,'LLV2C3B25S0000210',NULL,'2025-06-29 00:00:00',NULL,0,'MOH006140525014',NULL,NULL,14,NULL,NULL,'2025-05-07','2025-05-24','2025-05-07',250.00,'2025-06-13','2025-06-03',NULL,0,'OOCU8632437',0,NULL,0,0,NULL,1),
	(50,NULL,5,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:35',1),
	(51,NULL,5,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:26',1),
	(52,NULL,5,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:20',1),
	(53,NULL,5,'2025-07-15 00:00:00',334,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001150725163',NULL,NULL,163,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,0,NULL,0,0,'2025-07-15 15:27:11',1),
	(54,NULL,5,NULL,NULL,9750.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(55,NULL,5,NULL,NULL,9750.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(56,NULL,5,NULL,NULL,9750.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(57,NULL,5,NULL,NULL,9750.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',NULL,0,NULL,0,NULL,1,0,NULL,1),
	(58,NULL,6,'2025-06-23 00:00:00',257,19260.00,1500.00,NULL,3,'LFV2B2A19S5517219',NULL,NULL,NULL,0,'ADM001230425115',NULL,NULL,115,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,NULL,0,NULL,0,0,NULL,2),
	(59,NULL,6,'2025-06-23 00:00:00',256,19260.00,1500.00,NULL,3,'LFV2B2A13S5517300',NULL,NULL,NULL,0,'ADM001230425115',NULL,NULL,115,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,NULL,0,NULL,0,NULL,0,0,NULL,2),
	(60,NULL,6,NULL,NULL,19200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,103,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,2),
	(61,NULL,6,NULL,NULL,19200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,103,NULL,NULL,'2025-03-11','2025-03-28','2025-03-11',NULL,'2025-04-17','2025-04-07',NULL,0,NULL,0,NULL,1,0,NULL,2),
	(62,NULL,6,'2025-06-19 00:00:00',245,19600.00,1600.00,NULL,3,'LFV2B2A15S5529352',NULL,NULL,NULL,0,'ISH003190625083',NULL,NULL,83,NULL,NULL,'2025-06-12','2025-06-29','2025-06-12',250.00,'2025-07-19','2025-07-09',NULL,0,'CAIU7094512',0,13,0,0,NULL,2),
	(63,NULL,6,'2025-07-03 00:00:00',NULL,19200.00,NULL,NULL,NULL,'LFV2B2A10S5529386',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',NULL,0,'CAIU7094512',0,13,0,0,NULL,2),
	(64,'must send used car documents to Algeria \nthis car we changed the counter to 6000km',7,'2025-05-28 00:00:00',10,24900.00,1500.00,NULL,2,'LSVY460T8S2036060','cars/64/documents/documents_1751270410288.pdf','2025-05-02 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',NULL,0,'MEDUG6504393',0,NULL,0,0,NULL,1),
	(65,'FOB freight not payed ',8,'2025-06-13 00:00:00',128,7700.00,2000.00,NULL,2,'LLV2C3B22S0002996',NULL,'2025-06-18 00:00:00',NULL,0,'MOH006130625007',NULL,NULL,7,NULL,NULL,'2025-06-06','2025-06-23','2025-06-06',250.00,'2025-07-13','2025-07-03',0,0,'OOCU9215356',0,NULL,0,0,NULL,1),
	(66,NULL,8,'2025-06-23 00:00:00',270,8800.00,1600.00,NULL,3,'LLV2C3B2XS0001384','cars/66/documents/documents_1752789683750.pdf',NULL,NULL,0,'ISH003230625114',NULL,NULL,114,NULL,NULL,'2025-06-16','2025-07-03','2025-06-16',250.00,'2025-07-23','2025-07-13',0,0,'TCLU5400775',0,7,0,0,NULL,1),
	(67,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(68,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(69,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(70,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(71,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(72,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(73,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(74,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(75,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(76,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(77,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(78,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(79,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(80,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(81,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(82,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(83,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(84,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(85,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(86,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(87,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(88,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(89,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(90,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(91,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(92,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(93,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(94,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(95,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(96,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(97,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(98,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(99,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(100,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(101,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(102,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(103,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(104,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,102,NULL,NULL,'2025-03-27','2025-04-13','2025-03-27',NULL,'2025-05-03','2025-04-23',0,0,NULL,0,NULL,1,0,NULL,1),
	(105,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,94,NULL,NULL,'2025-03-25','2025-04-11','2025-03-25',NULL,'2025-05-01','2025-04-21',0,0,NULL,0,NULL,1,0,NULL,1),
	(106,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,94,NULL,NULL,'2025-03-25','2025-04-11','2025-03-25',NULL,'2025-05-01','2025-04-21',0,0,NULL,0,NULL,1,0,NULL,1),
	(107,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,94,NULL,NULL,'2025-03-25','2025-04-11','2025-03-25',NULL,'2025-05-01','2025-04-21',0,0,NULL,0,NULL,1,0,NULL,1),
	(108,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,94,NULL,NULL,'2025-03-25','2025-04-11','2025-03-25',NULL,'2025-05-01','2025-04-21',0,0,NULL,0,NULL,1,0,NULL,1),
	(109,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(110,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(111,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(112,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(113,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(114,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(115,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(116,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(117,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(118,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(119,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(120,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(121,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(122,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(123,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(124,'',8,NULL,NULL,8350.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(125,'',9,NULL,NULL,22000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,101,NULL,NULL,'2025-03-15','2025-04-01','2025-03-15',NULL,'2025-04-21','2025-04-11',0,0,NULL,0,NULL,1,0,NULL,3),
	(126,NULL,10,'2025-06-13 00:00:00',123,11400.00,1500.00,NULL,2,NULL,NULL,NULL,NULL,0,'MOH006120425036',NULL,NULL,36,NULL,NULL,'2025-04-05','2025-04-22','2025-04-05',250.00,'2025-05-12','2025-05-02',1,0,NULL,0,NULL,0,0,NULL,1),
	(127,'',11,NULL,NULL,12100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(128,'',11,NULL,NULL,12100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(129,'',11,NULL,NULL,12100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(130,NULL,12,'2025-06-14 00:00:00',234,18800.00,1600.00,NULL,3,'LFV2B2A10S5522163',NULL,NULL,NULL,0,'ATE005190525066',NULL,NULL,66,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,NULL,0,20,0,0,NULL,1),
	(131,NULL,12,'2025-06-13 00:00:00',179,18460.00,1500.00,NULL,2,'LFV2B2A12S5522097',NULL,NULL,NULL,0,'MOH006120525038',NULL,NULL,38,NULL,NULL,'2025-05-05','2025-05-22','2025-05-05',250.00,'2025-06-11','2025-06-01',0,0,NULL,0,20,0,0,NULL,1),
	(132,NULL,12,'2025-06-13 00:00:00',161,18460.00,1500.00,NULL,2,'LFV2B2A13S5513618',NULL,'2025-06-17 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'MRKU3780687',0,NULL,0,0,NULL,1),
	(133,NULL,12,'2025-06-13 00:00:00',44,18460.00,1500.00,NULL,2,'LFV2B2A18S5514246',NULL,'2025-06-17 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417451857',0,NULL,0,0,NULL,1),
	(134,NULL,13,'2025-07-02 00:00:00',295,13600.00,1500.00,NULL,2,'LB37622Z6SX610389',NULL,NULL,NULL,0,'ADM001200325135',NULL,NULL,135,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(135,NULL,13,'2025-06-30 00:00:00',291,13360.00,1600.00,NULL,3,'LB37622Z4SX611380',NULL,'2025-06-14 00:00:00',NULL,0,'ISH003160425130',NULL,NULL,130,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'COSU6417469880',0,NULL,0,0,NULL,1),
	(136,NULL,13,'2025-06-30 00:00:00',88,13600.00,1600.00,NULL,2,NULL,NULL,NULL,NULL,0,'ISH003210425129',NULL,NULL,129,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(137,'we add english system 360usd each',13,'2025-05-28 00:00:00',22,12500.00,1500.00,NULL,4,'LB37622Z9SX610306',NULL,'2025-06-26 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'COSU6417469880',0,NULL,0,0,NULL,1),
	(138,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(139,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(140,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(141,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,104,NULL,NULL,'2025-03-02','2025-03-19','2025-03-02',NULL,'2025-04-08','2025-03-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(142,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,99,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(143,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,99,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(144,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,99,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(145,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,99,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(146,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,98,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(147,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,98,NULL,NULL,'2025-03-13','2025-03-30','2025-03-13',NULL,'2025-04-19','2025-04-09',0,0,NULL,0,NULL,1,0,NULL,1),
	(148,'',14,NULL,NULL,9700.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,95,NULL,NULL,'2025-02-28','2025-03-17','2025-02-28',NULL,'2025-04-06','2025-03-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(149,NULL,14,'2025-06-14 00:00:00',210,10500.00,1500.00,NULL,2,'LJD4AA193S0134133',NULL,'2025-06-29 00:00:00',NULL,0,'ATE005060425048',NULL,NULL,48,NULL,NULL,'2025-03-30','2025-04-16','2025-03-30',250.00,'2025-05-06','2025-04-26',0,0,'TCNU2038063',0,NULL,0,0,NULL,1),
	(150,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(151,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(152,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(153,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(154,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(155,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(156,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(157,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(158,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(159,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(160,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(161,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(162,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(163,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(164,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(165,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(166,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(167,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(168,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(169,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(170,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(171,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(172,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(173,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(174,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(175,'',15,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,1),
	(176,NULL,15,'2025-06-17 00:00:00',243,7500.00,1500.00,NULL,2,'LSJA36U38SN052418',NULL,'2025-06-17 00:00:00',NULL,0,'ADM001150425081',NULL,NULL,81,NULL,NULL,'2025-04-08','2025-04-25','2025-04-08',250.00,'2025-05-15','2025-05-05',0,0,'MRKU3780687',0,NULL,0,0,NULL,1),
	(177,NULL,15,'2025-06-14 00:00:00',231,7900.00,1500.00,NULL,2,'LSJA36U38SN051723',NULL,NULL,NULL,0,'ATE005160425063',NULL,NULL,63,NULL,NULL,'2025-04-09','2025-04-26','2025-04-09',250.00,'2025-05-16','2025-05-06',0,0,NULL,0,NULL,0,0,NULL,1),
	(178,NULL,15,'2025-06-14 00:00:00',230,7900.00,1500.00,NULL,2,'LSJA36U32SN051751','cars/178/documents/documents_1750182303493.pdf','2025-06-17 00:00:00',NULL,0,'ATE005010525061',NULL,NULL,61,NULL,NULL,'2025-04-24','2025-05-11','2025-04-24',250.00,'2025-05-31','2025-05-21',0,0,'QWSEF25057042B',0,NULL,0,0,NULL,1),
	(179,NULL,15,'2025-06-14 00:00:00',226,7700.00,1500.00,NULL,2,'LSJA36U37SN051681','cars/179/documents/documents_1750180162068.pdf','2025-06-17 00:00:00',NULL,0,'ATE005110425057',NULL,NULL,57,NULL,NULL,'2025-04-04','2025-04-21','2025-04-04',250.00,'2025-05-11','2025-05-01',0,0,'QWSEF25056038B',0,NULL,0,0,NULL,1),
	(180,NULL,15,'2025-06-14 00:00:00',73,7500.00,1500.00,NULL,2,'LSJA36U38SN051768','cars/180/documents/documents_1750180039349.pdf','2025-06-17 00:00:00',NULL,0,'ATE005110425056',NULL,NULL,56,NULL,NULL,'2025-04-04','2025-04-21','2025-04-04',250.00,'2025-05-11','2025-05-01',0,0,'QWSEF25056038C',0,NULL,0,0,NULL,1),
	(181,NULL,15,'2025-06-14 00:00:00',224,7700.00,1500.00,NULL,2,'LSJA36U36SN051686',NULL,NULL,NULL,0,'ATE005120425055',NULL,NULL,55,NULL,NULL,'2025-04-05','2025-04-22','2025-04-05',250.00,'2025-05-12','2025-05-02',0,0,NULL,0,NULL,0,0,NULL,1),
	(182,NULL,15,'2025-06-14 00:00:00',76,7780.00,1500.00,NULL,2,'LSJA36U35SN051677',NULL,NULL,NULL,0,'ATE005140425054',NULL,NULL,54,NULL,NULL,'2025-04-07','2025-04-24','2025-04-07',250.00,'2025-05-14','2025-05-04',0,0,NULL,0,NULL,0,0,NULL,1),
	(183,'',15,'2025-05-28 00:00:00',5,7000.00,1500.00,NULL,2,'LSJA36U37SN051793',NULL,'2025-06-17 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'FFAU2174278',0,NULL,0,0,NULL,1),
	(184,'',15,'2025-05-28 00:00:00',NULL,7000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,NULL,0,NULL,0,0,NULL,1),
	(185,NULL,16,'2025-06-30 00:00:00',NULL,7600.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(186,NULL,16,'2025-06-26 00:00:00',65,7800.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ADM001260625122',NULL,NULL,122,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(187,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(188,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(189,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(190,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(191,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(192,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(193,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(194,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(195,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(196,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(197,'',16,NULL,NULL,7400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,97,NULL,NULL,'2025-04-02','2025-04-19','2025-04-02',NULL,'2025-05-09','2025-04-29',0,0,NULL,0,NULL,1,0,NULL,4),
	(198,NULL,16,'2025-06-17 00:00:00',242,7500.00,1500.00,NULL,2,NULL,NULL,NULL,NULL,0,'ADM001150425081','cars/198/sell_pi/sell_pi_1752826641414.pdf',NULL,81,NULL,NULL,'2025-04-08','2025-04-25','2025-04-08',250.00,'2025-05-15','2025-05-05',0,0,NULL,0,NULL,0,0,NULL,4),
	(199,'',16,'2025-05-28 00:00:00',15,7200.00,1500.00,NULL,2,'LSJA36U31SN051417',NULL,'2025-06-26 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'MRSU7439687',0,NULL,0,0,NULL,4),
	(200,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(201,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(202,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(203,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(204,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,124,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(205,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,100,NULL,NULL,'2025-03-20','2025-04-06','2025-03-20',NULL,'2025-04-26','2025-04-16',0,0,NULL,0,NULL,1,0,NULL,1),
	(206,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,100,NULL,NULL,'2025-03-20','2025-04-06','2025-03-20',NULL,'2025-04-26','2025-04-16',0,0,NULL,0,NULL,1,0,NULL,1),
	(207,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,100,NULL,NULL,'2025-03-20','2025-04-06','2025-03-20',NULL,'2025-04-26','2025-04-16',0,0,NULL,0,NULL,1,0,NULL,1),
	(208,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,96,NULL,NULL,'2025-04-22','2025-05-09','2025-04-22',NULL,'2025-05-29','2025-05-19',0,0,NULL,0,NULL,1,0,NULL,1),
	(209,'',17,NULL,NULL,12900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,96,NULL,NULL,'2025-04-22','2025-05-09','2025-04-22',NULL,'2025-05-29','2025-05-19',0,0,NULL,0,NULL,1,0,NULL,1),
	(210,'',18,'2025-05-28 00:00:00',NULL,24900.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,95,NULL,NULL,'2025-02-28','2025-03-17','2025-02-28',250.00,'2025-04-06','2025-03-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(211,'',19,'2025-05-28 00:00:00',9,10460.00,1500.00,NULL,2,'LJD4AA197S0134054',NULL,NULL,NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,NULL,0,NULL,0,0,NULL,1),
	(212,'',19,'2025-05-28 00:00:00',11,9650.00,1500.00,NULL,2,'LJD4AA19XS0133349',NULL,NULL,NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,NULL,0,NULL,0,0,NULL,1),
	(213,'',19,'2025-05-28 00:00:00',26,9650.00,1500.00,NULL,2,'LJD4AA194S0133413',NULL,NULL,NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,NULL,0,NULL,0,0,NULL,1),
	(214,'',20,'2025-05-28 00:00:00',7,17960.00,1500.00,NULL,4,'LFV2B2A14S5522067',NULL,NULL,NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,NULL,0,20,0,0,NULL,1),
	(215,'',20,'2025-05-28 00:00:00',8,17960.00,1500.00,NULL,4,'LFV2B2A15S5513846',NULL,'2025-06-18 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'MRKU3780687',0,NULL,0,0,NULL,1),
	(216,'',20,'2025-05-28 00:00:00',14,18400.00,1500.00,NULL,4,'LFV2B2A17S5513959',NULL,'2025-06-17 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'MRKU3780687',0,NULL,0,0,NULL,1),
	(217,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(218,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(219,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(220,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(221,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(222,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(223,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(224,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(225,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(226,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(227,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(228,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(229,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(230,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(231,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(232,'',21,NULL,NULL,8300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,87,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',NULL,'2025-06-06','2025-05-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(233,NULL,21,'2025-06-28 00:00:00',283,8800.00,1600.00,NULL,3,'LLV2C3B24S0005284',NULL,NULL,NULL,0,'ISH003240625118',NULL,NULL,118,NULL,NULL,'2025-06-13','2025-06-30','2025-06-13',250.00,'2025-07-20','2025-07-10',0,0,'MSDU7188538',0,21,0,0,NULL,1),
	(234,NULL,21,'2025-06-20 00:00:00',248,8800.00,1600.00,NULL,3,'LLV2C3B22S0000486',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003200625085',NULL,NULL,85,NULL,NULL,'2025-06-13','2025-06-30','2025-06-13',250.00,'2025-07-20','2025-07-10',0,0,'CSNU7709554',0,NULL,0,0,NULL,1),
	(235,NULL,55,'2025-06-19 00:00:00',246,8800.00,1600.00,NULL,3,'LLV2C3B29S0004311',NULL,NULL,NULL,0,'TOU004190625084',NULL,NULL,84,NULL,NULL,'2025-06-12','2025-06-29','2025-06-12',250.00,'2025-07-19','2025-07-09',0,0,'FFAU3202754',0,15,0,0,NULL,1),
	(236,'this client order used car not livan auto',22,'2025-06-14 00:00:00',NULL,8400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',250.00,'2025-07-14','2025-07-04',0,0,NULL,0,NULL,0,0,NULL,1),
	(237,NULL,55,'2025-06-14 00:00:00',87,8500.00,1500.00,NULL,2,'LLV2C3B28S0004722',NULL,NULL,NULL,0,'ATE005160425060',NULL,NULL,60,NULL,NULL,'2025-04-09','2025-04-26','2025-04-09',250.00,'2025-05-16','2025-05-06',0,0,'FFAU3202754',0,15,0,0,NULL,1),
	(238,NULL,55,'2025-06-24 00:00:00',274,8800.00,1600.00,NULL,3,'LLV2C3B26S0004847',NULL,NULL,NULL,0,'ISH003140625051',NULL,NULL,51,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',250.00,'2025-07-14','2025-07-04',0,0,'FFAU3202754',0,15,0,0,NULL,1),
	(239,NULL,23,'2025-06-14 00:00:00',220,8500.00,1500.00,NULL,2,'LLV2C3B25S0000367',NULL,'2025-06-24 00:00:00',NULL,0,'ATE005100425053',NULL,NULL,53,NULL,NULL,'2025-04-03','2025-04-20','2025-04-03',250.00,'2025-05-10','2025-04-30',0,0,'COSU6417451856',0,NULL,0,0,NULL,1),
	(240,NULL,23,'2025-06-14 00:00:00',217,8800.00,1600.00,NULL,3,'LLV2C3B23S0000318',NULL,'2025-06-29 00:00:00',NULL,0,'ISH003140625051',NULL,NULL,51,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',250.00,'2025-07-14','2025-07-04',0,0,'OOCU8517304',0,NULL,0,0,NULL,1),
	(241,NULL,23,'2025-06-14 00:00:00',215,8100.00,1500.00,NULL,2,'LLV2C3B2XS0001112',NULL,'2025-06-29 00:00:00',NULL,0,'ATE005080425052',NULL,NULL,52,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'COSU6417451850â€‹',0,NULL,0,0,NULL,1),
	(242,NULL,23,'2025-06-14 00:00:00',211,8500.00,1500.00,NULL,2,'LLV2C3B28S0000895',NULL,'2025-06-29 00:00:00',NULL,0,'ATE005080425049',NULL,NULL,49,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CSNU6337339',0,NULL,0,0,NULL,1),
	(243,'win this car for free',23,'2025-06-14 00:00:00',209,0.00,1500.00,NULL,2,'LLV2C3B21S0000897',NULL,'2025-06-29 00:00:00',NULL,0,'ATE005080425047','cars/243/sell_pi/sell_pi_1752826303885.pdf',NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'TGBU8695928',0,NULL,0,0,NULL,1),
	(244,NULL,23,'2025-06-14 00:00:00',208,8500.00,1500.00,NULL,2,'LLV2C3B21S0001161',NULL,'2025-05-16 00:00:00',NULL,0,'ATE005080425047',NULL,NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CCLU7612226',0,NULL,0,0,NULL,1),
	(245,NULL,23,'2025-06-14 00:00:00',207,8500.00,1500.00,NULL,2,'LLV2C3B23S0003042',NULL,'2025-06-18 00:00:00',NULL,0,'ATE005080425047','cars/245/sell_pi/sell_pi_1752826330063.pdf',NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'OOCU7424697',0,NULL,0,0,NULL,1),
	(246,NULL,23,'2025-06-14 00:00:00',206,8500.00,1500.00,NULL,2,'LLV2C3B26S0001284',NULL,'2025-05-27 00:00:00',NULL,0,'ATE005080425047',NULL,NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'PONU8180244',0,NULL,0,0,NULL,1),
	(247,NULL,23,'2025-06-14 00:00:00',205,8500.00,1500.00,NULL,2,'LLV2C3B23S0000903',NULL,'2025-05-16 00:00:00',NULL,0,'ATE005080425047',NULL,NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CCLU7324636',0,NULL,0,0,NULL,1),
	(248,NULL,23,'2025-06-14 00:00:00',204,8500.00,1500.00,NULL,2,'LLV2C3B20S0000888',NULL,'2025-05-16 00:00:00',NULL,0,'ATE005080425047',NULL,NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CSNU8512132',0,NULL,0,0,NULL,1),
	(249,NULL,23,'2025-06-14 00:00:00',203,8500.00,1500.00,NULL,2,'LLV2C3B22S0000908',NULL,'2025-05-16 00:00:00',NULL,0,'ATE005080425047',NULL,NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CSNU8512132',0,NULL,0,0,NULL,1),
	(250,NULL,23,'2025-06-14 00:00:00',202,8500.00,1500.00,NULL,2,'LLV2C3B25S0000899',NULL,'2025-05-15 00:00:00',NULL,0,'ATE005080425047','cars/250/sell_pi/sell_pi_1752825995437.pdf',NULL,47,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',0,0,'CCLU7324636',0,NULL,0,0,NULL,1),
	(251,NULL,23,'2025-06-14 00:00:00',200,8500.00,1500.00,NULL,2,'LLV2C3B25S0001468',NULL,'2025-06-29 00:00:00',NULL,0,'ATE005040425046',NULL,NULL,46,NULL,NULL,'2025-03-28','2025-04-14','2025-03-28',250.00,'2025-05-04','2025-04-24',0,0,'FCIU9463135',0,NULL,0,0,NULL,1),
	(252,NULL,23,'2025-06-14 00:00:00',199,8420.00,1500.00,NULL,2,'LLV2C3B28S0001173',NULL,'2025-05-30 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'CSNU6865613',0,NULL,0,0,NULL,1),
	(253,NULL,23,'2025-06-14 00:00:00',197,8420.00,1500.00,NULL,2,'LLV2C3B23S0001131',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'TRHU6571984',0,NULL,0,0,NULL,1),
	(254,NULL,23,'2025-06-14 00:00:00',196,8420.00,1500.00,NULL,2,'LLV2C3B22S0001153',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'TRHU6571984',0,NULL,0,0,NULL,1),
	(255,NULL,23,'2025-06-14 00:00:00',195,8420.00,1500.00,NULL,2,'LLV2C3B2XS0000767',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'OOCU6814788',0,NULL,0,0,NULL,1),
	(256,NULL,23,'2025-06-14 00:00:00',194,8420.00,1500.00,NULL,2,'LLV2C3B26S0000779',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'OOCU6814788',0,NULL,0,0,NULL,1),
	(257,NULL,23,'2025-06-14 00:00:00',193,8420.00,1500.00,NULL,2,'LLV2C3B21S0000768',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'OOCU6814788',0,NULL,0,0,NULL,1),
	(258,NULL,23,'2025-06-14 00:00:00',192,8420.00,1500.00,NULL,2,'LLV2C3B25S0000806',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'TRHU6571984',0,NULL,0,0,NULL,1),
	(259,NULL,23,'2025-06-14 00:00:00',107,8420.00,1500.00,NULL,2,'LLV2C3B29S0001151',NULL,'2025-05-16 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'TRHU6571984',0,NULL,0,0,NULL,1),
	(260,NULL,23,'2025-06-14 00:00:00',187,8500.00,1500.00,NULL,2,'LLV2C3B2XS0000901',NULL,'2025-04-08 00:00:00',NULL,0,'MOH006220425044',NULL,NULL,44,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'TGBU8695928',0,NULL,0,0,NULL,1),
	(261,NULL,23,'2025-06-14 00:00:00',189,8500.00,1500.00,NULL,2,'LLV2C3B26S0000894',NULL,'2025-04-08 00:00:00',NULL,0,'MOH006160425043',NULL,NULL,43,NULL,NULL,'2025-04-09','2025-04-26','2025-04-09',250.00,'2025-05-16','2025-05-06',0,0,'TGBU8695928',0,NULL,0,0,NULL,1),
	(262,NULL,23,'2025-06-14 00:00:00',190,8500.00,1500.00,NULL,2,'LLV2C3B23S0000898',NULL,'2025-04-08 00:00:00',NULL,0,'MOH006160425043',NULL,NULL,43,NULL,NULL,'2025-04-09','2025-04-26','2025-04-09',250.00,'2025-05-16','2025-05-06',0,0,'TGBU8695928',0,NULL,0,0,NULL,1),
	(263,NULL,23,'2025-06-14 00:00:00',92,8500.00,1500.00,NULL,2,'LLV2C3B20S0001393','cars/263/documents/documents_1750871842625.pdf','2025-06-25 00:00:00',NULL,0,'MOH006190425042',NULL,NULL,42,NULL,NULL,'2025-04-12','2025-04-29','2025-04-12',250.00,'2025-05-19','2025-05-09',0,0,'PONU8180244',0,NULL,0,0,NULL,1),
	(264,NULL,23,'2025-06-14 00:00:00',185,8500.00,1500.00,NULL,2,'LLV2C3B24S0001333','cars/264/documents/documents_1750871876574.pdf','2025-06-25 00:00:00',NULL,0,'MOH006190425042',NULL,NULL,42,NULL,NULL,'2025-04-12','2025-04-29','2025-04-12',250.00,'2025-05-19','2025-05-09',0,0,'COSU6453342960',0,NULL,0,0,NULL,1),
	(265,NULL,23,'2025-06-14 00:00:00',90,8500.00,1500.00,NULL,2,'LLV2C3B26S0001530',NULL,'2025-06-11 00:00:00',NULL,0,'MOH006190425042',NULL,NULL,42,NULL,NULL,'2025-04-12','2025-04-29','2025-04-12',250.00,'2025-05-19','2025-05-09',0,0,'TCNU5535823',0,NULL,0,0,NULL,1),
	(266,NULL,23,'2025-06-13 00:00:00',45,8500.00,1500.00,NULL,2,'LLV2C3B24S0003065',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003210425041',NULL,NULL,41,NULL,NULL,'2025-04-14','2025-05-01','2025-04-14',250.00,'2025-05-21','2025-05-11',0,0,'COSU6420944201',0,NULL,0,0,NULL,1),
	(267,NULL,23,'2025-06-13 00:00:00',181,8420.00,1500.00,NULL,2,'LLV2C3B28S0001187',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451853',0,NULL,0,0,NULL,1),
	(268,NULL,23,'2025-06-13 00:00:00',182,8420.00,1500.00,NULL,2,'LLV2C3B2XS0001188',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451853',0,NULL,0,0,NULL,1),
	(269,NULL,23,'2025-06-13 00:00:00',183,8420.00,1500.00,NULL,2,'LLV2C3B29S0001103',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040','cars/269/sell_pi/sell_pi_1752826521015.pdf',NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451853',0,NULL,0,0,NULL,1),
	(270,NULL,23,'2025-06-13 00:00:00',99,8420.00,1500.00,NULL,2,'LLV2C3B27S0001178',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451852D',0,NULL,0,0,NULL,1),
	(271,NULL,23,'2025-06-13 00:00:00',100,8420.00,1500.00,NULL,2,'LLV2C3B28S0001142',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451857',0,NULL,0,0,NULL,1),
	(272,NULL,23,'2025-06-13 00:00:00',98,8420.00,1500.00,NULL,2,'LLV2C3B22S0001122',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451852B',0,NULL,0,0,NULL,1),
	(273,NULL,23,'2025-06-13 00:00:00',103,8420.00,1500.00,NULL,2,'LLV2C3B24S0001171',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006230425040',NULL,NULL,40,NULL,NULL,'2025-04-16','2025-05-03','2025-04-16',250.00,'2025-05-23','2025-05-13',0,0,'COSU6417451852C',0,NULL,0,0,NULL,1),
	(274,NULL,23,'2025-06-13 00:00:00',41,8500.00,1500.00,NULL,2,'LLV2C3B25S0001485',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003210425039',NULL,NULL,39,NULL,NULL,'2025-04-14','2025-05-01','2025-04-14',250.00,'2025-05-21','2025-05-11',0,0,'COSU6417567830D',0,NULL,0,0,NULL,1),
	(275,'on sea car order',24,'2025-07-01 00:00:00',304,14000.00,1600.00,NULL,3,'LB37622Z0SX621548','cars/275/documents/documents_1751359123583.pdf','2025-07-01 00:00:00',NULL,0,'ADM001010725132',NULL,NULL,132,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'COSU6422362391',0,NULL,0,0,NULL,5),
	(276,'fright not payed',24,'2025-06-13 00:00:00',108,12800.00,2000.00,NULL,2,'LB37622Z5SX621559','cars/276/documents/documents_1751358823378.pdf','2025-07-01 00:00:00',NULL,0,'MOH006130625006',NULL,NULL,6,NULL,NULL,'2025-06-06','2025-06-23','2025-06-06',250.00,'2025-07-13','2025-07-03',0,0,'COSU6422362390',0,NULL,0,0,NULL,5),
	(277,NULL,25,'2025-06-28 00:00:00',124,9000.00,1600.00,NULL,3,'LB37622Z6SX819387','cars/277/documents/documents_1751207941937.pdf',NULL,NULL,0,'ADM001280525123',NULL,NULL,123,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TRHU7452784',0,3,0,0,NULL,1),
	(278,NULL,25,'2025-06-13 00:00:00',146,9000.00,1600.00,NULL,2,'Lb37622z3sx819394','cars/278/documents/documents_1751360796909.pdf','2025-07-01 00:00:00',NULL,0,'ISH003190525017',NULL,NULL,17,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,NULL,0,NULL,0,0,NULL,1),
	(279,NULL,25,'2025-06-13 00:00:00',162,9900.00,1600.00,NULL,2,'LB37622Z5SX819400','cars/279/documents/documents_1751195246644.pdf',NULL,NULL,0,'ISH003100525028',NULL,NULL,28,NULL,NULL,'2025-05-03','2025-05-20','2025-05-03',250.00,'2025-06-09','2025-05-30',0,0,'6422944000',0,2,0,0,NULL,1),
	(280,NULL,25,'2025-06-13 00:00:00',148,9000.00,1600.00,NULL,2,'LB37622ZXSX819358','cars/280/documents/documents_1751195290077.pdf',NULL,NULL,0,'ISH003190525019',NULL,NULL,19,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'6422944000',0,2,0,0,NULL,1),
	(281,NULL,25,'2025-06-13 00:00:00',151,9000.00,1600.00,NULL,3,'LB37622Z5SX819395','cars/281/documents/documents_1751194316291.pdf','2025-06-29 00:00:00',NULL,0,'ISH003190525022',NULL,NULL,22,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'6422944000',0,2,0,0,NULL,1),
	(282,NULL,25,'2025-06-13 00:00:00',149,9000.00,1600.00,NULL,2,'LB37622Z8SX819391','cars/282/documents/documents_1751195112313.pdf',NULL,NULL,0,'ISH003190525020',NULL,NULL,20,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'6422944000',0,2,0,0,NULL,1),
	(283,NULL,25,'2025-06-13 00:00:00',147,9000.00,1600.00,NULL,2,'LB37622Z7SX819382','cars/283/documents/documents_1751195319561.pdf',NULL,NULL,0,'ISH003190525018',NULL,NULL,18,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'TRHU7452784',0,3,0,0,NULL,1),
	(284,NULL,25,'2025-06-13 00:00:00',125,9000.00,1600.00,NULL,3,'LB37622Z3SX819329','cars/284/documents/documents_1751195531791.pdf',NULL,NULL,0,'ISH003190525016',NULL,NULL,16,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'TRHU7452784',0,3,0,0,NULL,1),
	(285,NULL,25,'2025-06-13 00:00:00',142,9000.00,1600.00,NULL,3,'LB37622Z2SX815062',NULL,NULL,NULL,0,'ISH003210525011',NULL,NULL,11,NULL,NULL,'2025-05-14','2025-05-31','2025-05-14',250.00,'2025-06-20','2025-06-10',0,0,'MRSU5017091',0,6,0,0,NULL,1),
	(286,NULL,25,'2025-06-13 00:00:00',28,9000.00,1600.00,NULL,3,'LB37622Z8SX815048',NULL,NULL,NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'MRSU5017091',0,6,0,0,NULL,1),
	(287,NULL,25,'2025-06-13 00:00:00',29,9000.00,1600.00,NULL,3,'LB37622Z2SX815000',NULL,NULL,NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'MRSU5017091',0,6,0,0,NULL,1),
	(288,NULL,25,'2025-06-13 00:00:00',30,9000.00,1600.00,NULL,3,'LB37622ZXSX816783',NULL,NULL,NULL,0,'MAS010250525010','cars/288/sell_pi/sell_pi_1752826030493.pdf',NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FFAU3368344',0,30,0,0,NULL,1),
	(289,NULL,25,'2025-06-13 00:00:00',141,9000.00,1600.00,NULL,3,'LB37622ZXSX816492',NULL,NULL,NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FFAU3368344',0,30,0,0,NULL,1),
	(290,NULL,25,'2025-06-13 00:00:00',140,9000.00,1600.00,NULL,3,'LB37622Z6SX816344','cars/290/documents/documents_1751957057027.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'TIIU5347213',0,10,0,0,NULL,1),
	(291,NULL,25,'2025-06-13 00:00:00',32,9000.00,1600.00,NULL,3,'LB37622Z3SX816544','cars/291/documents/documents_1751957105973.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'TIIU5347213',0,10,0,0,NULL,1),
	(292,NULL,25,'2025-06-13 00:00:00',137,9000.00,1600.00,NULL,3,'LB37622Z5SX816738',NULL,NULL,NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'TIIU5347213',0,10,0,0,NULL,1),
	(293,NULL,25,'2025-06-13 00:00:00',136,9000.00,1600.00,NULL,3,'LB37622268X816151','cars/293/documents/documents_1751957204008.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010','cars/293/sell_pi/sell_pi_1752826060979.pdf',NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'TIIU5347213',0,10,0,0,NULL,1),
	(294,NULL,25,'2025-06-13 00:00:00',33,9000.00,1600.00,NULL,3,'LB37622Z8SX816779','cars/294/documents/documents_1751954467374.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FCIU9681630',0,11,0,0,NULL,1),
	(295,NULL,25,'2025-06-13 00:00:00',34,9100.00,1500.00,NULL,3,'LB37622Z6SX816540','cars/295/documents/documents_1751954184208.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010','cars/295/sell_pi/sell_pi_1752826134602.pdf',NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FCIU9681630',0,11,0,0,NULL,1),
	(296,NULL,25,'2025-06-13 00:00:00',134,9000.00,1600.00,NULL,3,'B37622Z3SX816740','cars/296/documents/documents_1751957365977.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FCIU9681630',0,11,0,0,NULL,1),
	(297,NULL,25,'2025-06-13 00:00:00',35,9000.00,1600.00,NULL,3,'LB37622248X816021','cars/297/documents/documents_1751957449710.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FCIU9681630',0,11,0,0,NULL,1),
	(298,'',25,'2025-05-28 00:00:00',19,9100.00,1500.00,NULL,4,'LB37622ZOSX816520','cars/298/documents/documents_1751954980722.pdf','2025-07-08 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'FFAU2255430',0,8,0,0,NULL,1),
	(299,NULL,26,'2025-06-30 00:00:00',164,9000.00,1600.00,NULL,3,'LB37622Z3SX805446','cars/299/documents/documents_1751360923083.pdf','2025-07-01 00:00:00',NULL,0,'MOH006160525029',NULL,NULL,29,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'COSU6422364621',0,NULL,0,0,NULL,4),
	(300,NULL,26,'2025-06-24 00:00:00',252,9400.00,1600.00,NULL,3,'LB37622Z6SX805635',NULL,NULL,NULL,0,'ADM001120625116',NULL,NULL,116,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(301,NULL,26,'2025-06-13 00:00:00',46,9000.00,1600.00,NULL,2,'LB37622Z7SX804459','cars/301/documents/documents_1751360858653.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422364620',0,NULL,0,0,NULL,4),
	(302,NULL,26,'2025-06-13 00:00:00',49,9000.00,1600.00,NULL,2,'LB37622Z1SX815943','cars/302/documents/documents_1751361409838.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422365933',0,NULL,0,0,NULL,4),
	(303,NULL,26,'2025-06-13 00:00:00',48,9000.00,1600.00,NULL,2,'LB37622Z8SX815924','cars/303/documents/documents_1751361348240.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422365932',0,NULL,0,0,NULL,4),
	(304,NULL,26,'2025-06-13 00:00:00',50,9000.00,1600.00,NULL,2,'LB37622Z9SX815933','cars/304/documents/documents_1751361238517.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422365931',0,NULL,0,0,NULL,4),
	(305,NULL,26,'2025-06-13 00:00:00',51,9000.00,1600.00,NULL,2,'LB37622Z3SX805625','cars/305/documents/documents_1751361173017.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422365930',0,NULL,0,0,NULL,4),
	(306,NULL,26,'2025-06-13 00:00:00',47,9000.00,1600.00,NULL,2,'LB37622Z9SX804446','cars/306/documents/documents_1751361093556.pdf','2025-07-01 00:00:00',NULL,0,'MOH006230525008',NULL,NULL,8,NULL,NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'COSU6422364623',0,NULL,0,0,NULL,4),
	(307,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(308,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(309,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(310,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(311,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(312,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(313,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(314,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(315,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(316,'',27,NULL,NULL,13000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,89,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(317,NULL,28,'2025-06-14 00:00:00',109,14500.00,1500.00,NULL,2,'LSV5J60N7SN011551','cars/317/documents/documents_1750869676483.pdf','2025-06-24 00:00:00',NULL,0,'ATE005010525062',NULL,NULL,62,NULL,NULL,'2025-04-24','2025-05-11','2025-04-24',250.00,'2025-05-31','2025-05-21',0,0,'MSKU0077137',0,NULL,0,0,NULL,1),
	(318,NULL,28,'2025-06-13 00:00:00',170,14700.00,1500.00,NULL,2,'LSV5J60N3SN011482','cars/318/documents/documents_1750869529688.pdf','2025-06-25 00:00:00',NULL,0,'MOH006220425033',NULL,NULL,33,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'MSKU0077137',0,NULL,0,0,NULL,1),
	(319,NULL,28,'2025-06-13 00:00:00',173,14700.00,1500.00,NULL,2,'LSV5J60N3SN011448','cars/319/documents/documents_1750869382692.pdf','2025-06-25 00:00:00',NULL,0,'MOH006220425033',NULL,NULL,33,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'MSKU0077137',0,NULL,0,0,NULL,1),
	(320,NULL,28,'2025-06-13 00:00:00',163,14700.00,1500.00,NULL,2,'LSV5J60N7SN011582','cars/320/documents/documents_1750869324427.pdf','2025-06-25 00:00:00',NULL,0,'MOH006220425030',NULL,NULL,30,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'MSKU0077137',0,NULL,0,0,NULL,1),
	(321,NULL,28,'2025-06-13 00:00:00',27,14300.00,1500.00,NULL,3,'LSV5J60NXSN011592',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006210525009',NULL,NULL,9,NULL,NULL,'2025-05-14','2025-05-31','2025-05-14',250.00,'2025-06-20','2025-06-10',0,0,'TLLU5990047',0,NULL,0,0,NULL,1),
	(322,NULL,29,'2025-07-10 00:00:00',72,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001010425094',NULL,NULL,94,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(326,NULL,29,'2025-07-10 00:00:00',72,NULL,NULL,NULL,NULL,'LJD5AA1D3S0212645',NULL,NULL,NULL,0,'ADM001010425094',NULL,NULL,94,NULL,NULL,'2025-06-09','2025-06-26','2025-06-09',NULL,'2025-07-16','2025-07-06',0,0,'OOCU7143128',0,24,0,0,NULL,2),
	(328,NULL,29,'2025-06-30 00:00:00',284,13200.00,1600.00,NULL,3,'LJD5AA1D3S0203427',NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'CAIU7094512',0,13,0,0,NULL,1),
	(329,NULL,29,'2025-06-13 00:00:00',152,13200.00,1600.00,NULL,3,'LJD5AA1D1S0202518',NULL,'2025-06-17 00:00:00',NULL,0,'ISH003050525025',NULL,NULL,25,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',250.00,'2025-06-04','2025-05-25',0,0,'MRSU3484672',0,NULL,0,0,NULL,1),
	(330,NULL,29,'2025-06-12 00:00:00',85,13700.00,1500.00,NULL,3,'LJD5AA1D0S0202512',NULL,'2025-07-10 00:00:00',NULL,0,'ISH003120625005','cars/330/sell_pi/sell_pi_1752825923402.pdf',NULL,5,'',NULL,'2025-06-05',NULL,'2025-06-05',250.00,NULL,'2025-07-02',0,0,'MRSU3484672',0,NULL,0,0,NULL,1),
	(331,NULL,29,'2025-06-12 00:00:00',84,13900.00,1500.00,NULL,3,'LJD5AA1D0S0202526',NULL,'2025-07-04 00:00:00',NULL,0,'ISH003120625003',NULL,NULL,3,NULL,NULL,'2025-06-05','2025-06-22','2025-06-05',250.00,'2025-07-12','2025-07-02',0,0,'CSNU8157038',0,NULL,0,0,NULL,1),
	(332,'this is not loaded at all',30,'2025-06-12 00:00:00',68,8500.00,1500.00,NULL,2,'LLV2C3B29S0001361',NULL,'2025-06-18 00:00:00',NULL,0,'MOH006250525034',NULL,NULL,34,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FSCU9349902',0,NULL,0,0,NULL,1),
	(333,NULL,30,'2025-06-13 00:00:00',175,8500.00,1500.00,NULL,2,'LLV2C3B24S0000893',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425033','cars/333/sell_pi/sell_pi_1752826683236.pdf',NULL,33,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417469870',0,NULL,0,0,NULL,1),
	(334,NULL,30,'2025-06-13 00:00:00',174,8500.00,1500.00,NULL,2,'LLV2C3B20S0002849',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425033',NULL,NULL,33,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'TGBU7905296',0,NULL,0,0,NULL,1),
	(335,NULL,30,'2025-06-13 00:00:00',172,8500.00,1500.00,NULL,2,'LLV2C3B2XS0000896',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425032',NULL,NULL,32,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417469870',0,NULL,0,0,NULL,1),
	(336,NULL,30,'2025-06-13 00:00:00',171,8500.00,1500.00,NULL,2,'LLV2C3B25S0001129',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425032',NULL,NULL,32,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417451856',0,NULL,0,0,NULL,1),
	(337,NULL,30,'2025-06-13 00:00:00',169,8500.00,1500.00,NULL,2,'LLV2C3B20S0000891',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425031',NULL,NULL,31,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417469840',0,NULL,0,0,NULL,1),
	(338,NULL,30,'2025-06-13 00:00:00',168,8500.00,1500.00,NULL,2,'LLV2C3B29S0000887',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425031',NULL,NULL,31,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417469840',0,NULL,0,0,NULL,1),
	(339,'',30,'2025-05-28 00:00:00',6,8000.00,1500.00,NULL,4,'LLV2C3B20S0000230','cars/339/documents/documents_1750866823224.pdf','2025-06-25 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'COSU6415809610',0,NULL,0,0,NULL,1),
	(350,NULL,31,'2025-07-15 00:00:00',301,8292.00,1600.00,NULL,2,'LB37824S2SG058321',NULL,NULL,NULL,0,'ISH003050725141',NULL,NULL,141,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'OOLU9843750',0,19,0,0,'2025-07-15 14:27:32',4),
	(351,NULL,31,'2025-07-15 00:00:00',299,8292.00,1600.00,NULL,2,'LB37824S0SG058978',NULL,NULL,NULL,0,'ISH003030725138',NULL,NULL,138,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'OOLU9843750',0,19,0,0,'2025-07-15 14:36:50',4),
	(352,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,'LB37824S9SG061801',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'FFAU2465458',0,31,0,0,NULL,1),
	(353,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,'LB37824S3RG142772',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'FSCU8476604',0,29,0,0,NULL,1),
	(354,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,'LB37824S8SG061790',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'FSCU8476604',0,29,0,0,NULL,1),
	(355,NULL,31,'2025-07-10 00:00:00',311,8400.00,1600.00,NULL,3,'LB37824S8SG048022',NULL,NULL,NULL,0,'MOH006100725151',NULL,NULL,151,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'OOLU9843750',0,19,0,0,NULL,1),
	(356,NULL,31,'2025-07-21 00:00:00',382,8400.00,1600.00,NULL,2,'LB37824S2SG047996',NULL,NULL,NULL,0,'ISH003210725192',NULL,NULL,192,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FSCU8476604',0,29,0,0,'2025-07-21 12:50:58',1),
	(357,NULL,31,'2025-07-19 00:00:00',360,8400.00,1600.00,NULL,3,'LB37824S2SG030812',NULL,NULL,NULL,0,'ISH003190725184',NULL,NULL,184,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FFAU2465458',0,31,0,0,'2025-07-19 04:26:44',1),
	(358,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(359,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(360,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(361,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(362,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(363,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(364,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(365,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(366,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(367,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(368,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(369,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(370,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(371,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(372,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(373,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(374,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(375,'',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(376,'grey color not white',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(377,'grey color not white',31,NULL,NULL,8500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,91,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',0,0,NULL,0,NULL,1,0,NULL,1),
	(378,NULL,31,'2025-06-14 00:00:00',233,7900.00,1500.00,NULL,2,'LB37824S8RG144081',NULL,NULL,NULL,0,'ATE005070525064','cars/378/sell_pi/sell_pi_1752826486960.pdf',NULL,64,NULL,NULL,'2025-04-30','2025-05-17','2025-04-30',250.00,'2025-06-06','2025-05-27',0,0,'OOLU9843750',0,19,0,0,NULL,1),
	(379,'used emgrand',31,'2025-06-14 00:00:00',NULL,10820.00,NULL,NULL,NULL,'LB37824S2RG144061',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,'2025-04-08','2025-04-25','2025-04-08',250.00,'2025-05-15','2025-05-05',0,0,'FSCU8476604',0,29,0,0,NULL,1),
	(380,'with black pack',32,'2025-06-14 00:00:00',191,18700.00,1700.00,NULL,3,'LFV2B2A14S5523834','cars/380/documents/documents_1751359069918.pdf','2025-07-01 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'COSU6422362392',0,NULL,0,0,NULL,6),
	(381,NULL,32,'2025-06-14 00:00:00',198,18400.00,1700.00,NULL,3,'LFV2B2A11S5519336','cars/381/documents/documents_1751359185864.pdf','2025-07-01 00:00:00',NULL,0,'MOH006280425045',NULL,NULL,45,NULL,NULL,'2025-04-21','2025-05-08','2025-04-21',250.00,'2025-05-28','2025-05-18',0,0,'COSU6422362393',0,NULL,0,0,NULL,6),
	(382,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(383,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(384,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(385,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(386,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(387,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(388,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(389,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(390,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(391,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(392,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(393,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(394,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(395,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(396,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(397,'',33,NULL,NULL,11200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(398,NULL,33,'2025-07-17 00:00:00',56,11200.00,1600.00,NULL,3,'LFV2B2C11N3774047',NULL,NULL,NULL,0,'ATE005210625181',NULL,NULL,181,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',250.00,'2025-07-14','2025-07-04',1,0,'OOCU6816693',0,26,0,0,NULL,1),
	(399,NULL,33,'2025-06-14 00:00:00',NULL,11280.00,NULL,NULL,NULL,'LFV2B2C10P3701318',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',250.00,'2025-07-14','2025-07-04',1,0,'FFAU2465458',0,31,0,0,NULL,1),
	(400,NULL,33,'2025-07-20 00:00:00',364,11691.14,1600.00,NULL,3,'LFV2B2C11P3700310',NULL,NULL,NULL,0,'ADM001200725186',NULL,NULL,186,NULL,NULL,'2025-06-07','2025-06-24','2025-06-07',237.00,'2025-07-14','2025-07-04',1,0,'FFAU2465458',0,31,0,0,'2025-07-20 03:06:23',1),
	(401,NULL,48,'2025-06-13 00:00:00',120,11100.00,1600.00,NULL,3,'LFV2B2C13P3700311',NULL,NULL,NULL,0,'ISH003170525024',NULL,NULL,24,NULL,NULL,'2025-05-10','2025-05-27','2025-05-10',250.00,'2025-06-16','2025-06-06',1,0,'CSNU7094969',0,28,0,0,NULL,1),
	(404,NULL,49,'2025-06-14 00:00:00',213,15700.00,1600.00,NULL,4,'LSVYSAC12SN056967',NULL,'2025-06-25 00:00:00',NULL,0,'ATE005220525073',NULL,NULL,73,NULL,NULL,'2025-05-15','2025-06-01','2025-05-15',250.00,'2025-06-21','2025-06-11',0,0,'TRHU8440448',0,NULL,0,0,NULL,4),
	(405,NULL,49,'2025-06-14 00:00:00',214,15700.00,1600.00,NULL,4,'LSVYSAC15SN056901',NULL,'2025-06-25 00:00:00',NULL,0,'ATE005220625072',NULL,NULL,72,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',0,0,'TRHU8440448',0,NULL,0,0,NULL,4),
	(406,NULL,49,'2025-06-13 00:00:00',23,15600.00,1600.00,NULL,3,'LSVYSAC18SN056939',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003100525023',NULL,NULL,23,NULL,NULL,'2025-05-03','2025-05-20','2025-05-03',250.00,'2025-06-09','2025-05-30',0,0,'TRHU8440448',0,NULL,0,0,NULL,4),
	(417,'on sea car order',36,'2025-07-01 00:00:00',303,13800.00,1600.00,NULL,3,'LSVYNAC12SN060766','cars/417/documents/documents_1751359694986.pdf','2025-07-01 00:00:00',NULL,0,'ADM001010725132',NULL,NULL,132,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'COSU6422363970',0,NULL,0,0,NULL,4),
	(418,'on sea car order',36,'2025-07-01 00:00:00',317,13800.00,1600.00,NULL,3,'LSVYNAC19SN060781','cars/418/documents/documents_1751359531169.pdf','2025-07-01 00:00:00',NULL,0,'ADM001010725132',NULL,NULL,132,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'COSU6422363233',1,NULL,0,0,NULL,4),
	(419,NULL,36,'2025-06-13 00:00:00',144,13700.00,1600.00,NULL,3,'LSVYNAC19SN058013','cars/419/documents/documents_1751359400096.pdf','2025-07-01 00:00:00',NULL,0,'ISH003190525012',NULL,NULL,12,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6422363232',0,NULL,0,0,NULL,4),
	(420,'disc',36,'2025-05-28 00:00:00',17,13700.00,1500.00,NULL,4,'LSVYNAC15SN057750','cars/420/documents/documents_1751359291030.pdf','2025-07-01 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'COSU6422363231',0,NULL,0,0,NULL,4),
	(421,'disc',36,'2025-05-28 00:00:00',16,13700.00,1500.00,NULL,4,'LSVYNAC17SN057751','cars/421/documents/documents_1751359244247.pdf','2025-07-01 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'COSU6422363230',0,NULL,0,0,NULL,4),
	(447,'i get discount ',37,NULL,NULL,9100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,93,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',NULL,'2025-06-04','2025-05-25',0,0,NULL,0,NULL,1,0,NULL,1),
	(448,'i get discount ',37,NULL,NULL,9100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,93,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',NULL,'2025-06-04','2025-05-25',0,0,NULL,0,NULL,1,0,NULL,1),
	(449,NULL,37,'2025-06-13 00:00:00',NULL,9000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,93,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',250.00,'2025-06-04','2025-05-25',0,0,NULL,0,NULL,1,0,NULL,1),
	(450,'i get discount ',37,NULL,NULL,9100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,93,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',NULL,'2025-06-04','2025-05-25',0,0,NULL,0,NULL,1,0,NULL,1),
	(451,NULL,37,'2025-06-13 00:00:00',NULL,9000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,93,NULL,NULL,'2025-04-28','2025-05-15','2025-04-28',250.00,'2025-06-04','2025-05-25',0,0,NULL,0,NULL,1,0,NULL,1),
	(452,NULL,38,'2025-06-13 00:00:00',36,9000.00,1600.00,NULL,3,'LB37622Z6SX816571','cars/452/documents/documents_1751954819076.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,NULL,NULL,0,0,'SEKU4542148',0,12,0,0,NULL,1),
	(453,NULL,38,'2025-06-13 00:00:00',25,9000.00,1600.00,NULL,3,'LB37622Z6SX816781','cars/453/documents/documents_1751954882130.pdf','2025-07-08 00:00:00',NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'SEKU4542148',0,12,0,0,NULL,1),
	(454,NULL,38,'2025-06-13 00:00:00',24,9000.00,1600.00,NULL,3,'LB37622Z7SX816532',NULL,NULL,NULL,0,'MAS010250525010',NULL,NULL,10,NULL,NULL,'2025-05-18','2025-06-04','2025-05-18',250.00,'2025-06-24','2025-06-14',0,0,'FFAU3368344',0,30,0,0,NULL,1),
	(455,'disc',38,'2025-05-28 00:00:00',18,9100.00,1500.00,NULL,4,'LB37622Z9SX819366','cars/455/documents/documents_1751195914487.pdf',NULL,NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,'2025-05-02','2025-05-19','2025-05-02',250.00,'2025-06-08','2025-05-29',0,0,'TRHU7452784',0,3,0,0,NULL,1),
	(456,'not under ahmed name',40,'2025-06-12 00:00:00',105,8500.00,1500.00,NULL,3,'LLV2C3B22S0000391',NULL,'2025-06-29 00:00:00',NULL,0,'MOH006220425030',NULL,NULL,30,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'OOCU8517304',0,NULL,0,0,NULL,1),
	(457,NULL,40,'2025-06-13 00:00:00',167,8500.00,1500.00,NULL,2,'LLV2C3B22S0000892',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425030',NULL,NULL,30,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'COSU6417469840',0,NULL,0,0,NULL,1),
	(458,NULL,40,'2025-06-13 00:00:00',166,8500.00,1500.00,NULL,2,'LLV2C3B27S0003142',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006220425030',NULL,NULL,30,NULL,NULL,'2025-04-15','2025-05-02','2025-04-15',250.00,'2025-05-22','2025-05-12',0,0,'TGBU7905296',0,NULL,0,0,NULL,1),
	(459,NULL,40,'2025-06-13 00:00:00',80,8300.00,1500.00,NULL,2,'LLV2C3B26S0003116',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190425027',NULL,NULL,27,NULL,NULL,'2025-04-12','2025-04-29','2025-04-12',250.00,'2025-05-19','2025-05-09',0,0,'OOCU7424697',0,NULL,0,0,NULL,1),
	(460,NULL,40,'2025-06-13 00:00:00',43,8300.00,1500.00,NULL,2,'LLV2C3B25S0001499',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417567830C',0,NULL,0,0,NULL,1),
	(461,NULL,40,'2025-06-13 00:00:00',160,8300.00,1500.00,NULL,2,'LLV2C3B21S0003136',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'OOCU9215356',0,NULL,0,0,NULL,1),
	(462,NULL,40,'2025-06-13 00:00:00',39,8300.00,1500.00,NULL,2,'LLV2C3B25S0001485',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417567830D',0,NULL,0,0,NULL,1),
	(463,NULL,40,'2025-06-13 00:00:00',37,8300.00,1500.00,NULL,2,'LLV2C3B2XS0001594',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417567540D',0,NULL,0,0,NULL,1),
	(464,NULL,40,'2025-06-13 00:00:00',38,8300.00,1500.00,NULL,2,'LLV2C3B25S0001504',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,NULL,0,NULL,0,0,NULL,1),
	(465,NULL,40,'2025-06-13 00:00:00',157,8300.00,1500.00,NULL,2,'LLV2C3B24S0003020',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'FFAU2733798',0,NULL,0,0,NULL,1),
	(466,NULL,40,'2025-06-13 00:00:00',156,8300.00,1500.00,NULL,2,'LLV2C3B23S0003011',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'FFAU2733798',0,NULL,0,0,NULL,1),
	(467,NULL,40,'2025-06-13 00:00:00',155,8300.00,1500.00,NULL,2,'LLV2C3B25S0003043',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'FFAU2733798',0,NULL,0,0,NULL,1),
	(468,NULL,40,'2025-06-13 00:00:00',154,8300.00,1500.00,NULL,2,'LLV2C3B25S0003026',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026','cars/468/sell_pi/sell_pi_1752826356242.pdf',NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'FFAU2733798',0,NULL,0,0,NULL,1),
	(469,NULL,40,'2025-06-13 00:00:00',40,8300.00,1500.00,NULL,2,'LLV2C3B25S0001454',NULL,'2025-06-25 00:00:00',NULL,0,'ISH003190525026',NULL,NULL,26,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417567830B',0,NULL,0,0,NULL,1),
	(470,NULL,40,'2025-06-13 00:00:00',68,8800.00,1600.00,NULL,3,'LLV2C3B25S0000904',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006190525015',NULL,NULL,15,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'COSU6417469870',0,NULL,0,0,NULL,1),
	(471,NULL,40,'2025-06-13 00:00:00',71,8800.00,1600.00,NULL,3,'LLV2C3B25S0003057',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006190525015',NULL,NULL,15,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'OOCU9215356',0,NULL,0,0,NULL,1),
	(472,NULL,40,'2025-06-13 00:00:00',70,8800.00,1600.00,NULL,3,'LLV2C3B23S0003056',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006190525015',NULL,NULL,15,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'OOCU9215356',0,NULL,0,0,NULL,1),
	(473,NULL,40,'2025-06-13 00:00:00',145,8900.00,1500.00,NULL,3,'LLV2C3B2XS0003135',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006170525013',NULL,NULL,13,NULL,NULL,'2025-05-10','2025-05-27','2025-05-10',250.00,'2025-06-16','2025-06-06',0,0,'OOCU7424697',0,NULL,0,0,NULL,1),
	(474,NULL,40,NULL,86,7700.00,NULL,NULL,3,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,'2025-06-05','2025-06-22','2025-06-05',NULL,'2025-07-12','2025-07-02',0,0,NULL,0,NULL,1,0,NULL,1),
	(475,NULL,40,NULL,86,7700.00,0.00,NULL,3,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,'2025-06-05','2025-06-22','2025-06-05',NULL,'2025-07-12','2025-07-02',0,0,NULL,0,NULL,1,0,NULL,1),
	(476,NULL,40,NULL,86,7700.00,0.00,NULL,3,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,'2025-06-05','2025-06-22','2025-06-05',NULL,'2025-07-12','2025-07-02',0,0,NULL,0,NULL,1,0,NULL,1),
	(477,NULL,55,NULL,86,7700.00,0.00,NULL,3,NULL,NULL,'2025-06-25 00:00:00',NULL,0,NULL,NULL,NULL,4,NULL,NULL,'2025-06-05','2025-06-22','2025-06-05',NULL,'2025-07-12','2025-07-02',0,0,NULL,0,NULL,1,0,NULL,1),
	(478,NULL,39,'2025-06-24 00:00:00',251,9400.00,1600.00,NULL,3,'LB37622Z0SX816064',NULL,NULL,NULL,0,'ADM001120625116',NULL,NULL,116,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FFAU3368344',0,30,0,0,NULL,1),
	(479,NULL,39,'2025-06-21 00:00:00',166,9000.00,1600.00,NULL,3,'LB37622Z6SX816330','cars/479/documents/documents_1751955078986.pdf','2025-07-08 00:00:00',NULL,0,'MOH006160525029',NULL,NULL,29,NULL,NULL,'2025-05-09','2025-05-26','2025-05-09',250.00,'2025-06-15','2025-06-05',0,0,'FFAU2255430',0,8,0,0,NULL,1),
	(480,NULL,39,'2025-06-14 00:00:00',236,113400.00,1600.00,NULL,3,'LB37622Z8SX816524','cars/480/documents/documents_1751957003307.pdf','2025-07-08 00:00:00',NULL,0,'ATE005240525074',NULL,NULL,74,NULL,NULL,'2025-05-17','2025-06-03','2025-05-17',250.00,'2025-06-23','2025-06-13',0,0,'OOLU9101067',0,9,0,0,NULL,1),
	(481,NULL,39,'2025-06-14 00:00:00',227,9000.00,1600.00,NULL,3,'LB37622Z3SX815006',NULL,NULL,NULL,0,'ATE005200525070',NULL,NULL,70,NULL,NULL,'2025-05-13','2025-05-30','2025-05-13',250.00,'2025-06-19','2025-06-09',0,0,'MRSU5017091',0,6,0,0,NULL,1),
	(482,NULL,39,'2025-07-20 00:00:00',228,9000.00,1600.00,NULL,2,'LB37622Z9SX816483','cars/482/documents/documents_1751956945489.pdf','2025-07-08 00:00:00',NULL,0,'ATE005200725189',NULL,NULL,189,NULL,NULL,'2025-05-13','2025-05-30','2025-05-13',250.00,'2025-06-19','2025-06-09',0,0,'OOLU9101067',0,9,0,0,'2025-07-20 15:49:24',1),
	(483,NULL,39,'2025-06-14 00:00:00',229,9000.00,1600.00,NULL,3,'LB37622Z6SX816523','cars/483/documents/documents_1751956860466.pdf','2025-07-08 00:00:00',NULL,0,'ATE005200525068',NULL,NULL,68,NULL,NULL,'2025-05-13','2025-05-30','2025-05-13',250.00,'2025-06-19','2025-06-09',0,0,'OOLU9101067',0,9,0,0,NULL,1),
	(484,NULL,39,'2025-06-14 00:00:00',232,9000.00,1600.00,NULL,3,'LB37622Z5SX816562','cars/484/documents/documents_1751954361522.pdf',NULL,NULL,0,'ATE005190525067',NULL,NULL,67,NULL,NULL,'2025-05-12','2025-05-29','2025-05-12',250.00,'2025-06-18','2025-06-08',0,0,'OOLU9101067',0,9,0,0,NULL,1),
	(485,NULL,39,'2025-06-14 00:00:00',235,9800.00,1600.00,NULL,3,'LB37622Z4SX819386','cars/485/documents/documents_1751360742812.pdf','2025-07-01 00:00:00',NULL,0,'ATE005060525065',NULL,NULL,65,NULL,NULL,'2025-04-29','2025-05-16','2025-04-29',250.00,'2025-06-05','2025-05-26',0,0,'COSU6422363972',0,NULL,0,0,NULL,1),
	(486,NULL,39,'2025-06-21 00:00:00',163,9000.00,1600.00,NULL,3,'LB37622Z5SX816531','cars/486/documents/documents_1751956553586.pdf','2025-07-08 00:00:00',NULL,0,'MOH006160525029',NULL,NULL,29,NULL,NULL,'2025-05-09','2025-05-26','2025-05-09',250.00,'2025-06-15','2025-06-05',0,0,'FFAU2255430',0,8,0,0,NULL,1),
	(487,NULL,39,'2025-06-13 00:00:00',105,9900.00,1600.00,NULL,3,'LB37622Z3SX816561','cars/487/documents/documents_1751956618209.pdf','2025-07-08 00:00:00',NULL,0,'MOH006160525029',NULL,NULL,29,NULL,NULL,'2025-05-09','2025-05-26','2025-05-09',250.00,'2025-06-15','2025-06-05',0,0,'FFAU2255430',0,8,0,0,NULL,1),
	(488,'SHIPPING NOT PAID',41,'2025-06-13 00:00:00',177,13080.00,NULL,NULL,2,'LS4ASE2E7PD064450',NULL,'2025-06-25 00:00:00',NULL,0,'MOH006120525038',NULL,NULL,38,NULL,NULL,'2025-05-05','2025-05-22','2025-05-05',246.00,'2025-06-11','2025-06-01',1,0,'COSU6420944208',0,NULL,0,0,NULL,4),
	(489,NULL,42,'2025-06-30 00:00:00',282,14200.00,1600.00,NULL,3,'LB37622Z4SX617308',NULL,'2025-06-18 00:00:00',NULL,0,'ADM001260625121',NULL,NULL,121,NULL,NULL,'2025-05-09','2025-05-26','2025-05-09',250.00,'2025-06-15','2025-06-05',0,0,'MRSU4218260',0,NULL,0,0,NULL,4),
	(490,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(491,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(492,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(493,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(494,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(495,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(496,'',42,NULL,NULL,13200.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',NULL,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(497,NULL,42,'2025-06-13 00:00:00',178,13500.00,1500.00,NULL,3,'LB37622Z8SX617280','cars/497/documents/documents_1751361006749.pdf','2025-06-18 00:00:00',NULL,0,'MOH006120525038',NULL,NULL,38,NULL,NULL,'2025-05-05','2025-05-22','2025-05-05',250.00,'2025-06-11','2025-06-01',0,0,'MRSU4218260',0,NULL,0,0,NULL,4),
	(498,NULL,42,'2025-06-13 00:00:00',NULL,13500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,88,NULL,NULL,'2025-05-01','2025-05-18','2025-05-01',250.00,'2025-06-07','2025-05-28',0,0,NULL,0,NULL,1,0,NULL,4),
	(499,NULL,43,'2025-06-14 00:00:00',188,15864.00,NULL,NULL,2,'LVTDB21BOPD945426','cars/499/documents/documents_1750177259993.pdf','2025-06-17 00:00:00',NULL,0,'MOH006160425043',NULL,NULL,43,NULL,NULL,'2025-04-09','2025-04-26','2025-04-09',250.00,'2025-05-16','2025-05-06',1,1,'BMOU5848830',0,NULL,0,0,NULL,1),
	(500,'did not pay freight yet',44,'2025-06-25 00:00:00',281,11800.00,1700.00,NULL,2,'LDCA41V48P3016493',NULL,'2025-06-25 00:00:00',NULL,0,'ADM001250425119',NULL,NULL,119,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'TRHU4335450',0,NULL,0,0,NULL,1),
	(501,'did not pay the freight yet ',44,'2025-06-25 00:00:00',280,11800.00,1700.00,NULL,2,'LDCA41V47P3016517',NULL,'2025-06-25 00:00:00',NULL,0,'ADM001250425119',NULL,NULL,119,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'MIEU0041451',0,NULL,0,0,NULL,1),
	(502,'did not pay freight yet',44,'2025-06-25 00:00:00',279,11800.00,1700.00,NULL,2,'LDCA41V45P3013986',NULL,'2025-06-25 00:00:00',NULL,0,'ADM001250425119',NULL,NULL,119,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'MIEU0041451',0,NULL,0,0,NULL,1),
	(503,'did not pay freight',44,'2025-06-25 00:00:00',278,11800.00,1700.00,NULL,2,'LDCA41V41P3016674',NULL,'2025-06-25 00:00:00',NULL,0,'ADM001250425119',NULL,NULL,119,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'MIEU0041451',0,NULL,0,0,NULL,1),
	(504,'did not pay freight',44,'2025-06-25 00:00:00',277,11800.00,1700.00,NULL,2,'LDCA41V48P3016705',NULL,'2025-06-25 00:00:00',NULL,0,'ADM001250425119',NULL,NULL,119,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'MIEU0041451',0,NULL,0,0,NULL,1),
	(505,NULL,44,'2025-06-14 00:00:00',212,11500.00,1500.00,NULL,2,'LDCA41V46P3016718',NULL,'2025-06-17 00:00:00',NULL,0,'ATE005080425050',NULL,NULL,50,NULL,NULL,'2025-04-01','2025-04-18','2025-04-01',250.00,'2025-05-08','2025-04-28',1,0,'TRHU4335450',0,NULL,0,0,NULL,1),
	(506,NULL,45,'2025-06-28 00:00:00',143,19600.00,1600.00,NULL,3,'LFV2B2A14S5525860',NULL,'2025-07-01 00:00:00',NULL,0,'ISH003190525012',NULL,NULL,12,'Xp2357',NULL,'2025-05-16','2025-06-02','2025-05-16',250.00,'2025-06-22','2025-06-12',0,0,'CAIU4964250',0,NULL,0,0,NULL,2),
	(507,'akk\n',45,'2025-06-14 00:00:00',218,19200.00,1600.00,NULL,4,'LFV2B2A17S5525884',NULL,'2025-07-01 00:00:00',NULL,0,'ATE005220525071',NULL,NULL,71,NULL,NULL,'2025-05-15','2025-06-01','2025-05-15',250.00,'2025-06-21','2025-06-11',0,0,'CAIU4964250',0,NULL,0,0,NULL,2),
	(508,'',46,NULL,NULL,13500.00,NULL,NULL,NULL,'LJD5AA1D8S0212379',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU7143128',0,24,0,0,NULL,2),
	(509,'',46,NULL,NULL,13500.00,NULL,NULL,NULL,'LJD5AA1D1S0212644',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU7143128',0,24,0,0,NULL,2),
	(510,'sell dinar',47,'2025-07-18 00:00:00',356,13600.00,1600.00,NULL,3,'LJD5AA1D7S0211269',NULL,NULL,NULL,0,'ADM001180725183',NULL,NULL,183,NULL,NULL,'2025-06-07',NULL,'2025-06-07',237.00,NULL,NULL,0,0,'CSNU8581025',0,25,0,0,'2025-07-18 11:32:00',4),
	(511,NULL,47,'2025-07-10 00:00:00',240,13200.00,1600.00,NULL,3,'LJD5AA1D7S0211272',NULL,NULL,NULL,0,'ISH003160625079','cars/511/sell_pi/sell_pi_1752825834281.pdf',NULL,79,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',250.00,'2025-06-03','2025-05-24',0,0,'CSNU8581025',0,25,0,0,NULL,4),
	(512,'must be loaded with akkon ',47,'2025-07-10 00:00:00',238,13200.00,1600.00,NULL,3,'LJD5AA1D6S0211277',NULL,NULL,NULL,0,'ISH003140625077','cars/512/sell_pi/sell_pi_1752826171515.pdf',NULL,77,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',250.00,'2025-06-03','2025-05-24',0,0,'CSNU8581025',0,25,0,0,NULL,4),
	(513,NULL,47,'2025-06-16 00:00:00',241,13200.00,1600.00,NULL,2,'LJD5AA1D5S0211271',NULL,NULL,NULL,0,'ISH003160625079',NULL,NULL,79,NULL,NULL,'2025-06-09','2025-06-26','2025-06-09',250.00,'2025-07-16','2025-07-06',0,0,'CSNU8581025',0,25,0,0,NULL,4),
	(514,NULL,47,'2025-06-15 00:00:00',239,13200.00,1600.00,NULL,2,'LJD5AA1D8S0211278',NULL,NULL,NULL,0,'ISH003150625078',NULL,NULL,78,'',NULL,'2025-06-08',NULL,'2025-06-08',250.00,NULL,'2025-07-05',0,0,'OOCU7143128',0,24,0,0,NULL,4),
	(515,NULL,48,'2025-07-17 00:00:00',355,11000.00,1600.00,NULL,3,'LFV2B2C1XP3700192',NULL,NULL,NULL,0,'ATE005210625180',NULL,NULL,180,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'OOCU6816693',0,26,0,0,NULL,1),
	(516,NULL,48,'2025-07-17 00:00:00',354,11000.00,1600.00,NULL,3,'LFV2B2C16P3700870',NULL,NULL,NULL,0,'ATE005170725178',NULL,NULL,178,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'OOCU6816693',0,26,0,0,NULL,1),
	(517,NULL,48,'2025-07-16 00:00:00',352,10800.00,1600.00,NULL,3,'LFV2B2C11P3701974',NULL,NULL,NULL,0,'ISH003160725175',NULL,NULL,175,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'OOCU6816693',0,26,0,0,'2025-07-16 07:17:03',1),
	(518,NULL,48,'2025-07-17 00:00:00',353,11000.00,1600.00,NULL,3,'LFV2B2C10N3773987',NULL,NULL,NULL,0,'MOH006170725177',NULL,NULL,177,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'CSNU7094969',0,28,0,0,'2025-07-17 05:19:12',1),
	(519,NULL,48,'2025-07-04 00:00:00',300,11120.00,1600.00,NULL,3,'LFV2B2C10P3700184',NULL,NULL,NULL,0,'ISH003030725139',NULL,NULL,139,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'CSNU7094969',0,28,0,0,NULL,1),
	(520,NULL,48,'2025-06-24 00:00:00',272,11280.00,1600.00,NULL,3,'LFV2B2C14P3700544',NULL,NULL,NULL,0,'ISH003140625051',NULL,NULL,51,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'CSNU7094969',0,28,0,0,NULL,1),
	(521,'',33,NULL,NULL,11000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(522,'',48,NULL,NULL,11000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(523,'',48,NULL,NULL,11000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(524,'',48,NULL,NULL,11000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,92,NULL,NULL,'2025-04-27','2025-05-14','2025-04-27',NULL,'2025-06-03','2025-05-24',1,0,NULL,0,NULL,1,0,NULL,1),
	(525,NULL,49,'2025-06-25 00:00:00',237,15800.00,1600.00,NULL,3,'LSVYSAC1XSN056960',NULL,'2025-06-25 00:00:00',NULL,0,'ATE005230525076',NULL,NULL,76,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TRHU8440448',0,NULL,0,0,NULL,4),
	(529,'',50,NULL,NULL,18500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,95,NULL,NULL,'2025-02-28','2025-03-17','2025-02-28',NULL,'2025-04-06','2025-03-27',0,0,NULL,0,NULL,1,0,NULL,1),
	(530,'',50,NULL,NULL,18500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,100,NULL,NULL,'2025-03-20','2025-04-06','2025-03-20',NULL,'2025-04-26','2025-04-16',0,0,NULL,0,NULL,1,0,NULL,1),
	(531,'',50,NULL,NULL,18500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,100,NULL,NULL,'2025-03-20','2025-04-06','2025-03-20',NULL,'2025-04-26','2025-04-16',0,0,NULL,0,NULL,1,0,NULL,1),
	(532,NULL,52,'2025-06-22 00:00:00',250,8800.00,1600.00,NULL,3,'LLV2C3B23S0000318',NULL,'2025-06-25 00:00:00',NULL,0,'TOU004220625106',NULL,NULL,106,NULL,NULL,'2025-06-15','2025-07-02','2025-06-15',250.00,'2025-07-22','2025-07-12',0,0,'COSU6415809610',0,NULL,0,0,NULL,1),
	(533,NULL,53,'2025-06-24 00:00:00',255,20880.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ADM001230425115',NULL,NULL,115,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,2),
	(534,NULL,54,'2025-06-23 00:00:00',254,24580.00,1500.00,NULL,3,'LFV2B28Y3R6550515',NULL,NULL,NULL,0,'ADM001230425115',NULL,NULL,115,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(535,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(536,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(537,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(538,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(539,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(540,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(541,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(542,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(543,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(544,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(545,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(546,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(547,'READY FOR LOADING CAR',55,'2025-07-16 00:00:00',349,7600.00,1600.00,1,3,'LLV2C3B27S0005456',NULL,NULL,NULL,0,'MOH006160725172',NULL,NULL,172,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FCIU7377850',0,23,0,0,'2025-07-16 06:00:51',1),
	(548,'READY FOR LOADING CAR',55,'2025-07-16 00:00:00',348,7600.00,1600.00,1,3,'LLV2C3B25S0005455',NULL,NULL,NULL,0,'MOH006160725171',NULL,NULL,171,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FCIU7377850',0,23,0,0,'2025-07-16 05:57:42',1),
	(549,'SAME CONTAINER WITH BOUDJEMMA RACIM',55,'2025-07-16 00:00:00',337,7600.00,1600.00,1,3,'LLV2C3B28S0004400',NULL,NULL,NULL,0,'MOH006160725165',NULL,NULL,165,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'FCIU7377850',0,23,0,0,'2025-07-16 05:53:42',1),
	(550,'SAME CONTAINER WITH BOUDJEMAA RACIM',23,'2025-07-16 00:00:00',347,7600.00,1600.00,1,3,'LLV2C3B25S0005259',NULL,NULL,NULL,0,'MOH006160725170',NULL,NULL,170,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TGHU9855344',0,22,0,0,'2025-07-16 05:52:14',1),
	(551,'SAME CONTAINER WITH BOUDJEMAA RACIM',23,'2025-07-16 00:00:00',346,7600.00,1600.00,1,3,'LLV2C3B29S0005300',NULL,NULL,NULL,0,'MOH006160725169',NULL,NULL,169,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TGHU9855344',0,22,0,0,'2025-07-16 05:49:25',1),
	(552,'THIS ONE CAR IS THE ON THE PORT AND IS READY TO BE LOADED.',22,'2025-07-16 00:00:00',339,7600.00,1600.00,1,3,'LLV2C3B23S0005308',NULL,NULL,NULL,0,'MOH006160725168',NULL,NULL,168,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TGHU9855344',0,22,0,0,'2025-07-16 05:46:00',1),
	(553,NULL,55,'2025-07-15 00:00:00',13,8400.00,1600.00,1,2,'LLV2C3B20S0000907',NULL,'2025-05-13 00:00:00',NULL,0,'TOU004090525002',NULL,NULL,2,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'CSNU8512132',0,NULL,0,0,NULL,1),
	(554,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(555,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(556,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(557,'',55,NULL,NULL,8400.00,NULL,1,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(558,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(559,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(560,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(561,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,7,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(562,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(563,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(564,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(565,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(566,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(567,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(568,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(569,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(570,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(571,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(572,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(573,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(574,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(575,'',55,NULL,NULL,8400.00,NULL,2,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,3,'2025-07-08',NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(576,NULL,55,'2025-07-19 00:00:00',361,8948.52,1600.00,2,3,'LLV2C3B29S0001361',NULL,NULL,NULL,0,'ADM001190725185',NULL,NULL,185,NULL,3,'2025-07-08',NULL,NULL,237.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-19 14:21:15',1),
	(577,NULL,55,'2025-07-19 00:00:00',336,8948.52,1600.00,2,3,'LLV2C3B29S0001120',NULL,NULL,NULL,0,'ADM001190725185',NULL,NULL,185,NULL,3,'2025-07-08',NULL,NULL,237.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-19 14:20:32',1),
	(578,NULL,55,'2025-07-18 00:00:00',358,8400.00,1600.00,NULL,3,'LLV2C3B23S0000481',NULL,NULL,NULL,0,'ISH003140625051',NULL,NULL,51,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-18 13:44:20',1),
	(579,NULL,55,'2025-07-18 00:00:00',285,8948.52,1600.00,NULL,3,'LLV2C3B27S0003805',NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,237.00,NULL,NULL,0,0,'FCIU7377850',0,23,0,0,'2025-07-18 14:03:14',1),
	(580,NULL,55,'2025-07-16 00:00:00',320,8800.00,1600.00,NULL,3,'LLV2C3B26S0005318',NULL,NULL,NULL,0,'ISH003140725155',NULL,NULL,155,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'MSDU7188538',0,21,0,0,'2025-07-16 10:54:18',1),
	(581,NULL,40,'2025-07-16 00:00:00',338,8800.00,1600.00,NULL,3,'LLV2C3B2XS0005256',NULL,NULL,NULL,0,'ISH003160725166',NULL,NULL,166,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TGHU9855344',0,22,0,0,'2025-07-16 05:21:24',1),
	(582,'FREIGHT NOT PAID YET',55,'2025-07-02 00:00:00',294,8400.00,1600.00,NULL,2,'LLV2C3B28R0000423',NULL,'2025-04-08 00:00:00',NULL,0,'ISH003010725134',NULL,NULL,134,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'CAAU8051831',0,NULL,0,0,NULL,1),
	(583,NULL,55,'2025-07-01 00:00:00',293,8600.00,1600.00,NULL,2,'LLV2C3B25S0001406','cars/583/documents/documents_1752789638561.pdf',NULL,NULL,0,'ISH003010725133','cars/583/sell_pi/sell_pi_1752826103775.pdf',NULL,133,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TCLU5400775',0,7,0,0,NULL,1),
	(584,NULL,55,'2025-06-30 00:00:00',288,8400.00,1600.00,NULL,3,'LLV2C3B28S0001433','cars/584/documents/documents_1751370985985.pdf','2025-06-09 00:00:00',NULL,0,'ISH003120425128','cars/584/sell_pi/sell_pi_1752826204313.pdf',NULL,128,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TCNU5535823',0,NULL,0,0,NULL,1),
	(585,NULL,55,'2025-06-30 00:00:00',287,8400.00,1600.00,NULL,3,'LLV2C3B28S0001495','cars/585/documents/documents_1751370731097.pdf','2025-06-09 00:00:00',NULL,0,'ISH003120425128','cars/585/sell_pi/sell_pi_1752826234217.pdf',NULL,128,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TCNU5535823',0,NULL,0,0,NULL,1),
	(586,NULL,55,'2025-06-30 00:00:00',286,8200.00,1600.00,NULL,3,'LLV2C3B24S0001350','cars/586/documents/documents_1751370638884.pdf','2025-06-09 00:00:00',NULL,0,'ISH003150425127',NULL,NULL,127,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TCNU5535823',0,NULL,0,0,NULL,1),
	(587,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B20S0003709',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(588,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B23S0005129',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(589,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B26S0004301',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(590,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B22S0005476',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(591,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B27S0004775',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(592,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B27S0005117',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(593,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B25S0003897',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(594,'',55,NULL,NULL,8400.00,NULL,NULL,NULL,'LLV2C3B2XS0004883',NULL,NULL,NULL,0,NULL,NULL,NULL,4,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,1,0,NULL,1),
	(595,NULL,56,'2025-06-29 00:00:00',65,9350.00,1600.00,NULL,3,'LVVDB21B6PE040285',NULL,NULL,NULL,0,'ADM001290525125',NULL,NULL,125,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,'CAIU7094512',0,13,0,0,NULL,1),
	(596,'can not sell this',57,NULL,NULL,19000.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(597,'',58,'2025-07-10 00:00:00',72,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001270325100',NULL,NULL,100,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(598,'FREIGHT IS NOT PAID YET',58,'2025-07-01 00:00:00',292,19000.00,NULL,NULL,2,'LFV2B2A13S5506443',NULL,'2025-04-08 00:00:00',NULL,0,'ISH003250325131',NULL,NULL,131,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'CAAU8051831',0,NULL,0,0,NULL,1),
	(599,'',58,'2025-07-10 00:00:00',72,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'ADM001270325100',NULL,NULL,100,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(600,'',59,NULL,NULL,11500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(601,NULL,60,'2025-07-20 00:00:00',366,20340.93,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ADM001200725188',NULL,NULL,188,NULL,NULL,NULL,NULL,NULL,237.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-20 11:59:00',2),
	(602,NULL,60,'2025-07-20 00:00:00',298,18029.63,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ADM001140125137',NULL,NULL,137,NULL,NULL,NULL,NULL,NULL,270.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-20 11:38:58',2),
	(603,NULL,61,'2025-07-15 00:00:00',332,7600.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(604,NULL,61,'2025-07-15 00:00:00',331,7600.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(605,NULL,61,'2025-07-15 00:00:00',330,7600.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(606,NULL,61,'2025-07-15 00:00:00',328,7600.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(607,NULL,62,'2025-07-15 00:00:00',329,10200.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003300625126',NULL,NULL,126,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,1),
	(608,'we only doing freight',63,NULL,NULL,0.00,NULL,NULL,NULL,'LJD4AA192S0137198',NULL,'2025-07-08 00:00:00',NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'SEKU4542148',0,12,0,0,NULL,1),
	(609,NULL,64,'2025-07-15 00:00:00',335,7600.00,1600.00,NULL,3,'LLV2C3A25S0001214',NULL,NULL,NULL,0,'SOF002140725157',NULL,NULL,157,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,'TCLU8624608',0,14,0,0,'2025-07-15 15:29:36',1),
	(610,NULL,64,'2025-07-14 00:00:00',325,8187.23,1600.00,NULL,3,'LLV2C3A23S0001194',NULL,NULL,NULL,0,'SOF002140725157',NULL,NULL,157,NULL,NULL,NULL,NULL,NULL,235.00,NULL,NULL,0,0,'TCLU8624608',0,14,0,0,NULL,1),
	(611,NULL,64,'2025-07-14 00:00:00',326,8187.23,1600.00,NULL,3,'LLV2C3A20S0001234',NULL,NULL,NULL,0,'SOF002140725157',NULL,NULL,157,NULL,NULL,NULL,NULL,NULL,235.00,NULL,NULL,0,0,'TCLU8624608',0,14,0,0,NULL,1),
	(612,NULL,64,'2025-07-19 00:00:00',362,7893.67,1600.00,NULL,3,'LLV2C3A23S0001230',NULL,NULL,NULL,0,'ADM001190725185',NULL,NULL,185,NULL,NULL,NULL,NULL,NULL,237.00,NULL,NULL,0,0,'TCLU8624608',0,14,0,0,'2025-07-19 14:22:32',1),
	(613,'',65,NULL,NULL,13100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,5),
	(614,'',65,NULL,NULL,13100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,5),
	(615,'',65,NULL,NULL,13100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,5),
	(616,'',65,NULL,NULL,13100.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,5),
	(619,NULL,65,'2025-07-18 00:00:00',221,14011.81,1600.00,NULL,3,'LB37622ZXSX621556',NULL,NULL,NULL,0,'ISH003140625051',NULL,NULL,51,NULL,NULL,NULL,NULL,NULL,237.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-18 13:16:04',5),
	(620,'just for loading',66,NULL,NULL,0.00,NULL,NULL,NULL,'LJD5AAID9S0212116',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'SEKU4542148',0,12,0,0,NULL,1),
	(622,'',68,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(623,'',68,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(624,NULL,68,'2025-07-20 00:00:00',333,13800.00,1600.00,NULL,3,NULL,NULL,NULL,NULL,0,'ISH003150725161',NULL,NULL,161,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-20 06:01:35',4),
	(625,NULL,68,'2025-07-19 00:00:00',363,13550.00,1600.00,NULL,3,'LFV2B20L4S4031202',NULL,NULL,NULL,0,'ADM001190725185',NULL,NULL,185,NULL,NULL,NULL,NULL,NULL,237.00,NULL,NULL,0,0,NULL,0,NULL,0,0,'2025-07-19 14:52:20',4),
	(626,'',69,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,2),
	(627,'',69,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,2),
	(628,'',69,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,2),
	(629,'',69,NULL,NULL,14400.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,2),
	(630,'',70,NULL,NULL,7600.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,1,1,NULL,0,NULL,0,0,NULL,1),
	(632,NULL,73,'2025-07-17 00:00:00',12,10820.00,1500.00,NULL,2,NULL,NULL,NULL,NULL,0,'ATE005150425059',NULL,NULL,59,NULL,NULL,NULL,NULL,NULL,250.00,NULL,NULL,1,0,NULL,0,NULL,0,0,'2025-07-17 19:21:01',1),
	(633,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC13SN068304',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOLU9829392',0,32,0,0,NULL,4),
	(634,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC19SN069098',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOLU9829392',0,32,0,0,NULL,4),
	(635,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC19SN068825',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOLU9829392',0,32,0,0,NULL,4),
	(636,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC1XSN068994',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOLU9829392',0,32,0,0,NULL,4),
	(637,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC19SN068968',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU6823270',0,33,0,0,NULL,4),
	(638,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC10SN069118',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU6823270',0,33,0,0,NULL,4),
	(639,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC19SN069036',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU6823270',0,33,0,0,NULL,4),
	(640,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,'LSVYSAC12SN069492',NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,'OOCU6823270',0,33,0,0,NULL,4),
	(641,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(642,NULL,74,NULL,NULL,15500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(643,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(644,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(645,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(646,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(647,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(648,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(649,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(650,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(651,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(652,NULL,75,NULL,NULL,9500.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,4),
	(653,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(654,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(655,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(656,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(657,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(658,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(659,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(660,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(661,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(662,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(663,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(664,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(665,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(666,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(667,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(668,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(669,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(670,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(671,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3),
	(672,NULL,77,NULL,NULL,19300.00,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,0,NULL,0,NULL,0,0,NULL,3);

/*!40000 ALTER TABLE `cars_stock` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table chat_groups
# ------------------------------------------------------------

CREATE TABLE `chat_groups` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `id_user_owner` int(11) DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `chat_groups` WRITE;
/*!40000 ALTER TABLE `chat_groups` DISABLE KEYS */;

INSERT INTO `chat_groups` (`id`, `name`, `description`, `id_user_owner`, `is_active`)
VALUES
	(1,'sofia',NULL,1,1),
	(2,'[Car #534 - A3] check if it is loaded','Chat group for task: [Car #534 - A3] check if it is loaded',1,1),
	(3,'[Supplier #2 - MUMU Henan rongwei international trade co.,ltd] pay mumu balance 103600 usd','Chat group for task: [Supplier #2 - MUMU Henan rongwei international trade co.,ltd] pay mumu balance 103600 usd',2,1),
	(4,'[Sell Bill #4 - MOH006120625004] did walid finished payment of this order','Chat group for task: [Sell Bill #4 - MOH006120625004] did walid finished payment of this order',1,1),
	(5,'[Client #208 - HAMDACHE ACHOUR] 2 clients with same id number','Chat group for task: [Client #208 - HAMDACHE ACHOUR] 2 clients with same id number',1,1),
	(6,'[Client #281 - CHEGGA OUAIL] this guy took the peugeot 2008 no 5','Chat group for task: [Client #281 - CHEGGA OUAIL] this guy took the peugeot 2008 no 5',1,1),
	(7,'orders of kamel midat are not all filled in.','Chat group for task: orders of kamel midat are not all filled in.',3,1),
	(8,'[Sell Bill #124 - ADM001290525124] walid did not pay any of this order','Chat group for task: [Sell Bill #124 - ADM001290525124] walid did not pay any of this order',3,1),
	(9,'[Sell Bill #135 - ADM001200325135] send docs to fellowing address','Chat group for task: [Sell Bill #135 - ADM001200325135] send docs to fellowing address',1,1),
	(10,'[Car #489 - COOLRAY SPORT] Be sure with Greenline issue car to this buyer','Chat group for task: [Car #489 - COOLRAY SPORT] Be sure with Greenline issue car to this buyer',1,1),
	(11,'[Car #525 - THARU TOP] this guy take the car under ishaq name','Chat group for task: [Car #525 - THARU TOP] this guy take the car under ishaq name',2,1),
	(12,'[Sell Bill #144 - MOH006100725144] THISN GUY BOUGHT LOADED CAR','Chat group for task: [Sell Bill #144 - MOH006100725144] THISN GUY BOUGHT LOADED CAR',2,1),
	(13,'[Sell Bill #144 - MOH006100725144] THISN GUY BOUGHT LOADED CAR','Chat group for task: [Sell Bill #144 - MOH006100725144] THISN GUY BOUGHT LOADED CAR',2,1),
	(14,'ISH003010725133 want to change destination','Chat group for task: ISH003010725133 want to change destination',1,1),
	(15,'kerouicha orders not here','Chat group for task: kerouicha orders not here',1,1);

/*!40000 ALTER TABLE `chat_groups` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table chat_last_read_message
# ------------------------------------------------------------

CREATE TABLE `chat_last_read_message` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_group` int(11) DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `id_last_read_message` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_group` (`id_group`,`id_user`)
) ENGINE=InnoDB AUTO_INCREMENT=120 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `chat_last_read_message` WRITE;
/*!40000 ALTER TABLE `chat_last_read_message` DISABLE KEYS */;

INSERT INTO `chat_last_read_message` (`id`, `id_group`, `id_user`, `id_last_read_message`)
VALUES
	(1,1,2,7),
	(3,1,1,7),
	(13,2,1,6),
	(25,2,2,6),
	(30,3,2,8),
	(31,3,1,11),
	(37,6,1,40),
	(38,7,3,27),
	(42,8,1,22),
	(43,4,1,23),
	(44,7,1,28),
	(46,8,3,24),
	(47,4,3,25),
	(56,10,1,45),
	(58,6,2,31),
	(60,10,2,30),
	(61,11,2,34),
	(64,11,1,43),
	(92,10,3,45),
	(99,11,9,43),
	(106,12,2,50),
	(109,12,6,50),
	(115,12,9,51);

/*!40000 ALTER TABLE `chat_last_read_message` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table chat_messages
# ------------------------------------------------------------

CREATE TABLE `chat_messages` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_group` int(11) DEFAULT NULL,
  `message_from_user_id` int(11) DEFAULT NULL,
  `chat_replay_to_message_id` int(11) DEFAULT NULL,
  `message` text DEFAULT NULL,
  `time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=52 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `chat_messages` WRITE;
/*!40000 ALTER TABLE `chat_messages` DISABLE KEYS */;

INSERT INTO `chat_messages` (`id`, `id_chat_group`, `message_from_user_id`, `chat_replay_to_message_id`, `message`, `time`)
VALUES
	(1,1,1,NULL,'hello','2025-06-28 14:48:35'),
	(2,1,2,NULL,'hi','2025-06-28 14:49:05'),
	(3,1,1,NULL,'I love youu','2025-06-28 14:49:18'),
	(4,1,2,NULL,'mm','2025-06-28 14:50:21'),
	(5,2,1,NULL,'Hello','2025-06-28 15:29:57'),
	(6,2,1,NULL,'Yes','2025-06-28 15:31:52'),
	(7,1,1,NULL,'whats going on?','2025-06-28 15:56:00'),
	(8,3,2,NULL,'livan auto 	60*7400=444000\n		paid 	133200\n		paid 	207200\n		need to pay 	103600','2025-06-29 11:50:11'),
	(9,3,1,NULL,'thanks my love','2025-06-29 14:35:45'),
	(10,3,1,NULL,'I get it','2025-06-29 14:36:06'),
	(11,3,1,NULL,'[FILE]chat_files/group_3/1751232635727_CamScanner 20-05-2025 19.46.pdf|CamScanner 20-05-2025 19.46.pdf|2985993|application/pdf',NULL),
	(12,3,1,NULL,'this is just a test ðŸ˜œ','2025-06-29 21:31:52'),
	(13,3,1,NULL,'[VOICE]chat_files/group_3/1751235199975_voice_message.wav|voice_message.wav|53291|audio/wav|4',NULL),
	(14,5,1,NULL,'I did not find where is this problem','2025-06-30 09:41:19'),
	(15,6,1,NULL,'hello','2025-06-30 11:18:31'),
	(16,6,1,NULL,'this one','2025-06-30 11:18:37'),
	(17,7,3,NULL,'problem with Cars are not in sotck','2025-06-30 17:29:03'),
	(18,7,3,NULL,'still need 4 MG5 White Color','2025-06-30 17:29:28'),
	(19,7,3,NULL,'1MG Luxury with Sunroof','2025-06-30 17:29:39'),
	(20,7,3,NULL,'i will creat the client let me khow when stock is OK','2025-06-30 17:30:12'),
	(21,8,3,NULL,'Please confirm with sofia if she got 10000 USD From Walid let me calculat what he still need to pay','2025-06-30 17:34:03'),
	(22,8,3,NULL,'for Kia Sonet Invoice he sad he did  transfer i asked about Switf','2025-06-30 17:34:20'),
	(23,4,3,NULL,'send me PDF invoice to my email i cant found it here','2025-06-30 17:34:36'),
	(24,8,1,NULL,'[VOICE]chat_files/group_8/1751322621192_voice_message.wav|voice_message.wav|290927|audio/wav|19',NULL),
	(25,4,1,NULL,'I did not make it.','2025-06-30 22:32:12'),
	(26,7,1,NULL,'[VOICE]chat_files/group_7/1751322783919_voice_message.wav|voice_message.wav|243593|audio/wav|16',NULL),
	(27,7,1,NULL,'ðŸ˜¥','2025-06-30 22:33:23'),
	(28,7,3,NULL,'Ø§Ù„Ù„Ù‡ ÙŠØ¹ÙŠÙ†','2025-07-01 17:35:27'),
	(29,10,1,NULL,'Ø§Ù„Ø³Ù„Ø§Ù… Ø¹Ù„ÙŠÙƒÙ… ÙˆØ±Ø­Ù…Ø© Ø§Ù„Ù„Ù‡ ÙˆØ¨Ø±ÙƒØ§ØªÙ‡','2025-07-08 08:25:54'),
	(30,10,1,NULL,'Ù‡Ø°Ø§ Ø§Ù„Ø²Ø¨ÙˆÙ† Ø±Ø§Ù†ÙŠ Ø§Ø¹Ø·ÙŠØªÙ‡ ÙƒÙˆÙ„Ø±Ø§ÙŠ Ø³Ø¨ÙˆØ± Ø±Ù…Ø§Ø¯ÙŠØ©','2025-07-08 08:26:32'),
	(31,6,2,NULL,'you gave no.5 2008 to client 123 last time, please double confirm','2025-07-08 11:35:15'),
	(32,11,2,NULL,'this vin number is not wrong','2025-07-08 11:44:04'),
	(33,11,2,NULL,'the take tharu top , and i gave him vin.','2025-07-08 11:44:18'),
	(34,11,2,NULL,'only us give vin. no one else','2025-07-08 11:44:27'),
	(35,11,2,NULL,'LSVYSAC1XSN056960- under ishak , for sell','2025-07-08 11:51:12'),
	(36,11,2,NULL,'LSVYSAC12SN056967 for sell','2025-07-08 11:51:19'),
	(37,11,1,NULL,'ok','2025-07-08 14:44:02'),
	(38,6,1,NULL,'[FILE]chat_files/group_6/1751986108750_â€Ù„Ù‚Ø·Ø© Ø´Ø§Ø´Ø© 2025-07-08 ÙÙŠ 22.47.56.jpeg|â€Ù„Ù‚Ø·Ø© Ø´Ø§Ø´Ø© 2025-07-08 ÙÙŠ 22.47.56.jpeg|1702526|image/jpeg',NULL),
	(39,6,1,NULL,'islam took 4 unites of 2008\nand 1 more unit he booked and payed deposit. \nI am talking about this extra 1','2025-07-08 14:52:29'),
	(40,6,1,NULL,'client 123 I know he took 1 unit of 2008.','2025-07-08 14:52:59'),
	(41,6,1,NULL,'it is still for him.','2025-07-08 14:53:21'),
	(42,10,3,NULL,'send me his passport or give me nick name please','2025-07-08 15:43:15'),
	(43,11,9,NULL,'ok','2025-07-09 03:20:03'),
	(44,10,3,NULL,'done waiting for BL from greenling team','2025-07-09 11:04:29'),
	(45,10,1,NULL,'[FILE]chat_files/group_10/1752063762022_Screenshot 2025-07-09 at 8.22.09â€¯PM.png|Screenshot 2025-07-09 at 8.22.09â€¯PM.png|66056|image/png',NULL),
	(46,12,2,NULL,'it is done','2025-07-10 08:11:55'),
	(47,12,2,NULL,'with greenline also. hope they change it','2025-07-10 08:12:06'),
	(48,12,2,NULL,'ignore up message. it was a misake','2025-07-10 08:12:40'),
	(49,12,2,NULL,'if this bought a car on sea, he should not connect to the car with vin.','2025-07-10 08:13:18'),
	(50,12,2,NULL,'please do it again','2025-07-10 08:29:17'),
	(51,12,6,NULL,'boss is on the phone right now , can you tell me how to do it ?','2025-07-10 08:39:13');

/*!40000 ALTER TABLE `chat_messages` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table chat_read_by
# ------------------------------------------------------------

CREATE TABLE `chat_read_by` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_message` int(11) DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `read_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;



# Dump of table chat_users
# ------------------------------------------------------------

CREATE TABLE `chat_users` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_user` int(11) DEFAULT NULL,
  `id_chat_group` int(11) DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT 1,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `chat_users` WRITE;
/*!40000 ALTER TABLE `chat_users` DISABLE KEYS */;

INSERT INTO `chat_users` (`id`, `id_user`, `id_chat_group`, `is_active`)
VALUES
	(1,1,1,1),
	(2,2,1,1),
	(3,2,2,1),
	(4,1,2,1),
	(5,1,3,1),
	(6,2,3,1),
	(7,3,4,1),
	(8,1,4,1),
	(9,2,5,1),
	(10,1,5,1),
	(11,2,6,1),
	(12,1,6,1),
	(13,1,7,1),
	(14,3,7,1),
	(15,1,8,1),
	(16,3,8,1),
	(17,3,9,1),
	(18,1,9,1),
	(19,2,10,1),
	(20,3,10,1),
	(21,1,10,1),
	(22,1,11,1),
	(23,9,11,1),
	(24,2,11,1),
	(25,6,12,1),
	(26,9,12,1),
	(27,2,12,1),
	(28,6,13,1),
	(29,9,13,1),
	(30,2,13,1),
	(31,3,14,1),
	(32,1,14,1),
	(33,1,15,1);

/*!40000 ALTER TABLE `chat_users` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table clients
# ------------------------------------------------------------

CREATE TABLE `clients` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobiles` varchar(255) DEFAULT 'please provide mobile',
  `id_copy_path` varchar(255) DEFAULT NULL,
  `id_no` varchar(255) DEFAULT NULL,
  `is_broker` tinyint(1) DEFAULT 0,
  `is_client` tinyint(1) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `client_name_unic` (`name`),
  UNIQUE KEY `id_no` (`id_no`)
) ENGINE=InnoDB AUTO_INCREMENT=383 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `clients` WRITE;
/*!40000 ALTER TABLE `clients` DISABLE KEYS */;

INSERT INTO `clients` (`id`, `name`, `address`, `email`, `mobiles`, `id_copy_path`, `id_no`, `is_broker`, `is_client`, `notes`)
VALUES
	(4,'Ali Touhami','mostaganem','touhali@merhab.com','','','',1,0,NULL),
	(5,'RADJAH ABDERAHMANE','HASSIMAMECHE MOSTAGANEM','e@me.com','no mobile','ids/5.jpg','156583684',0,1,'please add mobile'),
	(6,'FARAOUN ABDESSAMAD','SIDI BEL ABBES','a@me.com','no mobile','ids/6.jpg','186897195',0,1,'please add mobile'),
	(7,'KORIDJIDJ MOHAMMED','SIDI M\'HAMED BENALI MOSTAGANEM','','no mobile','ids/7.jpg','166449718',0,1,'no mobile'),
	(8,'BOUALI-YOUCEF AHMED','MAZOUNA GHELIZANE','','no mobile','ids/8.jpg','309431053',0,1,'no mobile '),
	(9,'BENKERDAGH OMAR','MOSTAGANEM','','no moile','ids/9.jpg','312539781',0,1,'no mobile'),
	(10,'TRIQUI BOCHRA','MOSTAGANEM','','no mobile','ids/10.jpg','186320728',0,1,'no mobile'),
	(11,'ABDESSETTAR ABDELKARIM','MOSTAGANEM','','no mobile','ids/11.jpg','177477130',0,1,'no mobile'),
	(12,'SAHABA MOKHTAR','MEDJANA B.B.ARRERIDJ','','no mobile','ids/12.jpg','305164370',0,1,'no mobile'),
	(13,'BENABDALLAH KARIM','MOSTAGANEM','','no mobile','ids/13.jpg','156262895',0,1,'no mobile'),
	(14,'ABDELHADI MOHAMED','SAIDA','','no mobile','ids/14.jpg','303284775',0,1,'no mobile'),
	(15,'ABDERREZZAG MOHAMED','MOSTAGANEM','','no mobile','ids/15.jpg','107453755',0,1,'no mobile'),
	(16,'CHAREF-BENDAHA NABIL','MOSTAGANEM','','no mobile','ids/16.jpg','169642056',0,1,'no mobile'),
	(17,'BENESSALAH AHMED','MOSTAGANEM','','no mobile','ids/17.jpg','313568087',0,1,'no mobile'),
	(18,'BELLAREDJ NABIL','REMCHI','','no mobile','ids/18.jpg','313689586',0,1,'no mobile'),
	(19,'ZERROUKI HAFID','KHEIREDDINE MOSTAGANEM','','no mobile','ids/19.jpg','305510234',0,1,'no mobile'),
	(20,'BERRAHOU FARID','BETHIOUA ORAN','','no mobile','ids/20.jpg','313550779',0,1,'no mobile'),
	(21,'BENANTEUR ABDELKRIM','MOSTAGANEM','','no mobile','ids/21.jpg','312886392',0,1,'no mobile'),
	(22,'BOUKROUK LOTFI','JIJEL','','no mobile','ids/22.jpg','169240705',0,1,'no mobile'),
	(23,'BRIHI YACINE','Algeria,Algers,Kouba,Diar el Bahia Kouba Algers Payed By transfer Euro','','+213 558 01 70 14','ids/23.jpg','166721079',0,1,''),
	(24,'CHOUBANE Samih','Algeria,Bouira,Saharidj Centre,Aggoune Hocine  House : nÂ° 11','','no phone ','ids/24.jpg','186228837',0,1,''),
	(25,'CHOUBANE LOUNES','Algeria,Bouira,Saharidj Centre,Aggoune Hocine  House : nÂ° 11','','NO PHONE','ids/25.jpg','309753776',0,1,''),
	(26,'TAYEB BOUDJEMA DALILA','MOSTAGANEM','','no mobile','ids/26.jpg','186752195',0,1,'no mobile'),
	(27,'BELMAHI YOUNES','Algeria,EL Bayedh,EL bayedh,EL HOURIA CITY , EL BAYADH ALGERIA','','+213 560 27 63 26','ids/27.jpg','182210859',0,1,''),
	(28,'MERZOUK ZAKARIA','Algeria,Bouira,COOP ABA BAKER SEDIK,CITE 56 LOGEMENTS ','','no phone','ids/28.jpg','306564611',0,1,''),
	(29,'MERZOUK MOHAMED-TAHAR','Algeria ,Bouira,OOP ABA BAKER SEDIK,CITE 56 LOGEMENTS','','no phone','ids/29.jpg','305001046',0,1,''),
	(30,'ACHIT MANAL',' AlgÃ©ria,Bouira,CITE GRAND BOULEVARD RÃ‰SIDENCE EL YASMINE BLOC B','','no phone','ids/30.jpg','169730598',0,1,''),
	(32,'MOHAMMED OUSSAID TAREK','AlgÃ©ria,Bouira , LAKHDARIA,CITE 56 LOGTS,House: num 8','','no phone','ids/32.jpg','305874722',0,1,''),
	(33,'KESSOURI REDOUANE','AlgÃ©ria,Bouira ,grand boulevard,City :338','','no phone','ids/33.jpg','177647386',0,1,''),
	(34,'AMARENE El Yazid','Algeria,Bouira,citÃ© Iguenan,Street : Rue,House : nÂ° 10','','no pone','ids/34.jpg','307399147',0,1,''),
	(35,'MERROK LYES','Algeria, Bouira,citÃ© de la rÃ©volution, Rue sociÃ©tÃ© gÃ©nÃ©rale,House : nÂ° 100','','no phone','ids/35.jpg','142760853',0,1,''),
	(36,'CHOUBANE Fatiha','Algeria, Bouira,Saharidj Centre,Saharidj Centre,House : nÂ° 11','','no phone','ids/36.jpg','186482880',0,1,''),
	(37,'Guemmoun brahim','Algeria,Oum el bouaghi,Ain melila,Ain melila,Belabed Al-Jilali District 18, Ain Melila Residence, Door No. 8','','no phone','ids/37.jpg','177122150',0,1,''),
	(38,'Guemmoun imadeddine','Algeria,Oum el bouaghi,Ain melila,Belabed Al-Jilali District 18, Ain Melila Residence, Door No. 8','','no phone','ids/38.jpg','177414047',0,1,''),
	(39,'Belaouira samir','Algeria,Jijel,boulkali ali commune sidi maarouf wilaya jijel  Postal code :18018','','no phone','ids/39.jpg','177277054',0,1,''),
	(40,'Manaa bilal','Algeria,jijel,sidi marouf,Lot n:01. sidi marouf jijel 18018','','no phone','ids/40.jpg','313393532',0,1,''),
	(41,'khebbaz fairouz','Algeria,djelfa,City guenani Bloc 200 N 24. Djelfa Postal code :17000','','no phone','ids/41.jpg','182283371',0,1,''),
	(42,'Merayeh Khelifa','algeria,Constantine,Al-Qamas District, Bin Mahmoud Land, No. 16, Constantine 25073','','no phone','ids/42.jpg','187516218',0,1,''),
	(43,'Metai abdelkader ','Algeria,Annaba,District 140, Building 16, No. 123, Rahal Annaba 23000','','no phone','ids/43.jpg','312363278',0,1,''),
	(44,' boutellar dawoud','Algeria, Jijel,boulkali ali commune sidi maarouf wilaya jijel','','no phone','ids/44.jpg','313491002',0,1,''),
	(45,' Temmar Med Tarik',' Algeria,algiers,sidi abdellah,City Q12 llp 1200 housing block A1 number 12 mehalma sidi Abdallah 16093','','no phone','ids/45.jpg','309888075',0,1,''),
	(46,'MEBARKI MESSAOUD','','','+213 799916303','ids/46.png','166340034',0,1,''),
	(47,'BENGAS MESSAOUD ','','','+213 799916303','ids/47.png','187340147 ',0,1,''),
	(48,'NEDJAI AMINE ','','','+213 799916303','ids/48.jpg','304615189 ',0,1,''),
	(49,'MANI AHMED','','','+213 799916303','ids/49.png','307708651',0,1,''),
	(50,'BENBOUZID SALAH EDDINE','','','+213 799916303','ids/50.png','307779416 ',0,1,''),
	(51,'MEBARKI KAMEL EDDINE ','','','+213 799916303','ids/51.png','313827849',0,1,''),
	(53,'KHENNOUCI ZOHIR','dare bead','','0554860518','ids/53.png','197047861',0,1,''),
	(54,'ATALLAH DJELLOUL','Country :Algeria Wilaya : bouira City : bouira Street : lot hocini         House no : 03 Code postal :10001','','0552663908','ids/54.png','406072312',0,1,'for company use'),
	(55,'MAKHLOUFI SAMIRA','Country :Algeria Wilaya : bouira City : bouira Street : lot Karim Belkacem                  House no : 157 Code postal :10001','','0552663908','ids/55.png','118273686',0,1,'for company use'),
	(56,'MAKHLOUFI FATIMA','Country :Algeria Wilaya : bouira City : bouira Street : lot Karim Belkacem                  House no : 157 Code postal :10001','','0552663908','ids/56.png','177191733',0,1,'for company use'),
	(57,'ATALLAH NOUREDDINE','Country :Algeria Wilaya : bouira City : bouira Street : lot hocini         House no : 03 Code postal :10001','','0552663908','ids/57.jpg','168877237',0,1,'fo company use'),
	(58,'SEKSAOUI SARA','Street of El alia nursery Oued Smar Algeria door number 39  cp 16051','','0552663908','ids/58.jpg','177423646',0,1,'for company use'),
	(59,'FERHAT RADIA','Street of El alia nursery Oued Smar Algeria door number 39  cp 16051','','0552663908','ids/59.jpg','177425480',0,1,'for company use'),
	(60,'SEKSAOUI LYAZID','Street of El alia nursery Oued Smar Algeria door number 39  cp 15061','','0552663908','ids/60.jpg','177420012',0,1,'for company use'),
	(61,'SEKSAOUI ISHAK','Street of El alia nursery Oued Smar Algeria door number 39 cp 16051','','0552663908','ids/61.jpg','178006258',1,1,'for company use'),
	(62,'SEKSAOUI WASSIM','Street of El alia nursery Oued Smar Algeria door number 39 cp 16051','','0552663908','ids/62.jpg','309678828',0,1,'for company use'),
	(63,'MERHAB MOUADH','door 12 rue AEK BEN ABDELDJABAR ZONE 8 MASCARA 29000 ','','0552663908','ids/63.jpg','309669058',0,1,'for company use'),
	(64,'MERHAB MOHAMED','door 12 rue AEK BEN ABDELDJABAR ZONE 8 MASCARA 29000 ','','0552663908','ids/64.jpg','309656615',0,1,'for company use'),
	(65,'TOUHAMI ALI','door 12 rue AEK BEN ABDELDJABAR ZONE 8 MASCARA 29000 ','','0552663908','ids/65.jpg','156306817',0,1,'for company use'),
	(66,'BOUKABOUS FOUZIA','door 12 rue AEK BEN ABDELDJABAR ZONE 8 MASCARA 31000 ','','0552663908','ids/66.jpg','312614806',0,1,'for compay use'),
	(67,'MERHAB NOUREDDINE','door 12 rue AEK BEN ABDELDJABAR ZONE 8 MASCARA 29000','','0552663908','ids/67.jpg','189864111',0,1,'for company use'),
	(68,'KHENNOUCI OUALID','DAR EL BEIDA','','0797529235','ids/68.jpg','304719946',0,1,''),
	(70,'KHENNOUCI RAMI','EL MARRACH','','0554860518','ids/70.jpg','156932637',0,1,''),
	(71,'KHENNOUCI YOUCEF','DAR EL BEIDA','','0554860518','ids/71.jpg','307806133',0,1,''),
	(72,'lamine',NULL,NULL,'please provide mobile',NULL,NULL,1,1,NULL),
	(73,'FATMI NASSIM','Country:Algeria Wilaya : setif City: saleh bey Street : village ouled bouselama House no : Code postal : 19013',NULL,'no mobile','ids/73.jpg','302668285',0,1,NULL),
	(74,'DAHMANE HALIMA','Country:Algeria Wilaya : Alger City :ouled mandil DOUERA Street : 7032logts Bat 9 House no :08 Code postal :16049',NULL,'0550146720','ids/74.jpg','169321854',0,1,NULL),
	(75,'SEDDOUKI SMAIN',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(76,'MADI BOUMEDIENE',NULL,NULL,'0699-06-88-66','ids/76.jpg','168973949',0,1,NULL),
	(77,'KERDJA ZAHIA',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(78,'HIJAZI CHEMSEDDINE',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(79,'CHERFI MOHAMMED',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(80,'BEN AICHA BARIZA',' Country : Algeria  Wilaya : Batna Street :Adl Buildings District 1000 Residence Building H1 Door No. 01 Park Aforage Batna Postal code : 05000',NULL,'no mobile','ids/80.jpg','302212720',0,1,NULL),
	(81,'CHEFROUR NAAMANE',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(82,'GEROUICHA YOUCEF',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(83,'KAMAL MIDETTE',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(84,'MENASRIA RIDHA','blida','','0661490019','ids/84.jpg','313922020',1,1,''),
	(85,'ABDOUN ABDELMALEK','BAB ELOUED ALGIERS','','0770125406','ids/85.jpg','153160147',1,1,''),
	(86,'OUADI MOHAMED OUALID','DAR EL BEIDA ALGIERS','','0554676748','ids/86.jpg','308715913',1,1,''),
	(87,'TORCHI AHMED',NULL,NULL,'0779529852','ids/87.jpg','312039498',0,1,NULL),
	(88,'NASRI MOHAMMED',NULL,NULL,'+213796148838','ids/88.jpeg','178055029',0,1,'can\'t edit'),
	(89,'MEZIANE DJAOUAD',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(90,'MEZIANE SEDDAM','BATNA',NULL,'no mobile','ids/90.jpg','302160235',0,1,NULL),
	(91,'JAZIRI SOFIANE',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(92,'MEZIANI ZOULIKHA','BATNA',NULL,'no mobile','ids/92.jpg','192044233',0,1,NULL),
	(93,'CHIAHI WALID',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(94,'JIJEL MIMOU CLIENTS',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(98,'SOLTANI MERIEM',NULL,NULL,'no mobile','ids/98.jpg','187435322',0,1,NULL),
	(99,'SAYAD MOKHTAR',NULL,NULL,'no mobile','ids/99.jpg','176995083',0,1,NULL),
	(100,'DJABALI YOUCEF',NULL,NULL,'no mobile','ids/100.jpg','19731285',0,1,NULL),
	(102,'TAMAR MED TAREK',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(103,'DJABALI ATHMANE',NULL,NULL,'no mobile','ids/103.jpg','313430625',0,1,NULL),
	(104,'BOUKEDIRA YOUSSOUF',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(105,'BELGROUN AHMED','BELABBES',NULL,'no mobile','ids/105.jpg','305956559',0,1,NULL),
	(106,'BAGHDADI SAHEB BELGROUNE',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(107,'ATTABI ABDELHAMID','AIN DEFLA',NULL,'0771768386','ids/107.jpg','313137156',0,1,NULL),
	(108,'MATAOUI ABDELHAKIM','Country : Algeria Wilaya : SETIF City :  GUELLAL Adress : AL DJEBBAS City GUELLAL SETIF ALGERIA Door : none Postal code : 19201 ',NULL,'0667165385','ids/108.jpg','176453773',1,1,NULL),
	(109,'GARTI MOKRANE','Country:Algeria Wilaya : Alger City: Draria Street : cooperative el balda sebala House no :36 Code postal : L6050',NULL,'no mobile','ids/109.jpg','166701534',0,1,NULL),
	(110,'FERHATI YAHIA',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(111,'ZENATI DJAMEL',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(114,'BENKERDAL OMAR',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(115,'BENDERKAL OMAR',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(116,'BESSOU 44',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(118,'TAIBI WALID',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(119,'ASEKRI ABDELMOUMEN',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(120,'TOUAT MOURAD','BORDJ MENAIEL',NULL,'no mobile','ids/120.jpg','168947117',0,1,NULL),
	(121,'MASMOUDI HAMZA',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(122,'KHEDDAR MOHAMMED',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(123,'CHOUDER MOHAMED TAHAR','BIR MOURAD RAIS ALGIERS','','0661632094','ids/123.jpg','313824572',0,1,''),
	(124,'SENIGRA ABDELDJALIL',NULL,NULL,'no mobile','ids/124.jpg','176082069',0,1,NULL),
	(125,'OUMERZOUK ADEL','Country : Algeria Wilaya :Algers City : Algers Adress : Bir Mourad Rais Paid in Algeria Postal code : 16013',NULL,'+213552 36 09 57','ids/125.jpg','169274340',0,1,NULL),
	(126,'BOUMAZA KAMEL /MIMOU',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(127,'BABA AMMAR MOHAMMED',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(128,'MESSALTI YAHIA-AYMEN','â€Country : Algeria â€Wilaya : SETIF â€City :  GUELLAL â€Adress : AL DJEBBAS City GUELLAL SETIF ALGERIA â€Door : None  â€Paid in Algeria â€Postal code : 19201','','0667165385','ids/128.jpg','304501582',0,1,''),
	(129,'BOUGERA AMIR',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(131,'LAYACHI RECHDA',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(132,'MASSI BOUIRA',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(134,'MERZOUK ABDELWAFI','Country : Algeria  Wilaya : Bouira City : citÃ© de la rÃ©volution  Street : Rue sociÃ©tÃ© gÃ©nÃ©rale  House : nÂ° 100 Code postal : 10000','','no mobile','ids/134.jpg','179854858',0,1,''),
	(135,'BELHAMRI YOUNES',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(136,'ADDAR FARES','Country : AlgÃ©ria Wilaya : Bouira  City : grand boulevard  City :338 House :03 Code postal :10000','','no mobile','ids/136.jpg','313647315',0,1,''),
	(137,'MAHAMMED OUSAID AOMAR','Country : AlgÃ©ria  wilaya : Bouira , LAKHDARIA Street : CITE 56 LOGTS  House: num 8 Code postal :10002','','no mobile','ids/137.jpg','169389369',0,1,''),
	(140,'ZIANE NADIA','Country : Algeria  Wilaya : Bouira  Street :rue 56 City :thabet  House:2 Code postal :10000','','no mobile','ids/140.jpg','187167032',0,1,''),
	(141,'MERZOUK FOUAD','Country : Algeria  Wilaya : Bouira City : citÃ© de la rÃ©volution  Street : Rue sociÃ©tÃ© gÃ©nÃ©rale  House : nÂ° 100 Code postal : 10000','','no mobile','ids/141.jpg','305129368',0,1,''),
	(142,'BOUREGA AMIR','Country : Algeria Wilaya : Touggourt City :  Touggourt Adress : City 630 LGTS N6 Touggourt  Door : 212  Paid in Algeria Postal code : 30002 ','','+213696869543','ids/142.jpg','182247300',0,1,''),
	(143,'RAHMANI BENYOUCEF','Country : Algeria Wilaya : Algers City :  Deraria Adress : 300 Sebala Sreet  Paid in Algeria Postal code : 16050 ','','+213 661 63 20 94','ids/143.jpg','312425502',0,1,''),
	(144,'DINE IMANE','Country : Algeria Wilaya : Algers City :  Deraria Adress : 300 Sebala Sreet  Paid in Algeria Postal code : 16050','','+213 661 63 20 94','ids/144.jpg','312410775',0,1,''),
	(145,'KHEDDAR LEILA','Country : Algeria Wilaya : MEDEA  City :  Medea  Adress : Medea   Paid in Algeria Postal code : 26000','','+213 794 25 43 04','ids/145.jpg','313828146',0,1,''),
	(146,'FILALI  NOUR EL HOUDA','Country : Algeria Wilaya :Constenrine  City : Constenrine Adress : Constenrine Postal code : 25004','','+213 550 53 27 15','ids/146.jpg','156640169',0,1,''),
	(147,'TEBBOUB ZINEB','Country : Algeria Wilaya :Constenrine  City : Constenrine Adress : Constenrine Postal code : 25004','','+213 550 53 27 15','ids/147.jpg','306139009',0,1,''),
	(148,'ZOUAGHI YOUCEF','Country : Algeria Wilaya :Constenrine  City : Constenrine Adress : Constenrine Price : 2 650 000.00DZD Paid in Algeria Postal code : 25004','','+213 550 53 27 15','ids/148.jpg','186178833',0,1,''),
	(149,'HAMMADI RIAD','Country : Algeria Wilaya :Constenrine  City : Constenrine Adress : Constenrine Price : 2 650 000.00DZD Paid in Algeria Postal code : 25004','','+213 550 53 27 15','ids/149.jpg','197110394',0,1,''),
	(150,'TÃIBI WALID','Country : Algeria Wilaya :Algers City : Bordj El Kifane  Adress : Algers Price : 2 300 000.00DZD Paid in Algeria Postal code : 16006','','+213 661 98 04 28','ids/150.jpg','304918543',0,1,''),
	(151,'BABAAMER DJELMAM MOHAMMED',' Country : Algeria Wilaya :Algers City : Algers Adress : Algers Paid in Algeria Postal code : 16006','','+2137 79 22 55 81','ids/151.jpg','312496817',0,1,''),
	(152,'MEZIANI TOUFIK','BEJAYA','','0661832903','ids/152.jpg','403519952',0,1,''),
	(153,'mimou','','mimou@merhab.com','no mobile',NULL,NULL,1,0,''),
	(154,'ATTOUI ADEL','Country : Algeria Wilaya :Annaba City : Berahel Address:Hasiri Al-Sharif neighborhood, Berhal, Annaba Payed in Algeria  Contract in Algeria  Postal code : 23009','','no mobile','ids/154.jpg','308773854',0,1,''),
	(155,'GUERRAS MOHAMED HAITHEM','Country:Algeria Wilaya : Constantine City : El khroub Street: The address: subdivision Massinissa Num 81 25005 ','','no mobile','ids/155.jpg','309828212',0,1,''),
	(156,'ZEGROUR YASSER','Country : Algeria Wilaya Jijel City :boulkali ali commune sidi maarouf wilaya jijel  Postal code :18018 ','','no mobile','ids/156.jpg','308499107',0,1,''),
	(157,'BOUZERAIB ABDELGHANI','Country: Algeria  Wilaya : jijel City : sidi marouf Street :B p n:1 c sidi marouf w jijel Postal code : 18018','','no mobile','ids/157.jpg','308505597',0,1,''),
	(160,'MERAIHIA HOUSSEM EDDINE','Country : Algeria Wilaya: Constantine City : Sissaoui neighborhood, Constantine Maraih Khalifa, Al-Qammas neighborhood, Ben Mahmoud land, No. 16, Constantine 25073 ','','no mobile','ids/160.jpg','309718887',0,1,''),
	(161,'GHETTOUT YOUSSOUF','Country : Algeria Wilaya Jijel City :boulkali ali commune sidi maarouf wilaya jijel  Postal code :18018 ','','no mobile','ids/161.jpg','177960098',0,1,''),
	(162,'MESMOUDI HEMZA','OULED DJELLAL','','0771.78.34.09','ids/162.jpg','303017411',0,1,''),
	(163,'BELGROUN SANAA','BELABBAS','','no mobile','ids/163.jpg','411835901',0,1,''),
	(164,'BELGROUN MERIEM','BELABBES','','no mobile','ids/164.jpg','413804908',0,1,''),
	(165,'BOUBEKEUR NAWEL','BELABBES','','no mobile','ids/165.jpg','113505724',0,1,'Belgroun '),
	(166,'BELGROUN YOUNES','BELABBAS','','NO MOBILE','ids/166.jpg','114088287',0,1,''),
	(167,'MEKKI SAID','BELABBES','','no mobile','ids/167.jpg','411657966',0,1,'Belgroun'),
	(168,'HAGOUG DJAMEL','BELABBES','','no mobile','ids/168.jpg','305970078',0,1,''),
	(169,'NEDJADI LEILA','BELABBES','','no mobile','ids/169.jpg','402492129',0,1,''),
	(170,'BEGHDADI SID AHMED','','','no mobile','ids/170.jpg','103785400',0,1,''),
	(171,'BOURAHLA AYOUB','EL-AFFROUN','','no mobile','ids/171.jpg','307726222',0,1,''),
	(172,'GUECHOUT CHAIMA','','','no mobile','ids/172.jpg','309307419',0,1,''),
	(173,'BENSIDI AISSA ABDERRAHMANE','','','no mobile','ids/173.jpg','107753014',0,1,''),
	(174,'TOUMI RIADH','','','no mobile','ids/174.jpg','307904066',0,1,''),
	(175,'BEGHDADI ISMAIL','','','no mobile','ids/175.jpg','108275643',0,1,''),
	(176,'ASKRI ABDELMOUMEN','MEDIA','','no mobile','ids/176.jpg','313829885',0,1,''),
	(177,'KHEDDAR MOHAMED','MEDIA','','no mobile','ids/177.jpg','196732487',0,1,''),
	(178,'KHEDDAR HADJER','MEDIA','','no mobile','ids/178.jpg','186469313',0,1,''),
	(179,'DAMERDJI RABEA','MEDIA','','no mobile','ids/179.jpg','169621171',0,1,''),
	(180,'DJABALI SLIMANE','','soltani@me.com','0672618729',NULL,NULL,1,0,''),
	(181,'BOUNAB TOUFIK','','','no mobile','ids/181.jpg','166826289',0,1,''),
	(182,'EUTAMENE SALAH','','','no mobile','ids/182.jpg','177871858',0,1,''),
	(183,'BABOURI SOUHIL','','','no mobile','ids/183.jpg','155986297',0,1,''),
	(185,'MEZIANE ABDELDJOUAD','BATNA','','no mobile','ids/185.jpg','303247655',0,1,''),
	(187,'NAIT- SI- MOHAND MESSAOUD','','nait@me.com','no mobile','ids/187.jpg','176301258',1,1,''),
	(188,'SELATNIA NOUREDDINE','GUELMA','','no mobile','ids/188.jpg','157270147',0,1,''),
	(189,'ZIOUCHE MAROUA','ANNABA','','no mobile','ids/189.jpg','169005665',0,1,''),
	(190,'SELATINIA ZOUHEIR','ANNABA','','no mobile','ids/190.jpg','306108325',0,1,''),
	(191,'SADOU MOUSSA','AIN DEFLA','','no mobile','ids/191.jpg','197470640',0,1,''),
	(192,'ATTABI HADJER','AIN DEFLA','','no mobile','ids/192.jpg','167771745',0,1,''),
	(193,'SAIDIHIA DJILLALI','AIN DEFLA','','no mobile','ids/193.jpg','196219607',0,1,''),
	(194,'SAIDIHIA BADREDDINE','AIN DEFLA','','no mobile','ids/194.jpg','196944972',0,1,''),
	(195,'SLIMANE AICHA','AIN DEFLA','','no mobile','ids/195.jpg','166561918',0,1,''),
	(196,'GHENNAM ASMA NOR ELHOUDA','AIN DEFLA','','no mobile','ids/196.jpg','169216653',0,1,''),
	(197,'SAIDIHIA REDOUANE','AIN DEFLA','','no mobile','ids/197.jpg','313624847',0,1,''),
	(198,'AKMOUN BACHIR','AIN DEFLA','','no mobile','ids/198.jpg','197641134',0,1,''),
	(199,'ALI KOUIDER MOHAMED OUALID','AIN DEFLA','','no mobile','ids/199.jpg','168670367',0,1,''),
	(200,'CHERIFI MOHAMMED','Country:Algeria Wilaya : Bouira City: bouira Street : 400 logt AADL Bat 07 House no : 26 Code postal : 10000','','0560.01.61.84','ids/200.jpg','114838747',0,1,''),
	(201,'BOULHARES Nasr-eddine','BOUIRA','b@me.com','0655-30 -32-99',NULL,NULL,1,0,''),
	(202,'ACHIT Amine Mohammed','Country :Algeria Wilaya : Bouira City: El hachimia Street: VIg s\'dara House no: 47 Code postal: 10023','','no mobile','ids/202.jpg','117715305',0,1,''),
	(203,'RAKHOUAN E Abderrahmane','CountrY:Algeria Wilaya : Bouira City: Guerrouma Street: ain el baida House no :43 Code Postal :10230','','no mobile','ids/203.jpg','104220667',0,1,''),
	(204,'KHODJA Rabah','','','no mobile','ids/204.jpg','116482400',0,1,''),
	(205,'KAMIRI Mustapha','Country:Algeria Wilaya : Bouira City: Bouira Street: Said abid bouira House no :28 Code postal : 10084','','no mobile','ids/205.jpg','115589418',0,1,''),
	(206,'CHOUIA Salim','Country : Algeria Wilaya : Alger City: bordj El kiffan street: smail sidhom lot D House : n\" 22 Code postal :L6223','','no mobile','ids/206.jpg','108629994',0,1,''),
	(207,'ATALLAH SOUHILA','','','no mobile','ids/207.jpg','1773888713',0,1,''),
	(208,'HAMDACHE ACHOUR','Councry :Algeria Wllava: bouira Cir: bouira Screel : 66 LOGT BAT BTHABAT House : nÂ° 2 Code poscal: 10000','','no mobile','ids/208.jpg','105091804',0,1,''),
	(209,'ATALLAH samir','Counlry: Algerla Wilava: bouira CIcy: bouira Sleet :83/200/20D0 logLbal G Mousc: nÂ° 7 Code postal: 10000','','no mobile','ids/209.jpg','403177747',0,1,''),
	(210,'BENCHIKH ZAKIA','','','no mobile','ids/210.jpg','169452925',0,1,''),
	(211,'BENHEFFAF Kheir eddine','Country:Algeria Wilaya : djelfa City: Djelfa Street : la nouvelle rue bloc 135 House no : l-l- Code postal : 1-7000','','07 70-59-5 9 -6 1','ids/211.jpg','302012204',0,1,''),
	(212,'BENHEFFAF Ahmed','DJELFA','','0770595561','ids/212.jpg','182094900',0,1,''),
	(213,'KADRI RAYEN ALI',' Country:Algeria Wilaya: ORAN r City : oran Street :27 rue si mossadeq ,  ler etage House no : A Code postal :31000 ','',' No Mobile','ids/213.png','302185679',0,1,''),
	(214,'KADRI BENAMEUR',' Country :Algeria Wilaya : ORAN City : Bir El Djir Street : Quartier Khemisti, rue Chark, 5eme Ã©tage  House no : 34 Code postal : 31000 ','','0','ids/214.png','202257757',0,1,''),
	(215,'BEN HEFFAF ABDELKARIM A','Country:Algeria Wilaya : djelfa City : Djelfa Street: la nouvelle rue bloc 135 House no : l-l- Code postal : 17000','','0770-59-55-61','ids/215.jpg','302841657',0,1,''),
	(216,'CHOUAF HAFIDA','Country : Algeria Wilaya :Algers City : Bordj El Bahri Adress : City 20 AoÃ»t 1956 Terre Familliale Est  Villa : N; 137  Bordj el Bahri  Postal code : 16046 ','','+213770182939','ids/216.jpeg','105169349',0,1,''),
	(217,'HAMMADI MAHDI','52 Street HASSIBA BEN BOULI Algers 16014','','+213770182939','ids/217.jpeg','312763268',0,1,''),
	(218,'KADRI MOHAMED ABDELWAHID','\"Country :Algeria Wilaya : ORAN City : oran Street : Rue normandie   House no : 27 Code postal : 31000\"','','0','ids/218.jpg','156814378',0,1,''),
	(219,'BENIAICH KHADIDJA','Bordj el bahri','','+213770182939','ids/219.jpeg','304813773',0,1,''),
	(220,'GARTI Mohamed','Country:Algeria Wilaya:Bouira City: Bouira Street : si zoubir Code postal : 10000','','0655 30 32-99','ids/220.jpg','197458222',0,1,''),
	(221,'HACHANI HADIA','Bordj el bahri','','+213770182939','ids/221.jpeg','186929772',0,1,''),
	(223,'TOUATI YOUNES','Bordj el bahri','','+213770182939','ids/223.jpeg','197117245',0,1,''),
	(224,'SADOUKl smain','Country:Algeria Wilaya :chelf City : bir esafsaf Street : oued foudda House no : 30 Code postal :02034','','0674952881','ids/224.jpg','308539153',0,1,''),
	(226,'FERHATI YAHYA','','','0662249586','ids/226.jpg','100823559',0,1,''),
	(227,'NOUARI ALI',' Country :Algeria Wilaya : Taref City : chihani, drean  Street : chÃ¢teau d\'eau  House no :  Code postal : 36210 ','','0','ids/227.jpg','304580078',0,1,''),
	(228,'GHANNAI MARWA ','Country:Algeria Wilaya : Taref City : chihani. Street : sehili amara House no : Code postal : 360L4 ','','0','ids/228.png','187006395',0,1,'Car model : Coolray chaoneng auto'),
	(229,'RAHEM ILYES',' Country :Algeria Wilaya :Annaba City : annaba. Street : didouche mourad House no : 55 Code postal : 23000','','0','ids/229.jpg','197777823',0,1,''),
	(230,'ZENNATI Djamel','CountrY:Algeria Wilaya : bouira City: ras bouira Street : citÃ© 30 logts House no :07 Code Postal : 10103','','no mobile','ids/230.jpg','112263313',0,1,''),
	(231,'SARI HAMID','Country:Algeria Wilaya : Bouira City: bouira Street : 168 logt AADL BT D House no : 13 Code postal : 10000','','no mobile','ids/231.jpg','121202076',0,1,''),
	(232,'BOUAMRA NOUR EDDINE',' Country :Algeria Wilaya : TÃ©bessa City : Bir ElAter Street : Quartier de l\'AÃ©roport ( hai el matar) House no :38 Code postal : 12000','','0','ids/232.jpg','309896817',0,1,''),
	(233,'AYAT khirredine','Country :Algeria Wilaya : setif City : ain el kbira Street : quartier 3L parcelle deraa eltaraa House no :6 Code postal : 19008 Car model : Geely emgrand','','no mobile','ids/233.jpg','105270423',0,1,''),
	(234,'ZAKARIA YOUCEF',' Country:Algeria Wilaya :bouira City : bouira  . Street : rÃ©sidence AFZIM House no : 34 Code postal : L0000','','0','ids/234.jpg','118104881',0,1,''),
	(235,'BOUHABIB Mohamed el hadi','Country:AIgeria Wilaya : Constantine City: ain el bey Street: citÃ© El fedj zouaghi House no :86 Code postal :25000','','no mobile','ids/235.jpg','308779557',0,1,''),
	(236,'TALI BRAHIM','\"Country :Algeria Wilaya : bouira City : AIN BESSAM Street : hai senoussi ali House no : 10 Code postal : 10400\"','','0','ids/236.jpg','308315470',0,1,''),
	(237,'MOGAFI HAMZA','ELOMARIA','','0','ids/237.JPG','314176910',0,1,''),
	(238,'AMAROUCHE DJABER','14 STREET DOUZI MOHAMMED BORDJ EL KIFANE Algers ','','+213661996418','ids/238.jpeg','309634923',0,1,''),
	(239,'MEZAOUGH KHALID','El Djorf City Bab ezzouar Algiers','','+213770982382','ids/239.jpeg','166073372',0,1,''),
	(240,'FEKRACHE HEMZA','Aboudi Ahmed city number 162 didouche mourad Constentine','','213661 71 66 53','ids/240.jpeg','312648868',0,1,''),
	(241,'KIFOUCHE SAMIR','Zighoud Youcef Constantine','','213661716653','ids/241.jpeg','304741810',0,1,''),
	(242,'bazer adil','ghelma','','no mobile','ids/242.JPG','309488953',0,1,''),
	(243,'KOUACEMA HOUSSAM','','','no mobile','ids/243.jpg','312640791',0,1,''),
	(244,'ABEDI MOHAMED FAYCEL','','','no mobile','ids/244.jpg','186264803',0,1,''),
	(245,'BEKHOUCHE SAID','Sidi Abdellah City Zeralda Algeria ','','+213560738705//0555838181','ids/245.jpeg','309753819',0,1,''),
	(246,'BOUGOUSSA ABDELDJALIL','','','0660828160','ids/246.jpg','410831897',0,1,''),
	(247,'ATTAB RAMZI',NULL,NULL,NULL,NULL,NULL,0,1,NULL),
	(248,'FATHI SARAH','562 City 204 LGTS Ain Timouchent 46000','','+213797571031','ids/248.jpeg','302463041',0,1,''),
	(249,'MOUSSA MOHAMMED','AIN TADLES','','+213782067328','ids/249.jpeg','169336056',0,1,''),
	(250,'BRAHMI DJAMAL EDDINE','','','NO','ids/250.jpg','105975608',0,1,''),
	(251,'BELABIOD LARBI','','','NO','ids/251.jpg','309286713',0,1,''),
	(252,'BENHADIDA KHADIDJA','','','NO','ids/252.jpg','309652955',0,1,''),
	(253,'ZEAR YAMINA','','','NO','ids/253.jpg','308615130',0,1,''),
	(254,'KEROUICHA YOUCEF','','','NO','ids/254.jpg','305128216',0,1,''),
	(255,'BENDJEBBOUR SAKINA','','','NO','ids/255.jpg','102490151',0,1,''),
	(256,'KEROUICHA OMAR YACINE','','','NO','ids/256.jpg','414190458',0,1,''),
	(257,'BOUALEM LINEDA ZAKIA','','','NO','ids/257.jpg','106522254',0,1,''),
	(258,'AMRI BILLEL','DAR EL BEIDA','','+213560151576','ids/258.jpeg','A00290946',0,1,''),
	(259,'GHEMARI AYOUB','','','NO','ids/259.jpg','154787639',0,1,''),
	(260,'TOUATI RIADH','','','NO','ids/260.jpg','177748258',0,1,''),
	(261,'ABDELLI TAREK','ROUIBA Algeirs ALGERIA','','+213560151576','ids/261.png','186961128',0,1,''),
	(262,'RACHID ALI','ROUIBA Algiers Algeria','','+213560151576','ids/262.jpg','110476226',0,1,''),
	(263,'DRAGUENDOUL KHEIREDDINE','Rouiba Algiers Algeria','','+213560151576','ids/263.jpg','411271838',0,1,''),
	(264,'RAHMOUNE MOHAMED MOUNIR','','','NO','ids/264.jpg','100030558019820009',0,1,''),
	(265,'KEMITI RACHID','Rouiba Algiers Algeria','','+213560151576','ids/265.jpg','401776494',0,1,''),
	(268,'KEMITI OMAR','','','NO','ids/268.jpg','406035050',0,1,''),
	(269,'RAHMOUNE ABDERRAOUF','','','NO','ids/269.jpg','115899314',0,1,''),
	(270,'KADRI ZOUHIR','Makhloufi mohammed City EL OMARIA Media 26008','','+213662200609//+213550385556','ids/270.jpeg','314002235',0,1,''),
	(272,'SLIMANI AZZEDDINE','Bordj el Bahri Algiers Algeria','','+213770182939','ids/272.jpeg','155610544',0,1,''),
	(274,'SAYOUD REBIHA','','','+213770182939','ids/274.jpeg','197800856',0,1,''),
	(275,'LEMGHARBI MOHAMMED MOKHTAR','','','+213668743466','ids/275.jpeg','197023053',0,1,''),
	(276,'SACI LOTFI AYOUB','CHAHID LATTI HADJ AHMED CITY 01 SEBDOU TLEMCEN','','+213540932401','ids/276.jpeg','197523736',0,1,''),
	(277,'TIBOUCHE ISLAM','','','no mobile','ids/277.jpg','312172642',0,1,''),
	(278,'MILI REDHA','','','no mobile','ids/278.jpg','309882207',0,1,''),
	(279,'NOUIOUA ABDELOUEHAB','','','no mobile','ids/279.jpg','312539705',0,1,''),
	(280,'BOUDEN LAZHAR','','','no mobile','ids/280.jpg','306936221',0,1,''),
	(281,'CHEGGA OUAIL','','','no mobile','ids/281.jpg','177127001',0,1,''),
	(282,'MOUSSA FARES','','','no mobile','ids/282.jpg','104628654',0,1,''),
	(283,'HARRAR ZAKARIA','Cherchell tipaza 42100','','+213540932401','ids/283.jpeg','169547747',0,1,''),
	(284,'ZEGGADA SOFIANE','Reghaia Algiers 16036','','+213770182939','ids/284.jpeg','313263869',0,1,''),
	(285,'HACHANI HOCINE','Bordj el Bhari Algiers Algeria','','+213770504210','ids/285.jpeg','196905186',0,1,''),
	(286,'BELLOT OMAR','Q12 LPP CITY 1200 LGTS Mehalma Algiers ','','+213550470905','ids/286.jpeg','313665695',0,1,''),
	(287,'ARRAR LAKHDAR','El Eulma Setif','','+213557212404','ids/287.jpeg','167261685',0,1,''),
	(288,'ARRAR ABDELMALEK','EL Eulma setif','','+213557212404','ids/288.jpeg','156741182',0,1,''),
	(291,'CHEFROUR NAAMAN','SOUG AHRAS Algeria','','+213699017976','ids/291.jpeg','177161810',0,1,''),
	(292,'HASSIB HAMID','Deraria Algiers 16050','','+213797396830','ids/292.jpeg','304911085',0,1,''),
	(293,'ALISMAIL MOHAMMED','5 July 1962 Street sidi lakhdar Ain Defla postal code 44027','','+213771768386','ids/293.jpeg','110034796',0,1,''),
	(294,'HELALI AMINE HOUCINE','NO.15, BOUSHAKI BAB EZZOUAR CITY, BAB EZZOUAR, ALGERS, ALGERIA POST CODE : 16024','','+213671841324','ids/294.png','304927967',0,1,''),
	(295,'BEN ACHOUR KHEMAIS','','','â€ª+966Â 54Â 432Â 1476â€¬','ids/295.jpeg','177785041',0,1,''),
	(296,'Mohammed Khemici ',' CitÃ© 605 logts bt 24 n 03 Boudouaou -Boumerdes ','','+213  791 61 12 78  ','ids/296.png','186042344',0,1,'NIN:109791201008680007'),
	(297,'Keskes billel','Address : 10 street ben demegh khier city SÃ©tif ','samibilel102@yahoo.fr','+213771678665','ids/297.jpg','313290597',0,1,'NIN:109810675005070006'),
	(298,'MAIDAT ABDENOUR','Country :Algeria Wilaya : Alger City : Bordj el Bahri Street : Rue Houari boumedinne House no : 108 Code postal : 16046','','+213 770 18 29 39','ids/298.jpg','176404001',0,1,''),
	(299,'BOUGUERRA ABDELHAK','CITY 40 LGTS STAOUALI ALGIERS POSTAL CODE16062','','+213661885285','ids/299.jpeg','177312685',0,1,''),
	(300,'NEMDIL ABDELBARI','6 YEKHLEF ALI STREET SOUMAA BLIDA POSTAL CODE 09022','','+213658023099','ids/300.jpeg','176694795',0,1,''),
	(301,'BENTRAD ABDENNOUR','Ø­ÙŠ 364 Ù…Ø³ÙƒÙ† Ù…Ø¯Ø®Ù„ 21 Ø±Ù‚Ù…194 Ø§Ù„Ø­Ø¬Ø§Ø± Ø¹Ù†Ø§Ø¨Ø© 364 CITY','','+213661885285','ids/301.jpeg','400790715',0,1,'364 CITY ENTRE Nub 21 /194 EL HADJAR ANNABA 23200'),
	(302,'BOUKORTT MOHAMMED','MOSTAGANEM','','NO MOBILE','ids/302.png','176800789',0,1,''),
	(303,'MEDEBBER MOHAMMED','MASCARA','','NO MOBILE','ids/303.png','157033883',0,1,''),
	(304,'YEROU BELKHIR','MASCARA','','NO MOBILE','ids/304.png','407334007',0,1,''),
	(305,'SAKHRI SMAIL','','','NO MOBILE','ids/305.jpg','313929282',0,1,''),
	(306,'BENSAAD NASSIMA','','','MO MOBLIE ','ids/306.jpg','168771967',0,1,''),
	(307,'BOULTIF MOHAMED LAMINE','','','NO MOBILE','ids/307.jpg','172374410',0,1,''),
	(308,'Mebarka rabie','citÃ©  822 logement amirouche a11 n 14 reghaia algeria  16036','rabiemeb810@gmail.com','+213661878198','ids/308.jpg','169502027',0,1,''),
	(309,'AIT MOHAMED KARIM','','','0660100899','ids/309.jpg','156682957',0,1,''),
	(310,'DJEBBARI ALLAOUA','','','0699756143','ids/310.jpg','168841951',0,1,''),
	(311,'GHERAISSA DJAMEL','','','NO MOBILE','ids/311.jpg','157278491',0,1,''),
	(316,'TAIHI SOFIANE','','','NO MOBILE','ids/316.jpg','166682269',0,1,''),
	(317,'HAMMOUDI SI MOHAMMED','','','no mobile','ids/317.jpg','187247980',0,1,''),
	(320,'BOUZENOURA HALA ','48 EL MOUSTAKBEL Street oued el romane Algiers Algeria 16106','','+213671107264','ids/320.jpeg','115984866',0,1,''),
	(325,'ABERKANE YASSINE','','','550285709â€¬','ids/325.png','109077682',0,1,''),
	(326,'ZAIDI OUSSAMA RACHID','','','550285709â€¬','ids/326.JPG','196180746',0,1,''),
	(327,'BOUKERROUCHE MOHAMED','','','550285709â€¬','ids/327.png','406525253',0,1,''),
	(328,'BELLI SOUFIANE','Bordj el Bahri Algiers postal code 16046','','+213770182939','ids/328.jpeg','309307657',0,1,''),
	(329,'BARKAT LAKHDAR','Bordj el Bahri Algiers 16046','','+213770182939','ids/329.jpeg','186964107',0,1,''),
	(330,'MEHNOUNE FAROUK','Bordj el bahri Algiers 16046','','+213770182939','ids/330.jpeg','167587538',0,1,''),
	(331,'ACHAB SAMY ','BORDJ EL BAHRI Algiers 16046','','+213770182939','ids/331.jpeg','186822253',0,1,''),
	(332,'MESSOUDI FAHIMA','Bordj el Bahri  ALGIERS 16046','','+213770182939','ids/332.jpeg','312005665',0,1,''),
	(333,'MEKHALDI OUSSAMA','Road number 24 Bordj elbahri Sebaa City Algiers 16046','','+213552369105','ids/333.jpeg','187520333',0,1,''),
	(334,'ali chinese','','m@m.com','777777',NULL,NULL,1,0,''),
	(335,'ZOUAR AHMED','Ras Bouira ben Abdallah n 22 Bouira','','550285709â€¬','ids/335.jpeg','187381119',0,1,''),
	(336,'MAIDAT SAID','','','NO MOBILE','ids/336.jpg','313499893',0,1,''),
	(337,'OUKHAF RADHWANE','Ø­ÙŠ Ø§Ù„Ø´Ù‡ÙŠØ¯ Ø³ÙˆØ¹Ø¯ÙŠØ© Ø®Ø§Ù„Ø¯ Ù…Ø¬ 10Ø±ÙÙ… 04 Ø¯Ø§Ø±Ø§Ù„Ø¨ÙŠØ¶Ø§Ø¡ Ø§Ù„Ø¬Ø²Ø§Ø¦Ø±','','NO MOBILE','ids/337.png','196021934',0,1,''),
	(338,'BOUDJERIDA AHCENE','BERKOUKA City Djijel postal code 18000','','+213796978552','ids/338.jpeg','102383820',0,1,''),
	(339,'BOUDJEMAA RACIM','Ø­ÙŠ Ø´Ø±ÙŠÙ Ø¨ÙŠØ¯ÙŠ Ù…Ø¬Ù…ÙˆØ¹Ø© 09 Ø±Ù‚Ù… 33 Ø¨ÙˆØ±ÙˆØ¨Ø©','','NO MOBILE','ids/339.jpeg','166560322',0,1,''),
	(346,'BOUDJEMA WIAM','Ø­ÙŠ Ø´Ø±ÙŠÙ Ø¨ÙŠØ¯ÙŠ Ù…Ø¬Ù…ÙˆØ¹Ø© 09 Ø±Ù‚Ù… 33 Ø¨ÙˆØ±ÙˆØ¨Ø©','','NO MOBILE','ids/346.jpg','169636559',0,1,''),
	(347,'KEBIR BRAHIM','Ø­ÙŠ 2068 Ù…Ø³ÙƒÙ† Ø¹Ù…Ø§Ø±Ø© 52Ø¨  Ø±Ù‚Ù… 3  Ø¨Ø§Ø¨ Ø§Ù„Ø²ÙˆØ§Ø± Ø§Ù„Ø¬Ø²Ø§Ø¦Ø±','','NO MOBILE','ids/347.jpg','312037014',0,1,''),
	(348,'MEDDAH MOHAMED','Algeria,Alger,Douera,hay 1602 logement BT52,HOUSE NO : 32 , CODE POSTAL 42455','','NO MOBILE','ids/348.png','166562526',0,1,''),
	(349,'AGGOUN FATIHA','Ø­ÙŠ Ø³ÙŠ Ø§Ù…Ø­Ù…Ø¯ Ø¨ÙˆÙ‚Ø±Ø© Ø¯Ø§Ø¦Ø±Ø© ÙÙˆÙƒØ© ÙˆÙ„Ø§ÙŠØ© ØªÙŠØ¨Ø§Ø²Ø© , Ø±Ù‚Ù… Ø§Ù„Ø¨Ø§Ø¨ 106 Ø§Ù„Ø±Ù…Ø² Ø§Ù„Ø¨Ø±ÙŠØ¯ÙŠ 42440','','NO MOBILE','ids/349.png','312490030',0,1,''),
	(350,'AGGOUN ABDELHAK','Ø­ÙŠ 1350 Ù…Ø³ÙƒÙ† Ø¨Ø±Ù‚Ù… 22 Ø¹Ù…Ø§Ø±Ø© 37 Ø¨ÙˆØ³Ù…Ø§Ø¹ÙŠÙ„ ØªÙŠØ¨Ø§Ø²Ø© Code postal 4241','','NO MOBILE','ids/350.png','186803479',0,1,''),
	(351,'BEN ZERROUK ABDERREZZAK','Hay si Boualam Aloui Ain Taya alger','','NO MOBILE','ids/351.png','305332521',0,1,''),
	(352,'BELBAL ABDELAZIZ','','','+213770708735','ids/352.jpeg','302593320',0,1,''),
	(353,'OUAKER KARIM ','vilags said abid w bouira door number 22','','no mobile ','ids/353.png','169889374',0,1,''),
	(354,'GARTI LOUCIF','CHEMIN SI ZOUBIR A COTE ADE BOUIRA','atallah@merhab.com','no numbers','ids/354.JPG','304786772',0,1,''),
	(355,'bensalem aissa ','cite freres bel aissi no 17 mehalma ','atellah@merhab.com','no number','ids/355.JPG','172445870',0,1,''),
	(356,'KHEDIM ABDELKADER WISSAME','','','+213560068572','ids/356.jpg','304566117',0,1,''),
	(358,'MENHOUR ANIS','DOOR 96 RN 24 BORDJ ELBAHRI ALGIERS','','00213772141183','ids/358.jpg','308070715',0,1,''),
	(360,'LAICHE MALEK-EDDINE','144 CITY TASAHOUMI EL AGHOUAT 03000','','+213658744896','ids/360.jpeg','152320030',0,1,''),
	(361,'MEDDAH HAMZA','','','Ù‰Ø® Ø©Ø®Ù„Ø§Ù‡Ù…Ø«','ids/361.jpg','314032535',0,1,''),
	(362,'CHENAFI MOHAMED LAMINE','','','no mobile','ids/362.jpg','305215629',0,1,''),
	(363,'BOUCHACHIA NESRINE','','','no mobile','ids/363.jpg','314254863',0,1,''),
	(364,'BOUCETTA WALID','','','N/A','ids/364.jpg','308322637',0,1,''),
	(365,'ABDELHAKIM AHMED EL-AMIN','','','N/A','ids/365.jpg','177633948',0,1,''),
	(366,'KHODJA HASSENE','','','N/A','ids/366.jpg','196123396',0,1,''),
	(375,'boukhedimi charaf eddine','Ø´Ø§Ø±Ø¹ Ø§Ù„Ø£Ù…ÙŠØ± Ø¹Ø¨Ø¯ Ø§Ù„Ù‚Ø§Ø¯Ø± Ø¯Ø§Ø¦Ø±Ø© Ø³ÙŠØ¯ÙŠ Ø¹Ù„ÙŠ Ø¨ÙˆØ³ÙŠØ¯ÙŠ  ÙˆÙ„Ø§ÙŠØ© Ø³ÙŠØ¯ÙŠ Ø¨Ù„Ø¹Ø¨Ø§Ø³','','0673917817','ids/375.png','304854883',0,1,''),
	(382,'BENRAHMOUNE REDA','EL BOUNI ANNABA POSTAL CODE 23000','','+213553887750','ids/382.jpeg','304681902',0,1,'');

/*!40000 ALTER TABLE `clients` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table colors
# ------------------------------------------------------------

CREATE TABLE `colors` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `color` varchar(255) DEFAULT NULL,
  `hexa` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `color` (`color`),
  UNIQUE KEY `hexa` (`hexa`)
) ENGINE=InnoDB AUTO_INCREMENT=7 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `colors` WRITE;
/*!40000 ALTER TABLE `colors` DISABLE KEYS */;

INSERT INTO `colors` (`id`, `color`, `hexa`)
VALUES
	(1,'WHITE','#ffffff'),
	(2,'BLACK','#050505'),
	(3,'GRINARDO','#828580'),
	(4,'GREY','#8b8989'),
	(5,'CHAMILION','#7b647d'),
	(6,'PEARLY WHITE','#fbf9f9');

/*!40000 ALTER TABLE `colors` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table containers
# ------------------------------------------------------------

CREATE TABLE `containers` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `containers` WRITE;
/*!40000 ALTER TABLE `containers` DISABLE KEYS */;

INSERT INTO `containers` (`id`, `name`)
VALUES
	(1,'AKK'),
	(2,'40hc'),
	(3,'cosco');

/*!40000 ALTER TABLE `containers` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table defaults
# ------------------------------------------------------------

CREATE TABLE `defaults` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `rate` decimal(10,2) DEFAULT NULL,
  `freight_small` decimal(10,2) DEFAULT NULL,
  `freight_big` decimal(10,2) DEFAULT NULL,
  `alert_unloaded_after_days` int(11) DEFAULT NULL,
  `alert_not_arrived_after_days` int(11) DEFAULT NULL,
  `alert_no_licence_after_days` int(11) DEFAULT NULL,
  `alert_no_docs_sent_after_days` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

LOCK TABLES `defaults` WRITE;
/*!40000 ALTER TABLE `defaults` DISABLE KEYS */;

INSERT INTO `defaults` (`id`, `rate`, `freight_small`, `freight_big`, `alert_unloaded_after_days`, `alert_not_arrived_after_days`, `alert_no_licence_after_days`, `alert_no_docs_sent_after_days`)
VALUES
	(1,250.00,1600.00,2500.00,10,7,9,20);

/*!40000 ALTER TABLE `defaults` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table discharge_ports
# ------------------------------------------------------------

CREATE TABLE `discharge_ports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `discharge_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discharge_port` (`discharge_port`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `discharge_ports` WRITE;
/*!40000 ALTER TABLE `discharge_ports` DISABLE KEYS */;

INSERT INTO `discharge_ports` (`id`, `discharge_port`)
VALUES
	(3,'ALGIERS'),
	(2,'ANNABA'),
	(4,'MOSTAGANEM'),
	(1,'ORAN'),
	(5,'SKIKDA');

/*!40000 ALTER TABLE `discharge_ports` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table loaded_containers
# ------------------------------------------------------------

CREATE TABLE `loaded_containers` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_loading` int(11) DEFAULT NULL,
  `id_container` int(11) DEFAULT NULL,
  `ref_container` varchar(255) DEFAULT NULL,
  `date_departed` date DEFAULT NULL,
  `note` text DEFAULT NULL,
  `date_loaded` date DEFAULT NULL,
  `date_on_board` date DEFAULT NULL,
  `so` varchar(255) DEFAULT NULL,
  `is_released` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `loaded_containers` WRITE;
/*!40000 ALTER TABLE `loaded_containers` DISABLE KEYS */;

INSERT INTO `loaded_containers` (`id`, `id_loading`, `id_container`, `ref_container`, `date_departed`, `note`, `date_loaded`, `date_on_board`, `so`, `is_released`)
VALUES
	(1,1,1,'','2025-07-03',NULL,'2025-06-30',NULL,NULL,0),
	(2,2,3,'6422944000','2025-07-09','coolray super 4 white ','2025-07-06','2025-07-09','OOCU8716306',1),
	(3,2,3,'TRHU7452784','2025-07-09','coolray super 4 white ','2025-07-06','2025-07-09','6422944090',1),
	(6,3,2,'MRSU5017091',NULL,'4 white coolray chaoneng-walle',NULL,'2025-07-11','255583362',1),
	(7,3,2,'TCLU5400775',NULL,'4 livan -hsx',NULL,'2025-07-11','255583451',1),
	(8,4,2,'FFAU2255430',NULL,'Walle coolray super auto',NULL,NULL,'6422937590',0),
	(9,4,2,'OOLU9101067',NULL,'walle coolray super auto',NULL,NULL,'6423025190',0),
	(10,4,2,'TIIU5347213',NULL,'Walle coolray super auto',NULL,NULL,'6423025350 ',0),
	(11,4,2,'FCIU9681630',NULL,'Walle coolray super auto',NULL,NULL,'6423025500  ',0),
	(12,4,2,'SEKU4542148',NULL,'walle,nadia,sami',NULL,NULL,'6422683810',0),
	(13,5,2,'CAIU7094512',NULL,'SELTOS\nTIGGO 3X used car\nT-ROC BLACK used car',NULL,NULL,'6424253926',1),
	(14,5,2,'TCLU8624608',NULL,'LIVAN MT',NULL,NULL,'6424253925',1),
	(15,5,2,'FFAU3202754',NULL,'livan auto',NULL,NULL,'6423544901',1),
	(19,6,2,'OOLU9843750',NULL,'emgrand 2 white 2 grey\n',NULL,NULL,'6424603210',0),
	(20,7,2,NULL,NULL,NULL,NULL,NULL,NULL,0),
	(21,8,2,'MSDU7188538',NULL,'4 livan auto','2025-07-17',NULL,NULL,1),
	(22,8,2,'TGHU9855344',NULL,'4 livan auto\n','2025-07-17',NULL,NULL,1),
	(23,8,2,'FCIU7377850',NULL,'4 livan auto','2025-07-17',NULL,NULL,1),
	(24,6,2,'OOCU7143128','2025-07-22','seltos grey black','2025-07-19',NULL,'6424337620',0),
	(25,9,2,'CSNU8581025','2025-07-22','seltos grey','2025-07-19',NULL,'6424601970',0),
	(26,9,2,'OOCU6816693',NULL,'tacqua 4-1','2025-07-20',NULL,'6424602120',0),
	(28,11,3,'CSNU7094969',NULL,'4 tacqua-2','2025-07-20',NULL,'6424269630',0),
	(29,12,2,'FSCU8476604',NULL,'emgrand','2025-07-20',NULL,'6424269920',0),
	(30,12,2,'FFAU3368344',NULL,'coolray','2025-07-20',NULL,'6424269820',0),
	(31,13,2,'FFAU2465458',NULL,'emgrand white/ tacqua','2025-07-20',NULL,'6423898870',0),
	(32,13,2,'OOLU9829392',NULL,'4 tharu grey','2025-07-20',NULL,'6423898540',0),
	(33,14,2,'OOCU6823270',NULL,'4 tharu grey','2025-07-20',NULL,'6424270250',0);

/*!40000 ALTER TABLE `loaded_containers` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table loading
# ------------------------------------------------------------

CREATE TABLE `loading` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `date_loading` date DEFAULT NULL,
  `id_shipping_line` int(11) DEFAULT NULL,
  `freight` decimal(10,0) DEFAULT NULL,
  `id_loading_port` int(11) DEFAULT NULL,
  `id_discharge_port` int(11) DEFAULT NULL,
  `EDD` date DEFAULT NULL,
  `date_loaded` date DEFAULT NULL,
  `note` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=15 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `loading` WRITE;
/*!40000 ALTER TABLE `loading` DISABLE KEYS */;

INSERT INTO `loading` (`id`, `date_loading`, `id_shipping_line`, `freight`, `id_loading_port`, `id_discharge_port`, `EDD`, `date_loaded`, `note`)
VALUES
	(1,'2025-06-25',1,8000,1,3,'2025-08-09','2025-06-30','1 ctns'),
	(2,'2025-06-28',2,5850,2,3,'2025-07-09','2025-07-06','2 ctns coolray super white '),
	(3,'2025-06-28',3,6300,1,3,'2025-08-21','2025-06-30','2 ctns ç¾Žç‰¹è”åˆmason'),
	(4,'2025-07-03',2,6000,2,3,'2025-09-03',NULL,'3 ctns ç¾Žç‰¹è”åˆmason'),
	(5,'2025-07-10',2,5308,1,3,'2025-07-16','2025-07-12','3 ctns ç¾Žç‰¹è”åˆmason\n'),
	(6,'2025-07-20',2,NULL,2,2,NULL,'2025-07-20','2 ctns mason emgrand & seltos'),
	(7,'2025-07-21',2,NULL,2,2,NULL,NULL,'2 ctns t roc '),
	(8,'2025-07-17',4,NULL,1,2,NULL,'2025-07-17','3 ctns mason 12 livan auto'),
	(9,'2025-07-20',2,NULL,2,3,NULL,'2025-07-20','2 ctn seltos grey/ tacqua-mason'),
	(11,'2025-07-20',2,NULL,2,3,NULL,'2025-07-20','1 mason 4 tacqua-2'),
	(12,'2025-07-20',2,NULL,2,2,NULL,'2025-07-20','2 emgrand coolray white-mason'),
	(13,'2025-07-20',2,NULL,2,3,NULL,'2025-07-20','2 emgrand/ tacqua/ tharu grey'),
	(14,'2025-07-20',2,NULL,2,2,NULL,'2025-07-20','4 tharu grey');

/*!40000 ALTER TABLE `loading` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table loading_ports
# ------------------------------------------------------------

CREATE TABLE `loading_ports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `loading_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `loading_port` (`loading_port`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `loading_ports` WRITE;
/*!40000 ALTER TABLE `loading_ports` DISABLE KEYS */;

INSERT INTO `loading_ports` (`id`, `loading_port`)
VALUES
	(1,'NANSHA'),
	(2,'SHANGHAI');

/*!40000 ALTER TABLE `loading_ports` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table permissions
# ------------------------------------------------------------

CREATE TABLE `permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `permission_name` varchar(50) NOT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `permission_name` (`permission_name`)
) ENGINE=InnoDB AUTO_INCREMENT=32 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `permissions` WRITE;
/*!40000 ALTER TABLE `permissions` DISABLE KEYS */;

INSERT INTO `permissions` (`id`, `permission_name`, `description`)
VALUES
	(1,'can_manage_users','Can create, edit, and delete users'),
	(2,'can_manage_roles','Can create, edit, and delete roles'),
	(3,'can_manage_permissions','Can manage role permissions'),
	(4,'is_exchange_sender','Can send exchange requests'),
	(5,'is_exchange_receiver','Can receive exchange requests'),
	(6,'can_manage_cars','Can manage cars inventory'),
	(7,'can_edit_cars_prop',''),
	(8,'can_edit_vin',''),
	(9,'can_edit_car_client_name',''),
	(10,'can_edit_cars_sell_price',''),
	(11,'can_edit_cars_sell_rate',''),
	(12,'can_edit_cars_discharge_port',''),
	(13,'can_upload_car_files',NULL),
	(14,'can_edit_cars_ports',NULL),
	(15,'can_edit_car_money',NULL),
	(16,'can_receive_car',NULL),
	(17,'can_edit_car_documents',NULL),
	(18,'can_load_car',NULL),
	(19,'can_edit_sell_payments',NULL),
	(20,'can_delete_sell_paymets',NULL),
	(21,'can_c_sell_payments',NULL),
	(22,'can_delete_sell_bill',NULL),
	(23,'can_edit_sell_bill',NULL),
	(24,'can_access_cashier',NULL),
	(26,'can_purchase_cars',NULL),
	(27,'can_sell_cars',NULL),
	(28,'can_c_car_stock',NULL),
	(29,'can_assign_to_tmp_clients',NULL),
	(30,'can_change_car_color',NULL),
	(31,'can_c_other_users_sells',NULL);

/*!40000 ALTER TABLE `permissions` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table priorities
# ------------------------------------------------------------

CREATE TABLE `priorities` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `priority` varchar(255) DEFAULT NULL,
  `id_color` int(11) DEFAULT NULL,
  `power` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `priorities` WRITE;
/*!40000 ALTER TABLE `priorities` DISABLE KEYS */;

INSERT INTO `priorities` (`id`, `priority`, `id_color`, `power`)
VALUES
	(1,'very important',NULL,3),
	(2,'urgent',NULL,4),
	(3,'important',NULL,2);

/*!40000 ALTER TABLE `priorities` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table rates
# ------------------------------------------------------------

CREATE TABLE `rates` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `rate` decimal(10,2) DEFAULT NULL,
  `created_on` timestamp NULL DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `rates` WRITE;
/*!40000 ALTER TABLE `rates` DISABLE KEYS */;

INSERT INTO `rates` (`id`, `rate`, `created_on`, `id_user`, `notes`)
VALUES
	(1,250.00,'2025-05-26 10:50:54',1,''),
	(2,242.50,'2025-05-27 14:07:05',3,'');

/*!40000 ALTER TABLE `rates` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table role_permissions
# ------------------------------------------------------------

CREATE TABLE `role_permissions` (
  `role_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`role_id`,`permission_id`),
  KEY `permission_id` (`permission_id`),
  CONSTRAINT `role_permissions_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `role_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `role_permissions` WRITE;
/*!40000 ALTER TABLE `role_permissions` DISABLE KEYS */;

INSERT INTO `role_permissions` (`role_id`, `permission_id`)
VALUES
	(1,1),
	(1,2),
	(1,3),
	(1,4),
	(5,4),
	(6,4),
	(1,5),
	(1,6),
	(6,6),
	(7,6),
	(1,7),
	(1,8),
	(1,9),
	(6,9),
	(1,10),
	(6,10),
	(7,10),
	(1,11),
	(6,11),
	(7,11),
	(1,12),
	(6,12),
	(7,12),
	(1,13),
	(1,14),
	(6,14),
	(7,14),
	(1,15),
	(6,15),
	(7,15),
	(1,16),
	(1,17),
	(1,18),
	(1,19),
	(6,19),
	(1,20),
	(6,20),
	(1,21),
	(6,21),
	(7,21),
	(1,22),
	(1,23),
	(6,23),
	(1,24),
	(6,24),
	(7,24),
	(1,26),
	(1,27),
	(6,27),
	(7,27),
	(1,28),
	(6,28),
	(7,28),
	(1,29),
	(1,30),
	(6,31);

/*!40000 ALTER TABLE `role_permissions` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table roles
# ------------------------------------------------------------

CREATE TABLE `roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role_name` varchar(50) NOT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_name` (`role_name`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `roles` WRITE;
/*!40000 ALTER TABLE `roles` DISABLE KEYS */;

INSERT INTO `roles` (`id`, `role_name`, `description`)
VALUES
	(1,'admin','Administrator with full access'),
	(2,'user','Regular user with limited access'),
	(5,'transfers','do transfers'),
	(6,'ISHAQ','ONLY FOR ISHAQ'),
	(7,'SELLERS','');

/*!40000 ALTER TABLE `roles` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table sell_bill
# ------------------------------------------------------------

CREATE TABLE `sell_bill` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_broker` int(11) DEFAULT NULL,
  `date_sell` date DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `path_pi` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  `is_batch_sell` tinyint(1) DEFAULT 0,
  `time_created` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=193 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `sell_bill` WRITE;
/*!40000 ALTER TABLE `sell_bill` DISABLE KEYS */;

INSERT INTO `sell_bill` (`id`, `id_broker`, `date_sell`, `notes`, `id_user`, `path_pi`, `bill_ref`, `is_batch_sell`, `time_created`)
VALUES
	(2,4,'2025-05-09','date is wrong',4,NULL,'TOU004090525002',0,NULL),
	(3,84,'2025-06-12','',3,NULL,'ISH003120625003',0,NULL),
	(4,86,'2025-06-12','',6,NULL,'MOH006120625004',1,NULL),
	(5,85,'2025-06-12','',3,NULL,'ISH003120625005',0,NULL),
	(6,108,'2025-06-13','',6,NULL,'MOH006130625006',0,NULL),
	(7,128,'2025-06-13','',6,NULL,'MOH006130625007',0,NULL),
	(8,46,'2025-05-23','',6,NULL,'MOH006230525008',0,NULL),
	(9,27,'2025-05-21','',6,NULL,'MOH006210525009',0,NULL),
	(10,132,'2025-05-25','',10,NULL,'MAS010250525010',0,NULL),
	(11,142,'2025-05-21','',3,NULL,'ISH003210525011',0,NULL),
	(12,107,'2025-05-19','',3,NULL,'ISH003190525012',0,NULL),
	(13,177,'2025-05-17','',6,NULL,'MOH006170525013',0,NULL),
	(14,53,'2025-05-14','',6,NULL,'MOH006140525014',0,NULL),
	(15,53,'2025-05-19','',6,NULL,'MOH006190525015',0,NULL),
	(16,125,'2025-05-19','',3,NULL,'ISH003190525016',0,NULL),
	(17,126,'2025-05-19','',3,NULL,'ISH003190525017',0,NULL),
	(18,126,'2025-05-19','',3,NULL,'ISH003190525018',0,NULL),
	(19,126,'2025-05-19','',3,NULL,'ISH003190525019',0,NULL),
	(20,126,'2025-05-19','',3,NULL,'ISH003190525020',0,NULL),
	(21,53,'2025-05-14','',3,NULL,'ISH003140525021',0,NULL),
	(22,127,'2025-05-19','',3,NULL,'ISH003190525022',0,NULL),
	(23,23,'2025-05-10','',3,NULL,'ISH003100525023',0,NULL),
	(24,120,'2025-05-17','',3,NULL,'ISH003170525024',0,NULL),
	(25,152,'2025-05-05','',3,NULL,'ISH003050525025',0,NULL),
	(26,153,'2025-05-19','',3,NULL,'ISH003190525026',0,NULL),
	(27,80,'2025-04-19','',3,NULL,'ISH003190425027',0,NULL),
	(28,162,'2025-05-10','',3,NULL,'ISH003100525028',0,NULL),
	(29,105,'2025-05-16','please recheck passports from whatsapp',6,NULL,'MOH006160525029',0,NULL),
	(30,105,'2025-04-22','',6,NULL,'MOH006220425030',0,NULL),
	(31,168,'2025-04-22','',6,NULL,'MOH006220425031',0,NULL),
	(32,170,'2025-04-22','',6,NULL,'MOH006220425032',0,NULL),
	(33,170,'2025-04-22','',6,NULL,'MOH006220425033',0,NULL),
	(34,68,'2025-05-25','',6,NULL,'MOH006250525034',0,NULL),
	(35,176,'2025-05-14','',6,NULL,'MOH006140525035',0,NULL),
	(36,123,'2025-04-12','',6,NULL,'MOH006120425036',0,NULL),
	(38,177,'2025-05-12','',6,NULL,'MOH006120525038',0,NULL),
	(39,41,'2025-04-21','',3,NULL,'ISH003210425039',0,NULL),
	(40,180,'2025-04-23','',6,NULL,'MOH006230425040',0,NULL),
	(41,45,'2025-04-21','',3,NULL,'ISH003210425041',0,NULL),
	(42,185,'2025-04-19','',6,NULL,'MOH006190425042',0,NULL),
	(43,187,'2025-04-16','',6,NULL,'MOH006160425043',0,NULL),
	(44,187,'2025-04-22','',6,NULL,'MOH006220425044',0,NULL),
	(45,107,'2025-04-28','',6,NULL,'MOH006280425045',0,NULL),
	(46,200,'2025-04-04','',5,NULL,'ATE005040425046',0,NULL),
	(47,201,'2025-04-08','',5,NULL,'ATE005080425047',0,NULL),
	(48,210,'2025-04-06','',5,NULL,'ATE005060425048',0,NULL),
	(49,211,'2025-04-08','',5,NULL,'ATE005080425049',0,NULL),
	(50,212,'2025-04-08','',5,NULL,'ATE005080425050',0,NULL),
	(51,83,'2025-06-14','',3,NULL,'ISH003140625051',0,NULL),
	(52,215,'2025-04-08','',5,NULL,'ATE005080425052',0,NULL),
	(53,220,'2025-04-10','',5,NULL,'ATE005100425053',0,NULL),
	(54,76,'2025-04-14','',5,NULL,'ATE005140425054',0,NULL),
	(55,224,'2025-04-12','',5,NULL,'ATE005120425055',0,NULL),
	(56,73,'2025-04-11','',5,NULL,'ATE005110425056',0,NULL),
	(57,110,'2025-04-11','',5,NULL,'ATE005110425057',0,NULL),
	(58,74,'2025-04-14','',5,NULL,'ATE005140425058',0,NULL),
	(59,12,'2025-04-15','',5,NULL,'ATE005150425059',0,NULL),
	(60,87,'2025-04-16','',5,NULL,'ATE005160425060',0,NULL),
	(61,111,'2025-05-01','',5,NULL,'ATE005010525061',0,NULL),
	(62,109,'2025-05-01','',5,NULL,'ATE005010525062',0,NULL),
	(63,231,'2025-04-16','',5,NULL,'ATE005160425063',0,NULL),
	(64,233,'2025-05-07','',5,NULL,'ATE005070525064',0,NULL),
	(65,235,'2025-05-06','',5,NULL,'ATE005060525065',0,NULL),
	(66,234,'2025-05-19','',5,NULL,'ATE005190525066',0,NULL),
	(67,232,'2025-05-19','',5,NULL,'ATE005190525067',0,NULL),
	(68,229,'2025-05-20','',5,NULL,'ATE005200525068',0,NULL),
	(69,228,'2025-05-20','',5,NULL,'ATE005200525069',0,NULL),
	(70,227,'2025-05-20','',5,NULL,'ATE005200525070',0,NULL),
	(71,218,'2025-05-22','',5,NULL,'ATE005220525071',0,NULL),
	(72,214,'2025-06-22','',5,NULL,'ATE005220625072',0,NULL),
	(73,213,'2025-05-22','',5,NULL,'ATE005220525073',0,NULL),
	(74,236,'2025-05-24','',5,NULL,'ATE005240525074',0,NULL),
	(76,237,'2025-05-23','',5,NULL,'ATE005230525076',0,NULL),
	(77,238,'2025-06-14','',3,NULL,'ISH003140625077',0,NULL),
	(78,239,'2025-06-15','',3,NULL,'ISH003150625078',0,NULL),
	(79,240,'2025-06-16','',3,NULL,'ISH003160625079',0,NULL),
	(80,86,'2025-06-17','',1,NULL,'ADM001170625080',0,NULL),
	(81,242,'2025-04-15','',1,NULL,'ADM001150425081',0,NULL),
	(82,244,'2025-05-14','',1,NULL,'ADM001140525082',0,NULL),
	(83,245,'2025-06-19','',3,NULL,'ISH003190625083',0,NULL),
	(84,246,'2025-06-19','',4,NULL,'TOU004190625084',0,NULL),
	(85,248,'2025-06-20','',3,NULL,'ISH003200625085',0,NULL),
	(86,249,'2025-06-20','',3,NULL,'ISH003200625086',0,NULL),
	(87,72,'2025-05-07','',1,NULL,'ADM001070525087',1,NULL),
	(88,72,'2025-05-08','',1,NULL,'ADM001080525088',1,NULL),
	(89,72,'2025-05-04','',1,NULL,'ADM001040525089',1,NULL),
	(90,72,'2025-05-04','',1,NULL,'ADM001040525090',1,NULL),
	(91,72,'2025-05-04','',1,NULL,'ADM001040525091',1,NULL),
	(92,72,'2025-05-04','',1,NULL,'ADM001040525092',1,NULL),
	(93,72,'2025-05-05','',1,NULL,'ADM001050525093',1,NULL),
	(94,72,'2025-04-01','from lio',1,NULL,'ADM001010425094',1,NULL),
	(95,72,'2025-03-07','',1,NULL,'ADM001070325095',1,NULL),
	(96,72,'2025-04-29','',1,NULL,'ADM001290425096',1,NULL),
	(97,72,'2025-04-09','',1,NULL,'ADM001090425097',1,NULL),
	(98,72,'2025-03-20','',1,NULL,'ADM001200325098',1,NULL),
	(99,86,'2025-03-20','',1,NULL,'ADM001200325099',1,NULL),
	(100,72,'2025-03-27','',1,NULL,'ADM001270325100',1,NULL),
	(101,72,'2025-03-22','',1,NULL,'ADM001220325101',1,NULL),
	(102,72,'2025-04-03','',1,NULL,'ADM001030425102',1,NULL),
	(103,86,'2025-03-18','',1,NULL,'ADM001180325103',1,NULL),
	(104,72,'2025-03-09','',1,NULL,'ADM001090325104',1,NULL),
	(105,72,'2025-03-18','',1,NULL,'ADM001180325105',1,NULL),
	(106,250,'2025-06-22','',4,NULL,'TOU004220625106',0,NULL),
	(112,258,'2025-06-22','',3,NULL,'ISH003220625112',0,NULL),
	(114,270,'2025-06-23','',3,NULL,'ISH003230625114',0,NULL),
	(115,254,'2025-04-23','',6,NULL,'ADM001230425115',0,NULL),
	(116,251,'2025-06-12','',6,NULL,'ADM001120625116',0,NULL),
	(117,275,'2025-06-24','',3,NULL,'ISH003240625117',0,NULL),
	(118,276,'2025-06-24','',3,NULL,'ISH003240625118',0,NULL),
	(119,277,'2025-04-25','',6,NULL,'ADM001250425119',0,NULL),
	(121,282,'2025-06-26','',4,NULL,'ADM001260625121',0,NULL),
	(122,4,'2025-06-26','',4,NULL,'ADM001260625122',0,NULL),
	(123,83,'2025-05-28','',12,NULL,'ADM001280525123',0,NULL),
	(124,86,'2025-05-29','',6,NULL,'ADM001290525124',1,NULL),
	(125,4,'2025-05-29','',6,NULL,'ADM001290525125',0,NULL),
	(126,83,'2025-06-30','',3,NULL,'ISH003300625126',0,NULL),
	(127,286,'2025-04-15','',3,NULL,'ISH003150425127',0,NULL),
	(128,287,'2025-04-12','',3,NULL,'ISH003120425128',0,NULL),
	(129,88,'2025-04-21','',6,NULL,'ISH003210425129',0,NULL),
	(130,81,'2025-04-16','',3,NULL,'ISH003160425130',0,NULL),
	(131,292,'2025-03-25','',3,NULL,'ISH003250325131',0,NULL),
	(132,67,'2025-07-01','temporary ',1,NULL,'ADM001010725132',0,NULL),
	(133,107,'2025-07-01','',3,NULL,'ISH003010725133',0,NULL),
	(134,294,'2025-03-13','this one his LIVAN was from bella\nbut we return back to lamine \ni don\'t know from which supplier  ',1,NULL,'ISH003010725134',0,NULL),
	(135,295,'2025-03-20','',6,NULL,'ADM001200325135',0,NULL),
	(137,298,'2025-01-14','TOO OLD FORGOTTEN ORDER ðŸ˜­',1,NULL,'ADM001140125137',0,NULL),
	(138,299,'2025-07-03','',3,NULL,'ISH003030725138',0,NULL),
	(139,300,'2025-07-04','',3,NULL,'ISH003030725139',0,NULL),
	(140,296,'2025-07-05','',9,NULL,'ALI009050725140',0,NULL),
	(141,301,'2025-07-05','',3,NULL,'ISH003050725141',0,NULL),
	(142,NULL,'2025-07-08','AmÃ© Brahim',4,NULL,'TOU004080725142',1,NULL),
	(143,NULL,'2025-07-08','',2,NULL,'SOF002080725143',0,NULL),
	(144,302,'2025-07-10','THISN GUY BOUGHT LOADED CAR',4,NULL,'MOH006100725144',0,NULL),
	(146,306,'2025-07-10','',3,NULL,'MOH006100725146',0,NULL),
	(147,307,'2025-07-10','',6,NULL,'MOH006100725147',0,NULL),
	(148,308,'2025-07-10','',6,NULL,'MOH006100725148',0,NULL),
	(149,309,'2025-07-10','',6,NULL,'MOH006100725149',0,NULL),
	(150,310,'2025-07-10','',6,NULL,'MOH006100725150',0,NULL),
	(151,311,'2025-07-10','',6,NULL,'MOH006100725151',0,NULL),
	(152,316,'2025-07-10','',3,NULL,'MOH006100725152',0,NULL),
	(153,54,'2025-07-10','',5,NULL,'MOH006100725153',0,NULL),
	(155,320,'2025-07-14','',3,NULL,'ISH003140725155',0,NULL),
	(156,61,'2025-07-14','',3,NULL,'ISH003140725156',0,NULL),
	(157,57,'2025-07-14','',5,NULL,'SOF002140725157',0,NULL),
	(160,4,'2025-05-26','check if they finish payment ',9,NULL,'SOF002260525160',0,NULL),
	(161,333,'2025-07-15','',3,NULL,'ISH003150725161',0,NULL),
	(162,57,'2025-06-18','',5,NULL,'ATE005180625162',0,NULL),
	(163,334,'2025-07-15','just to fix stock',1,NULL,'ADM001150725163',1,'2025-07-15 15:26:48'),
	(164,336,'2025-07-16','',6,NULL,'MOH006160725164',0,'2025-07-16 04:57:20'),
	(165,337,'2025-07-16','',6,NULL,'MOH006160725165',0,'2025-07-16 05:10:22'),
	(166,338,'2025-07-16','',3,NULL,'ISH003160725166',0,'2025-07-16 05:20:39'),
	(168,339,'2025-07-16','',6,NULL,'MOH006160725168',0,'2025-07-16 05:42:07'),
	(169,346,'2025-07-16','',6,NULL,'MOH006160725169',0,'2025-07-16 05:48:10'),
	(170,347,'2025-07-16','',6,NULL,'MOH006160725170',0,'2025-07-16 05:51:06'),
	(171,348,'2025-07-16','',NULL,NULL,'MOH006160725171',0,'2025-07-16 05:56:30'),
	(172,349,'2025-07-16','',6,NULL,'MOH006160725172',0,'2025-07-16 06:00:14'),
	(173,350,'2025-07-16','',6,NULL,'MOH006160725173',0,'2025-07-16 06:02:50'),
	(174,351,'2025-07-16','',6,NULL,'MOH006160725174',0,'2025-07-16 06:23:23'),
	(175,352,'2025-07-16','',3,NULL,'ISH003160725175',0,'2025-07-16 07:15:59'),
	(176,57,'2025-07-17','',5,NULL,'ATE005170725176',0,NULL),
	(177,353,'2025-07-17','',5,NULL,'MOH006170725177',0,'2025-07-17 05:16:23'),
	(178,57,'2025-07-17','',5,NULL,'ATE005170725178',0,NULL),
	(179,57,'2025-07-17','',5,NULL,'ATE005170725179',0,NULL),
	(180,57,'2025-06-21','',5,NULL,'ATE005210625180',0,NULL),
	(181,57,'2025-06-21','',5,NULL,'ATE005210625181',0,NULL),
	(183,336,'2025-07-18','',6,NULL,'ADM001180725183',0,'2025-07-18 11:29:15'),
	(184,360,'2025-07-19','',3,NULL,'ISH003190725184',0,'2025-07-19 04:17:04'),
	(185,336,'2025-07-19','',6,NULL,'ADM001190725185',0,'2025-07-19 14:09:52'),
	(186,364,'2025-07-20','',5,NULL,'ADM001200725186',0,'2025-07-20 03:02:30'),
	(187,365,'2025-07-20','',4,NULL,'ADM001200725187',0,'2025-07-20 11:10:46'),
	(188,366,'2025-07-20','',6,NULL,'ADM001200725188',0,'2025-07-20 11:57:43'),
	(189,228,'2025-07-20','',5,NULL,'ATE005200725189',0,'2025-07-20 15:48:18'),
	(190,375,'2025-07-21','',6,NULL,'MOH006210725190',0,'2025-07-21 01:36:25'),
	(191,107,'2025-07-21','',3,NULL,'ISH003210725191',0,'2025-07-21 05:14:48'),
	(192,382,'2025-07-21','',3,NULL,'ISH003210725192',0,'2025-07-21 12:50:17');

/*!40000 ALTER TABLE `sell_bill` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table sell_payments
# ------------------------------------------------------------

CREATE TABLE `sell_payments` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_sell_bill` int(11) DEFAULT NULL,
  `amount_usd` decimal(10,2) DEFAULT NULL,
  `amount_da` decimal(10,2) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `date` date DEFAULT NULL,
  `path_swift` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=107 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `sell_payments` WRITE;
/*!40000 ALTER TABLE `sell_payments` DISABLE KEYS */;

INSERT INTO `sell_payments` (`id`, `id_sell_bill`, `amount_usd`, `amount_da`, `rate`, `date`, `path_swift`, `id_user`, `notes`)
VALUES
	(1,7,7700.00,1925000.00,250.00,'2025-04-24','payments_swift/swift_payment_7_2025-06-13_09-45-10_12800USD_1749807910950.jpg',6,'rate not exact\nswift i dont have'),
	(2,6,12800.00,3200000.00,250.00,'2025-06-13','payments_swift/swift_payment_6_2025-06-13_09-47-38_12800USD_1749808058495.jpg',6,'rate is not correct\nI dont have swift'),
	(3,8,49274.00,12318500.00,250.00,'2025-05-31','payments_swift/swift_payment_8_2025-06-13_10-01-52_49274.00USD_1749808912989.jpg',6,'rate not correct'),
	(4,8,14340.00,3585000.00,250.00,'2025-05-24','payments_swift/swift_payment_8_2025-06-13_10-03-05_1434.00USD_1749808985156.jpg',6,'rate not correct'),
	(5,9,15800.00,3950000.00,250.00,'2025-05-28','payments_swift/swift_payment_9_2025-06-13_10-10-38_15800.00USD_1749809438969.jpg',6,'rate not correct'),
	(6,10,159000.00,39750000.00,250.00,'2025-05-21','payments_swift/swift_payment_10_2025-06-13_10-46-38_159000USD_1749811598568.jpg',10,'rate not correct \nno swift'),
	(7,11,10600.00,2650000.00,250.00,'2025-05-22','payments_swift/swift_payment_11_2025-06-13_10-57-14_10600USD_1749812234235.jpg',3,''),
	(8,12,20285.00,5071250.00,250.00,'2025-05-30','payments_swift/swift_payment_12_2025-06-13_11-13-05_20285USD_1749813185400.jpg',3,'original he sent 20300'),
	(9,12,15815.00,3953750.00,250.00,'2025-06-13','payments_swift/swift_payment_12_2025-06-13_11-15-10_15815USD_1749813310464.jpg',3,'no swift payed da ishaq'),
	(10,13,10400.00,2600000.00,250.00,'2025-05-21','payments_swift/swift_payment_13_2025-06-13_15-46-59_10400USD_1749829619387.jpg',6,''),
	(11,14,800.00,200000.00,250.00,'2025-05-14','payments_swift/swift_payment_14_2025-06-13_15-51-48_800.00USD_1749829908374.jpg',6,''),
	(12,14,8400.00,2100000.00,250.00,'2025-05-20','payments_swift/swift_payment_14_2025-06-13_15-52-21_8400.00USD_1749829941157.jpg',6,''),
	(13,15,31200.00,7800000.00,250.00,'2025-05-19','payments_swift/swift_payment_15_2025-06-13_15-58-32_312000USD_1749830312232.jpg',6,''),
	(14,16,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_16_2025-06-13_16-05-32_10600USD_1749830732905.jpg',3,''),
	(15,17,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_17_2025-06-13_16-10-27_10600USD_1749831027952.jpg',3,''),
	(16,18,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_18_2025-06-13_16-16-33_10600USD_1749831393513.jpg',3,''),
	(17,19,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_19_2025-06-13_16-23-24_10600USD_1749831804755.jpg',3,''),
	(18,20,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_20_2025-06-13_16-29-07_10600USD_1749832147390.jpg',3,''),
	(19,21,1600.00,400000.00,250.00,'2025-05-14','payments_swift/swift_payment_21_2025-06-13_16-38-58_1600.00USD_1749832738887.jpg',3,''),
	(20,21,7600.00,1900000.00,250.00,'2025-05-25','payments_swift/swift_payment_21_2025-06-13_16-39-25_1900000USD_1749832765652.jpg',3,''),
	(21,22,10600.00,2650000.00,250.00,'2025-05-19','payments_swift/swift_payment_22_2025-06-13_16-46-31_10600USD_1749833191758.jpg',3,''),
	(22,23,17200.00,4300000.00,250.00,'2025-06-13','payments_swift/swift_payment_23_2025-06-13_16-51-39_17200USD_1749833499566.pdf',3,'no swift \nhe payed 15560 euro'),
	(23,24,12700.00,3175000.00,250.00,'2025-05-17','payments_swift/swift_payment_24_2025-06-13_17-18-29_12700USD_1749835109285.pdf',3,'no swift'),
	(24,25,14800.00,3700000.00,250.00,'2025-05-05','payments_swift/swift_payment_25_2025-06-13_17-35-39_14800USD_1749836139450.pdf',3,'no swift'),
	(25,26,137920.00,34480000.00,250.00,'2025-05-19','payments_swift/swift_payment_26_2025-06-13_18-22-25_137920USD_1749838945535.jpg',3,'no swift'),
	(26,27,9800.00,2450000.00,250.00,'2025-04-19','payments_swift/swift_payment_27_2025-06-13_18-29-19_9800USD_1749839359438.jpg',3,''),
	(27,28,11100.00,2775000.00,250.00,'2025-05-15','payments_swift/swift_payment_28_2025-06-13_18-37-29_11100.00USD_1749839849272.jpg',3,''),
	(28,28,400.00,100000.00,250.00,'2025-05-15','payments_swift/swift_payment_28_2025-06-13_18-38-11_400USD_1749839891367.pdf',3,'no swift \nto my ccp'),
	(29,29,42400.00,10600000.00,250.00,'2025-05-16','payments_swift/swift_payment_29_2025-06-13_19-06-48_42400.00USD_1749841608372.jpg',6,'payed cash in belabbes \nno swift'),
	(30,30,46200.00,11550000.00,250.00,'2025-04-22','payments_swift/swift_payment_30_2025-06-13_19-48-25_46200USD_1749844105386.jpg',6,'no swift'),
	(31,31,20000.00,5000000.00,250.00,'2025-04-22','payments_swift/swift_payment_31_2025-06-13_19-55-12_20000USD_1749844512064.jpg',6,'no swift'),
	(32,32,20000.00,5000000.00,250.00,'2025-04-22','payments_swift/swift_payment_32_2025-06-13_20-19-37_20000USD_1749845977399.jpg',6,'no swift'),
	(33,33,52400.00,13100000.00,250.00,'2025-04-28','payments_swift/swift_payment_33_2025-06-13_20-38-44_52400.00USD_1749847124326.jpg',6,''),
	(34,34,10000.00,2500000.00,250.00,'2025-05-14','payments_swift/swift_payment_34_2025-06-13_20-44-09_10000.00USD_1749847449741.jpg',6,''),
	(35,35,9200.00,2300000.00,250.00,'2025-05-14','payments_swift/swift_payment_35_2025-06-13_20-48-51_9200.00USD_1749847731398.jpg',6,''),
	(36,36,12900.00,3225000.00,250.00,'2025-04-12','payments_swift/swift_payment_36_2025-06-13_21-04-39_12900USD_1749848679340.pdf',6,'no swift'),
	(37,38,47880.00,11970000.00,250.00,'2025-05-12','payments_swift/swift_payment_38_2025-06-13_21-42-55_47880.00USD_1749850975087.jpg',6,''),
	(38,40,69440.00,17360000.00,250.00,'2025-04-19','payments_swift/swift_payment_40_2025-06-13_21-56-48_69440USD_1749851808436.jpg',6,'no swift'),
	(39,41,10000.00,2500000.00,250.00,'2025-06-13','payments_swift/swift_payment_41_2025-06-13_21-59-35_10000USD_1749851975865.jpg',3,'no swift'),
	(40,42,30000.00,7500000.00,250.00,'2025-04-19','payments_swift/swift_payment_42_2025-06-14_07-07-09_30000USD_1749884829721.pdf',6,''),
	(41,43,35864.00,8966000.00,250.00,'2025-04-22','payments_swift/swift_payment_43_2025-06-14_07-34-40_35864USD_1749886480528.pdf',6,'no swift'),
	(42,44,10000.00,2500000.00,250.00,'2025-04-22','payments_swift/swift_payment_44_2025-06-14_07-38-56_10000USD_1749886736832.pdf',6,''),
	(43,46,10000.00,2500000.00,250.00,'2025-04-04','payments_swift/swift_payment_46_2025-06-14_12-58-09_10000USD_1749905889709.pdf',5,''),
	(44,47,70000.00,17500000.00,250.00,'2025-04-08','payments_swift/swift_payment_47_2025-06-14_13-28-09_700000.00USD_1749907689726.jpg',5,''),
	(46,48,12000.00,3000000.00,250.00,'2025-04-06','payments_swift/swift_payment_48_2025-06-14_13-40-53_12000USD_1749908453398.jpg',5,''),
	(47,49,10000.00,2500000.00,250.00,'2025-04-08','payments_swift/swift_payment_49_2025-06-14_13-46-34_10000USD_1749908794782.jpg',5,''),
	(48,50,13000.00,3250000.00,250.00,'2025-04-08','payments_swift/swift_payment_50_2025-06-14_14-02-48_13000USD_1749909768598.jpg',5,''),
	(49,52,9600.00,2400000.00,250.00,'2025-04-08','payments_swift/swift_payment_52_2025-06-14_14-15-29_9600USD_1749910529025.jpg',5,''),
	(50,53,10000.00,2500000.00,250.00,'2025-04-10','payments_swift/swift_payment_53_2025-06-14_14-19-42_10000USD_1749910782802.jpg',5,''),
	(51,54,9280.00,2320000.00,250.00,'2025-04-14','payments_swift/swift_payment_54_2025-06-14_14-25-51_9280USD_1749911151560.jpg',5,''),
	(52,55,9200.00,2300000.00,250.00,'2025-04-15','payments_swift/swift_payment_55_2025-06-14_14-32-22_9200USD_1749911542705.jpg',5,''),
	(53,56,9000.00,2250000.00,250.00,'2025-04-11','payments_swift/swift_payment_56_2025-06-14_14-42-00_9000USD_1749912120395.jpg',5,''),
	(54,57,9200.00,2300000.00,250.00,'2025-04-17','payments_swift/swift_payment_57_2025-06-14_14-48-29_9200USD_1749912509376.jpg',5,''),
	(55,58,9280.00,2320000.00,250.00,'2025-04-14','payments_swift/swift_payment_58_2025-06-14_14-56-09_9280USD_1749912969013.jpg',5,''),
	(56,59,12320.00,3080000.00,250.00,'2025-04-19','payments_swift/swift_payment_59_2025-06-14_15-00-29_12320USD_1749913229946.jpg',5,''),
	(57,60,10000.00,2500000.00,250.00,'2025-04-16','payments_swift/swift_payment_60_2025-06-14_15-04-08_10000USD_1749913448162.jpg',5,''),
	(58,61,9400.00,2350000.00,250.00,'2025-05-01','payments_swift/swift_payment_61_2025-06-14_15-08-09_9400USD_1749913689629.jpg',5,''),
	(59,62,16000.00,4000000.00,250.00,'2025-05-01','payments_swift/swift_payment_62_2025-06-14_15-13-06_16000USD_1749913986461.jpg',5,''),
	(60,63,9400.00,2350000.00,250.00,'2025-05-01','payments_swift/swift_payment_63_2025-06-14_15-16-59_9400USD_1749914219559.jpg',5,''),
	(61,64,9400.00,2350000.00,250.00,'2025-05-07','payments_swift/swift_payment_64_2025-06-14_15-20-48_9400USD_1749914448390.jpg',5,''),
	(62,65,11400.00,2850000.00,250.00,'2025-05-06','payments_swift/swift_payment_65_2025-06-14_15-25-39_11400USD_1749914739403.jpg',5,''),
	(63,66,20400.00,5100000.00,250.00,'2025-05-17','payments_swift/swift_payment_66_2025-06-14_15-28-18_20400USD_1749914898036.jpg',5,''),
	(64,67,10600.00,2650000.00,250.00,'2025-05-18','payments_swift/swift_payment_67_2025-06-14_15-30-46_10600USD_1749915046318.jpg',5,''),
	(65,68,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_68_2025-06-14_15-32-33_10600USD_1749915153566.jpg',5,''),
	(66,69,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_69_2025-06-14_15-34-17_10600USD_1749915257598.jpg',5,''),
	(67,70,10600.00,2650000.00,250.00,'2025-05-20','payments_swift/swift_payment_70_2025-06-14_15-35-59_10600USD_1749915359598.jpg',5,''),
	(68,71,2080.00,520000.00,250.00,'2025-05-22','payments_swift/swift_payment_71_2025-06-14_15-45-35_2080USD_1749915935049.jpg',5,''),
	(69,51,10400.00,2600000.00,250.00,'2025-06-14','payments_swift/swift_payment_51_2025-06-14_15-46-09_10400.00USD_1749915969221.jpeg',3,''),
	(70,72,17300.00,4325000.00,250.00,'2025-05-22','payments_swift/swift_payment_72_2025-06-14_15-54-37_17300USD_1749916477288.jpg',5,''),
	(71,73,17300.00,4325000.00,250.00,'2025-05-22','payments_swift/swift_payment_73_2025-06-14_15-56-51_17300USD_1749916611902.jpg',5,''),
	(72,76,21200.00,5300000.00,250.00,'2025-05-23','payments_swift/swift_payment_76_2025-06-14_16-51-26_21200USD_1749919886678.jpg',5,''),
	(73,77,14800.00,3700000.00,250.00,'2025-06-14','payments_swift/swift_payment_77_2025-06-14_20-35-32_14800.00USD_1749933332034.jpeg',3,''),
	(74,78,14800.00,3700000.00,250.00,'2025-06-15','payments_swift/swift_payment_78_2025-06-15_13-53-40_14800.00USD_1749995620625.jpeg',3,''),
	(75,51,38640.00,9660000.00,250.00,'2025-06-15','payments_swift/swift_payment_51_2025-06-15_21-59-22_38640.00USD_1750024762025.pdf',3,''),
	(76,79,29600.00,7400000.00,250.00,'2025-06-16','payments_swift/swift_payment_79_2025-06-16_10-08-24_29600.00USD_1750068504950.jpeg',3,''),
	(77,81,18000.00,4500000.00,250.00,'2025-04-14','payments_swift/swift_payment_81_2025-06-17_14-35-25_18000USD_1750170925603.jpg',1,'no swift'),
	(78,82,9200.00,2300000.00,250.00,'2025-05-14','payments_swift/swift_payment_82_2025-06-17_14-44-32_9200USD_1750171472570.jpg',1,'payed euro cash'),
	(79,84,10400.00,2600000.00,250.00,'2025-06-19','payments_swift/swift_payment_84_2025-06-19_17-23-37_10400.00USD_1750353817564.pdf',4,'no swift yet'),
	(80,83,21200.00,5300000.00,250.00,'2025-06-19','payments_swift/swift_payment_83_2025-06-19_17-44-24_21200.00USD_1750355064908.jpeg',3,''),
	(81,5,15200.00,3800000.00,250.00,'2025-06-12','payments_swift/swift_payment_5_2025-06-19_17-51-00_15200.00USD_1750355460646.jpeg',3,''),
	(82,3,15400.00,3850000.00,250.00,'2025-06-12','payments_swift/swift_payment_3_2025-06-19_17-51-30_15400.00USD_1750355490854.jpeg',3,''),
	(83,85,10400.00,2600000.00,250.00,'2025-06-20','payments_swift/swift_payment_85_2025-06-20_09-35-32_10400.00USD_1750412132794.jpeg',3,''),
	(84,112,37200.00,9300000.00,250.00,'2025-06-22','payments_swift/swift_payment_112_2025-06-22_19-01-07_37200.00USD_1750618867446.pdf',1,''),
	(85,114,10400.00,2600000.00,250.00,'2025-06-23','payments_swift/swift_payment_114_2025-06-23_10-54-58_10400.00USD_1750676098142.jpeg',3,''),
	(86,119,47185.00,11796250.00,250.00,'2025-05-05','payments_swift/swift_payment_119_2025-06-25_13-43-20_47185USD_1750859000278.jpg',1,''),
	(87,118,20600.00,5150000.00,250.00,'2025-06-28','payments_swift/swift_payment_118_2025-06-28_09-06-30_20600USD_1751101590687.pdf',3,''),
	(88,117,10400.00,2600000.00,250.00,'2025-06-28','payments_swift/swift_payment_117_2025-06-28_09-08-54_10400.00USD_1751101734499.jpeg',3,'paid by BANK BADR'),
	(89,127,9800.00,2450000.00,250.00,'2025-04-10','payments_swift/swift_payment_127_2025-06-30_18-29-26_9800.00USD_1751308166887.pdf',3,''),
	(90,128,20000.00,5000000.00,250.00,'2025-06-30','payments_swift/swift_payment_128_2025-06-30_18-31-41_20000.00USD_1751308301386.pdf',3,''),
	(91,129,15200.00,3800000.00,250.00,'2025-06-30','payments_swift/swift_payment_129_2025-06-30_18-37-13_15200.00USD_1751308633784.jpeg',3,''),
	(92,130,14960.00,3740000.00,250.00,'2025-06-30','payments_swift/swift_payment_130_2025-06-30_18-39-26_14960.00USD_1751308766975.jpeg',3,''),
	(93,131,19000.00,4750000.00,250.00,'2025-02-20','payments_swift/swift_payment_131_2025-07-01_15-48-30_19000USD_1751384910235.png',3,''),
	(94,141,9892.00,2473000.00,250.00,'2025-07-08','payments_swift/swift_payment_141_2025-07-08_16-44-22_9892.00USD_1751993062965.jpg',3,''),
	(95,138,9892.00,2473000.00,250.00,'2025-07-08','payments_swift/swift_payment_138_2025-07-08_16-46-56_9892.00USD_1751993216023.jpg',3,''),
	(96,139,12720.00,3180000.00,250.00,'2025-07-10','payments_swift/swift_payment_139_2025-07-10_12-03-52_12720USD_1752149032304.jpeg',3,''),
	(97,133,8200.00,2050000.00,250.00,'2025-07-12','payments_swift/swift_payment_133_2025-07-12_16-54-30_10200USD_1752339270213.jpeg',3,''),
	(98,133,2000.00,500000.00,250.00,'2025-07-12','payments_swift/swift_payment_133_2025-07-12_16-57-16_2000.00USD_1752339436081.jpeg',3,''),
	(99,155,8000.00,2000000.00,250.00,'2025-07-14','payments_swift/swift_payment_155_2025-07-15_08-34-06_8000.00USD_1752568446152.jpeg',3,''),
	(100,152,10400.00,2600000.00,250.00,'2025-07-15','payments_swift/swift_payment_152_2025-07-15_08-34-37_10400.00USD_1752568477750.jpeg',3,''),
	(101,146,10200.00,2550000.00,250.00,'2025-07-15','payments_swift/swift_payment_146_2025-07-15_10-05-47_10200.00USD_1752573947258.jpeg',3,'REMOVED 95000.00DZD FROM CHANGING TROC TO THARU ADDED TO THIS PAIMENT HE ADD ONLY 1600000.00DZD'),
	(102,175,11600.00,2900000.00,250.00,'2025-07-16','payments_swift/swift_payment_175_2025-07-17_20-23-15_11600.00USD_1752783795268.jpeg',3,''),
	(103,166,10400.00,2600000.00,250.00,'2025-07-16','payments_swift/swift_payment_166_2025-07-16_11-21-15_10400.00USD_1752664875666.jpeg',3,''),
	(104,184,10000.00,2500000.00,250.00,'2025-07-19','payments_swift/swift_payment_184_2025-07-19_08-35-45_2500000USD_1752914145057.jpeg',3,''),
	(105,161,13200.00,3300000.00,250.00,'2025-07-21','payments_swift/swift_payment_161_2025-07-21_08-16-23_13200.00USD_1753085783299.jpeg',3,''),
	(106,155,2400.00,600000.00,250.00,'2025-07-21','payments_swift/swift_payment_155_2025-07-21_08-19-06_2400.00USD_1753085946783.jpeg',3,'');

/*!40000 ALTER TABLE `sell_payments` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table shipping_lines
# ------------------------------------------------------------

CREATE TABLE `shipping_lines` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `shipping_lines` WRITE;
/*!40000 ALTER TABLE `shipping_lines` DISABLE KEYS */;

INSERT INTO `shipping_lines` (`id`, `name`)
VALUES
	(1,'AKK'),
	(2,'cosco'),
	(4,'msc'),
	(3,'msk');

/*!40000 ALTER TABLE `shipping_lines` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table suppliers
# ------------------------------------------------------------

CREATE TABLE `suppliers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `contact_info` text DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `supplier_name_unic` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `suppliers` WRITE;
/*!40000 ALTER TABLE `suppliers` DISABLE KEYS */;

INSERT INTO `suppliers` (`id`, `name`, `contact_info`, `notes`)
VALUES
	(1,'zhengzhou walle ','',''),
	(2,'MUMU Henan rongwei international trade co.,ltd','+8613333818028\nhigh tech industrial development zone, zhengzhou, china',''),
	(3,'MESSOUD KX1','','messoud friend of Lamine'),
	(4,'Foshan Hui Shunxing Automobile Trading Co., Ltd','','our nei'),
	(5,'LEO','',''),
	(6,'ALI CHINESE','',''),
	(7,'bella','',''),
	(8,'ChongQin huanyu','','kenken'),
	(9,'Weifang Century Sovereign Automobile Sales Co., Ltd','','');

/*!40000 ALTER TABLE `suppliers` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table tasks
# ------------------------------------------------------------

CREATE TABLE `tasks` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_user_create` int(11) DEFAULT NULL,
  `id_user_receive` int(11) DEFAULT NULL,
  `date_create` datetime DEFAULT NULL,
  `date_declare_done` datetime DEFAULT NULL,
  `date_confirm_done` datetime DEFAULT NULL,
  `id_user_confirm_done` int(11) DEFAULT NULL,
  `id_priority` int(11) DEFAULT NULL,
  `title` varchar(255) DEFAULT NULL,
  `desciption` text DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `is_task_for_car` tinyint(1) DEFAULT 0,
  `is_task_for_transfer` tinyint(1) DEFAULT 0,
  `is_task_for_supplier` tinyint(1) DEFAULT 0,
  `is_task_for_client` tinyint(1) DEFAULT 0,
  `is_task_for_user` tinyint(1) DEFAULT 0,
  `id_chat_grroup` int(11) DEFAULT NULL,
  `assigned_users_ids` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`assigned_users_ids`)),
  `subject_ids` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`subject_ids`)),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=51 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `tasks` WRITE;
/*!40000 ALTER TABLE `tasks` DISABLE KEYS */;

INSERT INTO `tasks` (`id`, `id_user_create`, `id_user_receive`, `date_create`, `date_declare_done`, `date_confirm_done`, `id_user_confirm_done`, `id_priority`, `title`, `desciption`, `notes`, `is_task_for_car`, `is_task_for_transfer`, `is_task_for_supplier`, `is_task_for_client`, `is_task_for_user`, `id_chat_grroup`, `assigned_users_ids`, `subject_ids`)
VALUES
	(2,1,2,'2025-06-29 10:54:02','2025-06-29 17:22:34','2025-06-30 09:40:40',1,2,'[Car #281 - COOLRAY BASE AUTO] BL name wrong','the name of coassignee in thhe BL is BABAAMER DJELMAMMO HAMMED should be BABAAMER DJELMAM MOHAMMED','',1,0,0,0,0,NULL,NULL,NULL),
	(3,1,2,'2025-06-29 11:06:34','2025-06-29 17:22:53','2025-06-30 09:40:35',1,1,'[Car #282 - COOLRAY BASE AUTO] wrong client name in the BL','in the BL is HAMMADI RIDA correct is HAMMADI RIAD','',1,0,0,0,0,NULL,NULL,NULL),
	(4,1,2,'2025-06-29 11:16:34','2025-06-29 17:23:06','2025-06-30 09:40:28',1,1,'[Car #277 - COOLRAY BASE AUTO] client name wrong','this id no and car are for SENIGRA ABDELDJALIL but they make the name AMROUNI DJAMILA','',1,0,0,0,0,NULL,NULL,NULL),
	(5,2,1,'2025-06-29 11:45:30','2025-06-29 12:04:23','2025-06-29 12:04:29',1,1,'pay mumu balance','balance : 103600 usd','send me swift after you pay.',0,0,0,0,0,NULL,NULL,NULL),
	(6,2,1,'2025-06-29 11:47:17','2025-06-29 12:03:52','2025-06-29 12:04:03',1,1,'[Supplier #2 - MUMU Henan rongwei international trade co.,ltd] pay mumu balance 103600 usd','','20 livan 7400*20',0,0,0,0,0,NULL,NULL,NULL),
	(7,1,2,'2025-06-29 12:10:39','2025-07-08 11:38:11','2025-07-09 12:25:19',NULL,2,'[Client #281 - CHEGGA OUAIL] this guy took the peugeot 2008 no 5','this is a client of Islam','',0,0,0,0,0,NULL,NULL,NULL),
	(8,1,3,'2025-06-29 12:12:56',NULL,NULL,NULL,2,'[Sell Bill #119 - ADM001250425119] this order is not full payed','this is islam from kasantina he did not payy all his orders please make him pay','',0,0,0,0,0,NULL,NULL,NULL),
	(9,1,3,'2025-06-29 12:14:56',NULL,NULL,NULL,1,'[Sell Bill #124 - ADM001290525124] walid did not pay any of this order','please push hiom to pay urgently','',0,0,0,0,0,NULL,NULL,NULL),
	(10,1,3,'2025-06-29 12:17:39',NULL,NULL,NULL,2,'[Sell Bill #4 - MOH006120625004] did walid finished payment of this order','please check if he finish otherwise push him pay us our money','',0,0,0,0,0,NULL,NULL,NULL),
	(11,1,3,'2025-06-29 12:18:35','2025-06-30 18:40:00','2025-06-30 22:29:23',1,1,'please fill in your old orders','','',0,0,0,0,0,NULL,NULL,NULL),
	(12,1,3,'2025-06-29 12:19:07','2025-07-20 10:13:13',NULL,NULL,2,'orders of kamel midat are not all filled in.','','',0,0,0,0,0,NULL,NULL,NULL),
	(13,2,9,'2025-06-29 17:02:21',NULL,NULL,NULL,3,'[Client #215 - BEN HEFFAF ABDELKARIM A] BL name is wrong','this client name is BEN HEFFAF ABDELKARIM A.æœ€åŽå°‘äº†ä¸€ä¸ªA','',0,0,0,0,0,NULL,NULL,NULL),
	(15,2,9,'2025-06-29 17:21:41',NULL,NULL,NULL,1,'[Client #200 - CHERIFI MOHAMMED] å®¢æˆ·è½¦æž¶å·é”™ï¼Œå†ç¡®è®¤','LLV2C3B2XS0001496è¿˜æ˜¯LLV2C3B25S0001468','',0,0,0,0,0,NULL,NULL,NULL),
	(16,2,1,'2025-06-29 18:10:15',NULL,NULL,NULL,2,'[Client #208 - HAMDACHE ACHOUR] 2 clients with same id number','your id photo is wrong person','',0,0,0,0,0,NULL,NULL,NULL),
	(17,1,2,'2025-06-30 08:42:26','2025-07-08 11:31:39','2025-07-09 12:25:37',NULL,2,'[Car #64 - TIGUAN L] client nam wrong in BL','corrrect is TRIQUI BOCHRA in BL it is: TRIOUI BOCHRA','',1,0,0,0,0,NULL,NULL,NULL),
	(18,1,2,'2025-07-01 12:14:50',NULL,NULL,NULL,3,'[Car #62 - T-ROC STARLIGHT] this car no VIN','please check this car VIN','',1,0,0,0,0,NULL,NULL,NULL),
	(20,1,3,'2025-07-02 10:33:07',NULL,NULL,NULL,3,'[Sell Bill #135 - ADM001200325135] send docs to fellowing address','Ø®Ù…ÙŠØ³ Ø¨Ù† Ø¹Ø§Ø´ÙˆØ±  Ø­ÙŠ Ø³ØªØ± Ø§Ù„Ù„Ø­ÙŠØ§Ù†ÙŠ . Ø§Ù„Ø¹Ø²ÙŠØ²ÙŠØ© Ø§Ù„Ø¬Ù†ÙˆØ¨ÙŠØ© .Ù…ÙƒØ© Ø§Ù„Ù…ÙƒØ±Ù…Ø© .Ø§Ù„Ù…Ù…Ù„ÙƒØ© Ø§Ù„Ø¹Ø±Ø¨ÙŠØ© Ø§Ù„Ø³Ø¹ÙˆØ¯ÙŠØ© .................Ø±Ù‚Ù… Ø§Ù„Ø¬ÙˆØ§Ù„:  00966544321476','',0,0,0,0,0,NULL,NULL,NULL),
	(21,1,3,'2025-07-03 00:38:52',NULL,NULL,NULL,3,'[Client #88 - NASRI MOHAMMED] client name must excactly like passport','his name is NASRI MOHAMMED ALAMINE\nif client name is wrong bl will be wrong','',0,0,0,0,0,NULL,NULL,NULL),
	(22,6,3,'2025-07-08 06:54:02',NULL,NULL,NULL,1,'not found in stock','KHEMICI MOHAMMED ØŒ KESKES BILLEL ØŒ MOHAMMED OUSAID AOMAR \nnot found in stock','',0,0,0,0,0,NULL,NULL,NULL),
	(25,1,9,'2025-07-08 10:00:44',NULL,NULL,NULL,2,'[Car #525 - THARU TOP] this guy take the car under ishaq name','the vin is wrong','',1,0,0,0,0,NULL,NULL,NULL),
	(26,2,1,'2025-07-08 18:41:34',NULL,NULL,NULL,3,'[Car #608 - K3] once load ,should not show avalialle','like we do loading only, we don\'t sell but only load.','app problem',1,0,0,0,0,NULL,NULL,NULL),
	(27,2,1,'2025-07-09 09:06:48','2025-07-10 14:37:07',NULL,NULL,3,'[Client #87 - TORCHI AHMED] bouira office said he change to coolray, please confirm','this is the customer from 20 leo cars.','',0,0,0,0,0,NULL,NULL,NULL),
	(32,1,2,'2025-07-10 13:58:41',NULL,NULL,NULL,2,'[Car #49 - LIVAN MAN] switch cars and clients','he order manual he get auto he is army\ngive this car to angry man','',1,0,0,0,0,NULL,NULL,NULL),
	(33,1,2,'2025-07-10 13:59:05',NULL,NULL,NULL,2,'[Car #48 - LIVAN MAN] switch cars and clients','he order manual he get auto he is army\ngive this car to angry man','',1,0,0,0,0,NULL,NULL,NULL),
	(34,1,2,'2025-07-10 14:11:19',NULL,NULL,NULL,1,'[Car #131 - T-ROC STARLIGHT] load this with faster company akkon or any other.','I order 3 t-roc from walle give her one','',1,0,0,0,0,NULL,NULL,NULL),
	(35,1,2,'2025-07-10 14:17:15',NULL,NULL,NULL,1,'[Car #214 - T-ROC STARLIGHT] load this car with fastest shipping company like akkon','I order 3 t-roc from walle give this guy one','',1,0,0,0,0,NULL,NULL,NULL),
	(36,1,1,'2025-07-10 14:24:42',NULL,NULL,NULL,1,'kerouicha orders not here','please add kerouicha orders','',0,0,0,0,0,NULL,NULL,NULL),
	(37,2,2,'2025-07-10 16:09:15',NULL,NULL,NULL,3,'[Car #322 - SELTOS LUXERY BLACK ROOF] remove this to freight seltos car','make this car is avaliable to buy','',1,0,0,0,0,NULL,NULL,NULL),
	(39,2,1,'2025-07-10 17:05:19',NULL,NULL,NULL,3,'[Car #331 - SELTOS LUXERY BLACK ROOF] seltos, can change white to grey','','',1,0,0,0,0,NULL,NULL,NULL),
	(40,2,2,'2025-07-10 17:23:17',NULL,NULL,NULL,2,'[Car #331 - SELTOS LUXERY BLACK ROOF] he pay more to get earlier','give car LJD5AA1D0S0202526 to him','',1,0,0,0,0,NULL,NULL,NULL),
	(41,1,1,'2025-07-10 17:55:03',NULL,NULL,NULL,1,'[Client #305 - SAKHRI SMAIL] order coolray base auto for this guy','order coolray base auto for this guy','',0,0,0,0,0,NULL,NULL,NULL),
	(42,1,9,'2025-07-10 17:56:44',NULL,NULL,NULL,1,'[Client #87 - TORCHI AHMED] please dont forget to load livan auto for this guy','please dont forget to load livan auto for this guy','',0,0,0,0,0,NULL,NULL,NULL),
	(43,1,3,'2025-07-10 19:46:52','2025-07-20 10:10:43',NULL,NULL,1,'[Car #511 - SELTOS LUXERY BLACK ROOF] FEKRACHE HEMZA','this guy was normally white seltos \nbut because we was mixing cars finally he get grey\ncan he accept?\nplease confirm.','',1,0,0,0,0,NULL,NULL,NULL),
	(44,2,1,'2025-07-11 15:27:31',NULL,NULL,NULL,1,'[Car #236 - LIVAN AUTO] he didn\'t order livan','this is mostaganem ali customer, he order used car vin LB37824Z7NG219702 not livan, please confirm','',1,0,0,0,0,NULL,NULL,NULL),
	(45,2,2,'2025-07-15 08:07:58',NULL,NULL,NULL,2,'[cars] å®¢æˆ·ä¸‹çš„æ˜¯coolray sport','ç»™å®¢æˆ·è£…äº†ç¼¤è¶Šè¶…èƒ½LB37622Z6SX805635\n\nCar IDs: #497','',1,0,0,0,0,NULL,X'5B325D',X'5B3439375D'),
	(46,3,2,'2025-07-15 17:48:04','2025-07-20 08:33:25',NULL,NULL,1,'[cars] this guy has emgrand used car not new.','Car IDs: #379','',1,0,0,0,0,NULL,X'5B325D',X'5B3337395D'),
	(47,2,1,'2025-07-15 19:56:53',NULL,NULL,NULL,3,'[cars] please provide buyer information','Car IDs: #600','already sold',1,0,0,0,0,NULL,X'5B315D',X'5B3630305D'),
	(49,2,1,'2025-07-15 20:05:59',NULL,NULL,NULL,3,'[cars] delete 619/618 stock','Car IDs: #619','',1,0,0,0,0,NULL,X'5B315D',X'5B3631395D'),
	(50,3,NULL,'2025-07-21 09:13:28',NULL,NULL,NULL,1,'REPLACE RAHMANI BENYOUCEF WITH BOUKERROUCHE MOHAMMED','ATTABI ABDELHAMID WILL PAY 150 USD','',0,0,0,0,0,NULL,X'5B322C315D',NULL);

/*!40000 ALTER TABLE `tasks` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table tracking
# ------------------------------------------------------------

CREATE TABLE `tracking` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `container_ref` varchar(255) DEFAULT NULL,
  `tracking` varchar(255) DEFAULT NULL,
  `time` timestamp NULL DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `container_ref` (`container_ref`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;



# Dump of table transfer_details
# ------------------------------------------------------------

CREATE TABLE `transfer_details` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `id_transfer` int(11) DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `date` datetime DEFAULT NULL,
  `client_name` varchar(255) DEFAULT NULL,
  `client_mobile` varchar(255) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `description` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `id_client` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=160 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

LOCK TABLES `transfer_details` WRITE;
/*!40000 ALTER TABLE `transfer_details` DISABLE KEYS */;

INSERT INTO `transfer_details` (`id`, `id_transfer`, `amount`, `date`, `client_name`, `client_mobile`, `rate`, `description`, `notes`, `id_client`)
VALUES
	(1,4,8647884.00,'2025-05-29 05:00:16','KAMEL MIDATE','',260.40,'GOLF 8.5','',NULL),
	(2,5,4797500.00,'2025-05-29 05:02:02','HASSIB HAMID','',252.50,'TROC','',NULL),
	(4,6,17640000.00,'2025-05-29 06:35:43','lamine','',252.00,'','',72),
	(5,7,1380000.00,'2025-05-29 07:22:46','lamine','',251.00,'','',72),
	(6,8,7182000.00,'2025-05-29 07:26:17','massi bouira','',252.00,'KIA KX1','',NULL),
	(7,9,3276000.00,'2025-05-29 07:28:58','KAMEL MIDATE','',252.00,'GOLF 8.5','',NULL),
	(8,10,18000000.00,'2025-05-29 07:30:16','lamine','',250.00,'','',72),
	(9,11,11400000.00,'2025-05-29 07:33:51','merhab noureddine','',249.50,'','I get this money from agri bank',NULL),
	(10,11,7060000.00,'2025-05-29 07:35:28','DJAMEL BOUIRA','',249.50,'kx1','order of nacer',NULL),
	(11,12,33400000.00,'2025-05-29 07:37:02','lamine','',249.20,'','',72),
	(12,13,13104000.00,'2025-05-29 07:39:24','WALID BORDJ EL KIFAN','',248.20,'','',NULL),
	(13,13,3000000.00,'2025-05-29 07:39:59','DJAMEL BOUIRA','',250.00,'kx1','',NULL),
	(14,14,2650000.00,'2025-05-29 08:30:36','HAMZA BLIDA','',249.00,'kx1','',NULL),
	(15,15,12000000.00,'2025-05-29 08:32:37','DJAMEL BOUIRA','',247.50,'','',NULL),
	(16,16,44900000.00,'2025-05-29 08:33:56','lamine','',247.20,'','',72),
	(17,17,2450000.00,'2025-05-29 08:35:23','AMINE HELALI','',250.00,'livan auto','',NULL),
	(18,18,2025000.00,'2025-05-29 08:40:14','lamine','',248.50,'','',72),
	(19,19,40400000.00,'2025-05-29 08:41:41','lamine','',248.50,'','',72),
	(20,20,5060000.00,'2025-05-29 08:47:08','DJAMEL BOUIRA','',248.50,'livan auto x2','',NULL),
	(21,20,1150000.00,'2025-05-29 08:47:38','DJAMEL BOUIRA','',248.50,'','',NULL),
	(22,20,1100000.00,'2025-05-29 08:48:37','NOUREDDINE BOUIRA','',248.50,'JETOUR DASHING','',NULL),
	(23,20,2500000.00,'2025-05-29 08:49:12','AHMED BLIDA','',248.50,'livan auto','',NULL),
	(24,20,2400000.00,'2025-05-29 08:49:48','MASSI BOUIRA','',248.50,'','',NULL),
	(25,20,2400000.00,'2025-05-29 08:50:46','LOTFI DJIJEL','',248.50,'GELLY COOLRAY SPORT','from ALI TOUHAMI',NULL),
	(27,20,3000000.00,'2025-05-29 08:56:28','DJAMEL BOUIRA','',248.50,'','',NULL),
	(28,21,38504081.00,'2025-05-29 09:00:40','lamine','',248.50,'','',72),
	(29,22,38504081.00,'2025-05-29 09:02:17','lamine','',248.50,'','',72),
	(30,23,21000000.00,'2025-05-29 09:04:11','lamine','',248.50,'','',72),
	(31,24,29143000.00,'2025-05-29 09:05:20','lamine','',247.50,'','',72),
	(32,25,2450000.00,'2025-05-29 09:08:56','BELLOUT OMAR','',247.00,'livan auto','',NULL),
	(33,25,2500000.00,'2025-05-29 09:09:33','ARRAR LAKHDAR','',247.00,'livan auto','',NULL),
	(34,25,2500000.00,'2025-05-29 09:10:09','ARRAR ABDELMALEK','',247.00,'livan auto','',NULL),
	(35,25,2500000.00,'2025-05-29 09:10:39','LAYACHI RACHDA','',247.00,'livan auto','',NULL),
	(36,25,800000.00,'2025-05-29 09:11:06','CHOUDER MED TAHAR','',247.00,'PEUGEOT 2008','',NULL),
	(37,25,3250000.00,'2025-05-29 09:11:46','BEN HAFFAF AHMED','',247.00,'PEUGEOT 2008','client of bouira office',NULL),
	(38,25,15000000.00,'2025-05-29 09:13:27','BOULHARES NASREDDINE','',247.00,'livan auto','',NULL),
	(39,25,2500000.00,'2025-05-29 09:13:56','GARTI MOHAMMED','',247.00,'livan auto','',NULL),
	(40,25,2400000.00,'2025-05-29 09:14:46','BENHAFFAF KHIREDDINE','',247.00,'livan auto','client of bouira office',NULL),
	(41,25,2400000.00,'2025-05-29 09:16:14','BENHAFFAF ABDELKARIM','',247.00,'livan auto','client of bouira office',NULL),
	(42,26,27065025.00,'2025-05-29 09:18:06','lamine','',247.00,'','',72),
	(43,27,25935000.00,'2025-05-29 09:19:36','lamine','',247.00,'','',72),
	(44,28,73300000.00,'2025-05-29 09:22:30','lamine','',246.80,'','',72),
	(45,30,900000.00,'2025-06-12 08:01:05','FATMI NASSIM',NULL,246.80,'LIVAN MANUELLE','',73),
	(46,30,2230000.00,'2025-06-12 08:02:15','SEDDOUKI SMAIN',NULL,246.80,'MG MANUELLE','',75),
	(47,30,2230000.00,'2025-06-12 08:19:53','DAHMANE HALIMA',NULL,246.80,'LIVAN MANUELLE','',74),
	(48,30,2230000.00,'2025-06-12 08:22:41','SEDDOUKI SMAIN',NULL,246.80,'MG5 MANUELLE','',75),
	(49,30,2230000.00,'2025-06-12 08:23:13','MADI BOUMEDIENE',NULL,246.80,'MG5 MANUELLE','',76),
	(50,30,2490000.00,'2025-06-12 08:24:39','KERDJA ZAHIA',NULL,246.80,'LIVAN AUTO','',77),
	(51,30,2490000.00,'2025-06-12 08:26:09','HIJAZI CHEMSEDDINE',NULL,246.80,'LIVAN AUTO ','',78),
	(52,30,2480000.00,'2025-06-12 08:26:46','CHERFI MOHAMMED',NULL,246.80,'LIVAN AUTO','',79),
	(53,30,2500000.00,'2025-06-12 08:27:26','BEN AICHA BARIZA',NULL,246.80,'LIVAN AUTO','',80),
	(54,30,2248000.00,'2025-06-12 08:37:00','KAMAL MIDETTE',NULL,246.80,'','',83),
	(55,30,6900000.00,'2025-06-12 08:37:21','TOUHAMI ALI','0552663908',246.80,'','',65),
	(56,30,3740000.00,'2025-06-12 08:39:13','CHEFROUR NAAMANE',NULL,246.80,'GEELY COOLRAY ','',81),
	(57,30,21390000.00,'2025-06-12 08:41:10','GEROUICHA YOUCEF',NULL,246.80,'','',82),
	(58,29,1330000.00,'2025-06-13 01:01:41','FATMI NASSIM',NULL,246.80,'MG5 MANUELLE','RESTE OF LAST TRANSFRT TO HK OCEANS IS HERE ',73),
	(59,29,3800000.00,'2025-06-13 01:05:06','NASRI MOHAMMED',NULL,246.80,'GEELY COOLRAY FULL','',88),
	(60,29,2500000.00,'2025-06-13 01:05:49','MEZIANE DJAOUAD',NULL,246.80,'LIVAN X3PRO FULL','',89),
	(61,29,2500000.00,'2025-06-13 01:06:22','MEZIANE SEDDAM',NULL,246.80,'LIVAN X3PRO FULL','',90),
	(62,29,2500000.00,'2025-06-13 01:06:54','MEZIANI ZOULIKHA',NULL,246.80,'LIVAN X3PRO','',92),
	(63,29,45500000.00,'2025-06-13 01:08:57','lamine','please provide mobile',246.80,'TRANSFER','HIS MONEY IS IN SAME SWIFT OF 271766.61 CONFIRMED 23-4-2025',72),
	(64,31,24500000.00,'2025-06-13 01:21:55','JIJEL MIMOU CLIENTS',NULL,247.60,'10 LIVAN X3PRO FULL','',94),
	(65,31,5000000.00,'2025-06-13 01:22:34','','',247.60,'','',NULL),
	(66,31,2240000.00,'2025-06-13 01:23:00','','',247.60,'1 LIVAN MANUELLE','',NULL),
	(67,31,2500000.00,'2025-06-13 01:23:32','khebbaz fairouz','no phone',247.60,'LIVAN X3PRO AUTO','',41),
	(68,31,2480000.00,'2025-06-13 01:24:10','SOLTANI MERIEM',NULL,247.60,'LIVAN X3 PRO AUTO','',98),
	(69,31,2480000.00,'2025-06-13 01:24:38','SAYAD MOKHTAR',NULL,247.60,'LIVAN X3PRO FULL','',99),
	(70,31,720000.00,'2025-06-13 01:26:13','DJABALI YOUCEF',NULL,247.60,'LIVAN X3 PRO ','DEPOSIT',100),
	(71,31,3080000.00,'2025-06-13 01:26:51','SAHABA MOKHTAR','no mobile',247.60,'2008 PEUGUEOT','',12),
	(72,31,2500000.00,'2025-06-13 01:27:21','TAMAR MED TAREK',NULL,247.60,'LIVAN X3PRO AUTO','',102),
	(73,31,6496000.00,'2025-06-13 01:30:29','JIJEL MIMOU CLIENTS',NULL,247.60,'2 TROC STARLIGHT','REST OF THE MONEY IS IN SEC TRANSFERT -145008-',94),
	(74,32,3484000.00,'2025-06-13 01:34:14','JIJEL MIMOU CLIENTS',NULL,247.60,'2 TROC STARLIGHT','REST OF THE MONEY',94),
	(75,32,2450000.00,'2025-06-13 01:34:55','','',247.60,'LIVAN X3PRO AUTO','',NULL),
	(76,32,3111000.00,'2025-06-13 01:35:37','MERHAB MOUADH','0552663908',247.60,'JABHA B3AH M DAR','',63),
	(77,32,500000.00,'2025-06-13 01:36:40','','',247.60,'','REST OF MONEY AT ISHAQ HOUSE ON LAST TRANSFERT',NULL),
	(78,32,2921000.00,'2025-06-13 01:37:37','DJABALI OTMANE',NULL,247.60,'','',103),
	(79,32,2500000.00,'2025-06-13 01:38:10','BOUKEDIRA YOUSSOUF',NULL,247.60,'LIVAN X3PRO AUTO','',104),
	(80,32,22200000.00,'2025-06-13 01:38:40','lamine','please provide mobile',247.60,'','',72),
	(81,33,14300000.00,'2025-06-13 02:33:38','OUADI MOHAMED OUALID','0554676748',247.70,'','',86),
	(82,33,1416000.00,'2025-06-13 02:37:15','SEKSAOUI ISHAK','0552663908',247.70,'5500 EURO SOLD','',61),
	(83,33,127500.00,'2025-06-13 02:38:46','SEKSAOUI ISHAK','0552663908',247.70,'500 EURO SOLD','',61),
	(84,35,16550000.00,'2025-06-13 03:23:20','AHMED BELGROUNE OFFUX',NULL,247.60,'INVOICE ','',105),
	(85,35,13100000.00,'2025-06-13 03:24:20','BAGHDADI SAHEB BELGROUNE',NULL,247.60,'','',106),
	(86,35,23000000.00,'2025-06-13 03:25:32','ATTABI HAMID 44',NULL,247.60,'','',107),
	(87,36,4493000.00,'2025-06-13 03:30:53','ATTABI HAMID 44',NULL,246.40,'TROC STARLIGHT','',107),
	(88,36,5000000.00,'2025-06-13 03:31:40','BAGHDADI SAHEB BELGROUNE',NULL,246.40,'2 LIVAN X3PRO ','BENCHENINA ABDENNOUR',106),
	(89,36,3288000.00,'2025-06-13 03:32:46','MATAOUI ABDELHAMID',NULL,246.40,'','',108),
	(90,36,14100000.00,'2025-06-13 03:33:47','OUADI MOHAMED OUALID','0554676748',246.40,'','',86),
	(91,36,3930000.00,'2025-06-13 03:35:08','GARTI MOKRANE',NULL,246.40,'','',109),
	(92,36,2230000.00,'2025-06-13 03:37:41','FERHATI YAHIA',NULL,246.40,'','',110),
	(93,36,2230000.00,'2025-06-13 03:38:12','ZENATI DJAMEL',NULL,246.40,'','',111),
	(95,36,48000.00,'2025-06-13 03:40:37','SEKSAOUI ISHAK','0552663908',246.40,'','',61),
	(97,36,4990000.00,'2025-06-13 03:42:17','KORIDJIDJ MOHAMMED','no mobile',246.40,'','',7),
	(98,36,4990000.00,'2025-06-13 03:43:24','BOUALI-YOUCEF AHMED','no mobile',246.40,'TROC STARLIGHT','',8),
	(99,36,2000000.00,'2025-06-13 03:44:10','BENDERKAL OMAR',NULL,246.40,'','',115),
	(100,37,51723000.00,'2025-06-13 04:00:07','lamine','please provide mobile',246.30,'','',72),
	(101,38,63277000.00,'2025-06-13 04:02:00','lamine','please provide mobile',246.30,'','',72),
	(102,39,27920000.00,'2025-06-13 04:03:27','lamine','please provide mobile',246.00,'','',72),
	(103,40,-2500000.00,'2025-06-13 04:10:15','BOUKEDIRA YOUSSOUF',NULL,246.50,'CANCEL ORDER ','LIVANX3PRO',104),
	(104,40,4000000.00,'2025-06-13 04:20:13','BESSOU 44',NULL,246.50,'','',116),
	(105,40,17450000.00,'2025-06-13 04:22:31','KAMAL MIDETTE',NULL,246.50,'','',83),
	(106,40,200000.00,'2025-06-13 04:24:44','KHENNOUCI ZOHIR','0554860518',246.50,'','',53),
	(107,40,400000.00,'2025-06-13 04:25:51','TAIBI WALID',NULL,246.50,'','',118),
	(108,40,2600000.00,'2025-06-13 04:26:17','KHENOUCI OUALID','0797529235',246.50,'','',68),
	(109,40,2300000.00,'2025-06-13 04:27:25','ASEKRI ABDELMOUMEN',NULL,246.50,'','',119),
	(110,40,1030000.00,'2025-06-13 04:27:53','ATTABI HAMID 44',NULL,246.50,'','',107),
	(111,40,42000.00,'2025-06-13 04:28:23','BESSOU 44',NULL,246.50,'','',116),
	(112,40,3175000.00,'2025-06-13 04:30:35','TOUAT MOURAD',NULL,246.50,'','',120),
	(113,40,2775000.00,'2025-06-13 04:32:10','MESMOUDI HAMZA',NULL,246.50,'','',121),
	(114,40,791000.00,'2025-06-13 04:32:41','SEKSAOUI ISHAK','0552663908',246.50,'','KANET 3AND SALIMA FI ORAN',61),
	(115,40,11957000.00,'2025-06-13 04:34:30','KHEDDAR MOHAMMED',NULL,264.50,'','',122),
	(116,40,2425000.00,'2025-06-13 04:40:15','CHOUDER MOHAMED TAHAR','0661632094',246.50,'','',123),
	(118,40,-7995000.00,'2025-06-13 04:42:38','MERHAB NOUREDDINE','0552663908',246.00,'SI ALI SAID','32500 USD X 246',67),
	(119,41,10600000.00,'2025-06-13 05:01:13','AHMED BELGROUNE OFFUX',NULL,246.20,'','',105),
	(120,41,2650000.00,'2025-06-13 05:19:39','SENIGRA ABDELDJALIL',NULL,246.20,'GEELY CHONENG AUTO','',124),
	(121,41,2650000.00,'2025-06-13 05:22:35','OUMERZAK ADEL',NULL,246.20,'GEELY CHONENG AUTO','',125),
	(122,41,1900000.00,'2025-06-13 05:23:35','TAIBI WALID',NULL,246.20,'LIVAN MANUELLE','VERS 400000 14/5\n',118),
	(123,41,10600000.00,'2025-06-13 05:26:59','BOUMAZA KAMEL /MIMOU',NULL,246.20,'4 GEELY CHONENG AUTO','2900000 \nREST IN THE NEXT SWIFT',126),
	(125,2,7800000.00,'2025-06-13 05:33:26','KHENOUCI OUALID','0797529235',243.78,'3 LIVAN X3 PRO AUTO','',68),
	(126,2,2650000.00,'2025-06-13 05:36:09','BABA AMMAR MOHAMMED',NULL,243.78,'1 GEELY CHONENG AUTO','',127),
	(127,2,2600000.00,'2025-06-13 05:38:32','KHEDDAR MOHAMMED',NULL,243.78,'LIVAN X3PRO AUTO','KHEDDAR LEILA CAR',122),
	(128,2,-880000.00,'2025-06-13 05:42:20','MERHAB NOUREDDINE','0552663908',275.00,'3200 EURO X 275','BUYED FROM KHEDDAR MOHAMMED',67),
	(129,2,2650000.00,'2025-06-13 05:52:13','BOUGERA AMIR',NULL,243.78,'','',129),
	(130,2,2100000.00,'2025-06-13 05:55:04','KHENNOUCI ZOHIR','0554860518',243.78,'LIVAN X3PRO MANUAL','DEPOSIT 14-4-2025\n',53),
	(131,2,-4980000.00,'2025-06-13 05:57:19','HIJAZI CHEMSEDDINE',NULL,243.78,'CANCEL ORDER','2 LIVAN AUTO',78),
	(132,2,-1911000.00,'2025-06-13 05:59:53','MERHAB NOUREDDINE','0552663908',275.00,'','1600 USD X 246 DZD\n2900 USD X 275 DZD',67),
	(133,2,-12200000.00,'2025-06-13 06:02:08','MERHAB NOUREDDINE','0552663908',244.00,'50000 USD X 244','SI ALI SAID TRAVEL',67),
	(134,2,-2500000.00,'2025-06-13 06:03:16','LAYACHI RECHDA',NULL,250.00,'CANCEL ORDER','LIVAN X3PRO',131),
	(135,2,34450000.00,'2025-06-13 06:11:39','MASSI BOUIRA',NULL,243.78,'13 GEELY CHONENG AUTO','',132),
	(136,42,35468000.00,'2025-06-13 06:21:27','ATALLAH NOUREDDINE','0552663908',241.50,'NOUREDDINE BOUIRA','',57),
	(137,42,-4260000.00,'2025-06-13 06:22:11','MERHAB NOUREDDINE','0552663908',242.50,'17569 USD X 242.5','SI ALI SAID TRAVEL',67),
	(138,42,-120000.00,'2025-06-13 06:23:00','KHENOUCI OUALID','0797529235',241.50,'Ù…Ø³Ø§Ø¹Ø¯Ø© ÙÙŠ Ø§Ù„Ø³Ø¹Ø± ØªØ§Ø¹ Ù„ÙŠÙØ§Ù†','',68),
	(139,43,3950000.00,'2025-06-13 06:28:15','BELHAMRI YOUNES',NULL,240.70,'SKODA KAMIQ GT FULL','',135),
	(140,43,2400000.00,'2025-06-13 06:28:53','lamine','please provide mobile',240.70,'TA3 ALI TOUHAMI','',72),
	(141,43,5300000.00,'2025-06-13 06:29:22','MASSI BOUIRA',NULL,240.70,'2 LAST GEELY CHONENG','',132),
	(142,43,12318500.00,'2025-06-13 06:40:33','MEBARKI MESSAOUD','+213 799916303',240.70,'6 GEELY CHONENG AUTO','DEPOSIT 13000 EURO X 275.5',46),
	(143,44,6934000.00,'2025-06-19 14:11:05','MEBARKI MESSAOUD','+213 799916303',238.50,'','Ø¨Ø§Ù‚ÙŠ Ø§Ù„ÙØ§ØªÙˆØ±Ø© 5384000 ÙÙŠ Ø§Ù„ØªØ­ÙˆÙŠÙ„ Ø§Ù„Ù‚Ø§Ø¯Ù…',46),
	(144,44,1816000.00,'2025-06-19 14:13:02','MATAOUI ABDELHAMID','0667165385',238.50,'ÙƒÙ…Ù„ Ø­Ø³Ø§Ø¨ Ø§Ù„Ø³Ø§Ø±Ø§Øª ','Ø¨Ø§Ù‚ÙŠ Ø­Ø³Ø§Ø¨ Ø§Ù„Ù†Ù‚Ù„',108),
	(145,44,3850000.00,'2025-06-19 14:13:42','MENASRIA RIDHA','0661490019',238.50,'SELTOS LUXURY ','',84),
	(146,44,3800000.00,'2025-06-19 14:14:40','ABDOUN ABDELMALEK','0770125406',238.50,'KIA SELTOS LUXURY','',85),
	(147,44,6600000.00,'2025-06-19 14:15:16','OUADI MOHAMED OUALID','0554676748',238.50,'','',86),
	(149,45,2500000.00,'2025-06-19 14:49:32','OUADI MOHAMED OUALID','0554676748',238.50,'','',86),
	(150,45,3700000.00,'2025-06-19 14:50:06','AMAROUCHE DJABER','+213661996418',238.50,'KIA SELTOS LUXURY','',238),
	(151,45,2600000.00,'2025-06-19 14:50:26','KAMAL MIDETTE',NULL,238.50,'','',83),
	(152,45,1200000.00,'2025-06-19 14:51:14','MOGAFI HAMZA','0',238.50,'T-ROC','ÙƒÙ…Ù„ Ø§Ù„Ø­Ø³Ø§Ø¨\nCOMPLET PAIMENT',237),
	(153,45,2000000.00,'2025-06-19 14:55:28','ATTAB RAMZI',NULL,238.50,'MG 5 GREY','DEPOSIT ',247),
	(154,45,3700000.00,'2025-06-19 14:55:54','MEZAOUGH KHALID','+213770982382',238.50,'KIA SELTOS LUXURY ','',239),
	(155,45,12300000.00,'2025-06-19 14:57:31','OUADI MOHAMED OUALID','0554676748',238.50,'','Ø²Ø§Ø¯Ù†ÙŠ 65000.00 Ø¨Ø§Ø´ ÙŠÙƒÙ…Ù„ 90 Ø§Ù„Ù Ø¯ÙˆÙ„Ø§Ø±\n\nHE ADDED 65000.00 DZD TO FINISH 90000 DOOLAR PAIMENT',86),
	(156,45,9600000.00,'2025-06-19 14:58:40','KAMAL MIDETTE',NULL,238.50,'','Ø®Ù„Øµ 9660000.00\n6 Ù…Ù„Ø§ÙŠÙ† Ø¨Ø§Ù‚ÙŠØ© Ø¹Ù†Ø¯ÙŠ ÙÙŠ Ø§Ù„Ù…ÙƒØªØ¨\nHE PAID 9660000.00 \n60000 STILL IN HOUSE',83),
	(157,45,3700000.00,'2025-06-19 14:59:48','KIFOUCHE SAMIR','213661716653',238.50,'KIA SELTOS LUXURY','',241),
	(158,45,3700000.00,'2025-06-19 15:00:19','FEKRACHE HEMZA','213661 71 66 53',238.50,'KIA SELTOS LUXURY','',240),
	(159,46,16730000.00,'2025-06-19 15:08:13','BEKHOUCHE SAID','+213560738705//0555838181',239.00,'','RESTE 3627000 FROM TROC MONEY 5300000',245);

/*!40000 ALTER TABLE `transfer_details` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table transfers
# ------------------------------------------------------------

CREATE TABLE `transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_user_do_transfer` int(11) NOT NULL,
  `date_do_transfer` datetime NOT NULL,
  `amount_sending_da` decimal(10,2) NOT NULL,
  `rate` decimal(10,2) NOT NULL,
  `id_user_receive_transfer` int(11) DEFAULT NULL,
  `amount_received_usd` decimal(10,2) DEFAULT 0.00,
  `date_receive` datetime DEFAULT NULL,
  `details_transfer` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`details_transfer`)),
  `notes` text DEFAULT NULL,
  `receiver_notes` text DEFAULT NULL,
  `id_bank` int(11) DEFAULT NULL,
  `ref_pi_transfer` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_user_do_transfer` (`id_user_do_transfer`),
  CONSTRAINT `transfers_ibfk_1` FOREIGN KEY (`id_user_do_transfer`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=47 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `transfers` WRITE;
/*!40000 ALTER TABLE `transfers` DISABLE KEYS */;

INSERT INTO `transfers` (`id`, `id_user_do_transfer`, `date_do_transfer`, `amount_sending_da`, `rate`, `id_user_receive_transfer`, `amount_received_usd`, `date_receive`, `details_transfer`, `notes`, `receiver_notes`, `id_bank`, `ref_pi_transfer`)
VALUES
	(2,3,'2025-05-26 08:37:10',40800000.00,243.78,NULL,167364.02,NULL,NULL,'9 300 000.00 DZD X 243\n31 150 000.00DZD X 244\n',NULL,1,NULL),
	(4,3,'2025-01-13 00:00:00',8647884.00,260.40,NULL,33210.00,NULL,NULL,NULL,NULL,2,'MER 2504'),
	(5,3,'2025-02-18 00:00:00',12113687.50,252.50,NULL,19000.00,NULL,NULL,NULL,NULL,2,'MER 2510A'),
	(6,3,'2025-03-05 00:00:00',17640000.00,252.00,NULL,70000.00,NULL,NULL,NULL,NULL,2,'MER 2502'),
	(7,3,'2025-03-10 00:00:00',34638000.00,251.00,NULL,138000.00,NULL,NULL,NULL,NULL,2,'no pi'),
	(8,3,'2025-03-11 00:00:00',7182000.00,252.00,NULL,28500.00,NULL,NULL,NULL,NULL,2,'PY 2501'),
	(9,3,'2025-03-11 00:00:00',3276000.00,252.00,NULL,13000.00,NULL,NULL,NULL,NULL,2,'PY 2504'),
	(10,3,'2025-03-13 00:00:00',18000000.00,250.00,NULL,72000.00,NULL,NULL,NULL,NULL,2,'PY 2505'),
	(11,3,'2025-03-15 00:00:00',18460000.00,249.50,NULL,73987.98,NULL,NULL,NULL,NULL,2,'PY 2507'),
	(12,3,'2025-03-17 00:00:00',33400000.00,249.20,NULL,134028.89,NULL,NULL,NULL,NULL,2,'PY 2505'),
	(13,3,'2025-03-19 00:00:00',16104000.00,248.20,NULL,64883.16,NULL,NULL,NULL,NULL,2,'PY 2509'),
	(14,3,'2025-03-22 00:00:00',2650000.00,249.00,NULL,10642.57,NULL,NULL,NULL,NULL,2,'SY 2509'),
	(15,3,'2025-03-24 00:00:00',12000000.00,247.50,NULL,48484.85,NULL,NULL,NULL,NULL,2,'PY 2510'),
	(16,3,'2025-03-24 00:00:00',44900000.00,247.20,NULL,181634.30,NULL,NULL,NULL,NULL,2,'PY 2511'),
	(17,3,'2025-03-24 00:00:00',2450000.00,250.00,NULL,9800.00,NULL,NULL,NULL,NULL,2,'SY 2513'),
	(18,3,'2025-03-26 00:00:00',2025000.00,248.50,NULL,8148.89,NULL,NULL,'6000 EURO TO MOUAD AND MOHAMMED\nI am noureddine they already return back this ',NULL,3,NULL),
	(19,3,'2025-03-28 00:00:00',40400000.00,248.50,NULL,162575.45,NULL,NULL,NULL,NULL,1,'ACCDZ 250328'),
	(20,3,'2025-04-01 00:00:00',17501855.00,248.50,NULL,70430.00,NULL,NULL,'calculation problem please check',NULL,2,'MER251701-02'),
	(21,3,'2025-04-06 00:00:00',38504081.00,248.50,NULL,154946.00,NULL,NULL,NULL,NULL,1,'ISH2501'),
	(22,3,'2025-04-06 00:00:00',38504081.00,248.50,NULL,154946.00,NULL,NULL,NULL,NULL,1,'ISH2501'),
	(23,3,'2025-04-06 00:00:00',21000000.00,248.50,NULL,84507.04,NULL,NULL,NULL,NULL,1,'ISH2502'),
	(24,3,'2025-04-12 00:00:00',29143000.00,247.50,NULL,117749.49,NULL,NULL,NULL,NULL,2,'PY2516'),
	(25,3,'2025-04-12 00:00:00',36300000.00,247.00,NULL,146963.56,NULL,NULL,NULL,NULL,2,'no pi'),
	(26,3,'2025-04-15 00:00:00',27065025.00,247.00,NULL,109575.00,NULL,NULL,NULL,NULL,1,'no pi'),
	(27,3,'2025-04-15 00:00:00',25935000.00,247.00,NULL,105000.00,NULL,NULL,NULL,NULL,2,'no pi'),
	(28,3,'2025-04-17 00:00:00',73300000.00,246.80,NULL,297001.62,NULL,NULL,NULL,NULL,2,'no pi'),
	(29,3,'2025-04-20 00:00:00',67072000.00,246.80,NULL,271766.61,NULL,NULL,'87406,81 MERHAB MONEY\n184359    LAMINE ANNABA MONEY',NULL,1,NULL),
	(30,3,'2025-04-21 00:00:00',51828000.00,246.80,NULL,210000.00,NULL,NULL,NULL,NULL,4,NULL),
	(31,3,'2025-04-24 00:00:00',51996000.00,247.60,NULL,210000.00,NULL,NULL,NULL,NULL,2,NULL),
	(32,3,'2025-04-23 00:00:00',35904000.00,247.60,NULL,145008.08,NULL,NULL,'89660.74 LAMINE MONEY\n55347.33 MERHAB MONEY',NULL,1,NULL),
	(33,3,'2025-04-28 00:00:00',28800000.00,247.70,NULL,116269.68,NULL,NULL,NULL,NULL,1,NULL),
	(34,3,'2025-04-28 00:00:00',59700000.00,247.50,NULL,241212.12,NULL,NULL,'LAMINE MONEY',NULL,1,NULL),
	(35,3,'2025-04-29 00:00:00',52700000.00,247.60,NULL,212843.30,NULL,NULL,NULL,NULL,1,NULL),
	(36,3,'2025-05-05 00:00:00',47300000.00,246.40,NULL,191964.29,NULL,NULL,NULL,NULL,1,NULL),
	(37,3,'2025-05-05 00:00:00',51723000.00,246.30,NULL,210000.00,NULL,NULL,NULL,NULL,4,NULL),
	(38,3,'2025-05-05 00:00:00',63277000.00,246.30,NULL,256910.27,NULL,NULL,NULL,NULL,2,NULL),
	(39,3,'2025-05-10 00:00:00',27920000.00,246.00,NULL,113495.93,NULL,NULL,NULL,NULL,2,NULL),
	(40,3,'2025-05-19 00:00:00',36073000.00,246.50,NULL,146340.77,NULL,NULL,NULL,NULL,1,NULL),
	(41,3,'2025-05-21 00:00:00',20700000.00,246.20,NULL,84077.99,NULL,NULL,NULL,NULL,1,NULL),
	(42,3,'2025-05-29 00:00:00',31177000.00,241.50,NULL,129097.31,NULL,NULL,NULL,NULL,1,NULL),
	(43,3,'2025-05-31 00:00:00',15800000.00,240.70,NULL,65641.88,NULL,NULL,NULL,NULL,1,NULL),
	(44,3,'2025-06-14 00:00:00',23000000.00,238.50,NULL,96436.06,NULL,NULL,NULL,NULL,1,NULL),
	(45,3,'2025-06-15 00:00:00',45000000.00,238.50,NULL,188679.25,NULL,NULL,NULL,NULL,1,NULL),
	(46,3,'2025-06-19 00:00:00',1673000.00,239.00,6,7000.00,'2025-06-20 00:00:00',NULL,'7000 FROM SI ALI SAID TRAVEL ',NULL,3,NULL);

/*!40000 ALTER TABLE `transfers` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table transfers_inter
# ------------------------------------------------------------

CREATE TABLE `transfers_inter` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `amount` decimal(10,2) DEFAULT NULL,
  `date_transfer` date DEFAULT NULL,
  `from_user_id` int(11) DEFAULT NULL,
  `to_user_id` int(11) DEFAULT NULL,
  `id_admin_confirm` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `date_received` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `transfers_inter` WRITE;
/*!40000 ALTER TABLE `transfers_inter` DISABLE KEYS */;

INSERT INTO `transfers_inter` (`id`, `amount`, `date_transfer`, `from_user_id`, `to_user_id`, `id_admin_confirm`, `notes`, `date_received`)
VALUES
	(1,12000000.00,'2025-06-27',4,3,NULL,'we was all in mostaganem office nour ali ishaq',NULL),
	(2,450000.00,'2025-06-27',4,1,NULL,'TO CCP OF NOUREDDINE',NULL);

/*!40000 ALTER TABLE `transfers_inter` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table users
# ------------------------------------------------------------

CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `role_id` (`role_id`),
  CONSTRAINT `users_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;

INSERT INTO `users` (`id`, `username`, `email`, `password`, `role_id`)
VALUES
	(1,'admin','admin@example.com','$2y$10$2jK5Ztrj1.DzRMKXM2fDqeeMzZRtyKk7n09NrmtdmCm4rgpbkPb1q',1),
	(2,'sofia','sofia@merhab.com','$2y$10$ulZdwEdrxuBmMpETAbCkt.2Xv/fyAo6xDIRkNru1ZabL.dqbTokM6',1),
	(3,'ishaq','ishaq@merhab.com','$2y$10$A3bjUPMv8W.Hw1MCLTBUQOa4r72WyZ5gI0pm1./FpnL26te9V7Nay',6),
	(4,'touhami','touhali@merhab.com','$2y$10$NL6iBEJjhKdifFJmZwfFEuRhSqdCnuSbG7kIQpo5NIA1BpFwT0q2C',7),
	(5,'atellah','atellah@merhab.com','$2y$10$lIWVvIhpb81chBuFYTjHgudJfw4EmXVjus5WpDaUMmTVFSA2G.o4G',7),
	(6,'mohamed','mohamed@merhab.com','$2y$10$f/8LJ9iYlUDheZTPIxml7OzRMqrDcfusKos37m/B6lBl8jsr67xBO',1),
	(7,'moadh','moadh@merhab.com','$2y$10$HvSJrZaZN90GvKXCokHk2Ojw1bKLrBPuEPqCI4gELJcDUPEYrYnhK',7),
	(9,'ali','ali@merhab.com','$2y$10$xk9Kh/WwDsRGOBvSQG4jO.Lwvwfnl1wXFmjCdvrD2ahCq5oGHYGSO',1),
	(10,'massi','massi@merhab.com','$2y$10$fPm4SeJlZn8tRqt1SX6/Me6v26aP3Ul63XiQYLgXHxWA/X3hsO/GC',7),
	(11,'boukabous','boukabous@merhab.com','$2y$10$O.xE0wABzujU1RL/wQN4se0LOQHKMEq6qpxZIQTEm.qXw5tfI0dXm',1),
	(12,'kamel','kamel@merhab.com','$2y$10$4xt1e6TlU7HfdIuOz3HF4e4QTlgbEvKJjvAWrJFy9jHW2E8AHV86u',7);

/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table versions
# ------------------------------------------------------------

CREATE TABLE `versions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `version` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `versions` WRITE;
/*!40000 ALTER TABLE `versions` DISABLE KEYS */;

INSERT INTO `versions` (`id`, `version`)
VALUES
	(1,4);

/*!40000 ALTER TABLE `versions` ENABLE KEYS */;
UNLOCK TABLES;


# Dump of table warehouses
# ------------------------------------------------------------

CREATE TABLE `warehouses` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `warhouse_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;

LOCK TABLES `warehouses` WRITE;
/*!40000 ALTER TABLE `warehouses` DISABLE KEYS */;

INSERT INTO `warehouses` (`id`, `warhouse_name`, `notes`)
VALUES
	(1,'nansha minfeng1 ',''),
	(2,'chunhua 1',''),
	(3,'MERHAB shanghai warehouse ',''),
	(5,'Nansha merhab warehouse',''),
	(7,'supplier warehouse','');

/*!40000 ALTER TABLE `warehouses` ENABLE KEYS */;
UNLOCK TABLES;



/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
