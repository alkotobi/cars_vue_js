# ************************************************************
# Sequel Ace SQL dump
# Version 20086
#
# https://sequel-ace.com/
# https://github.com/Sequel-Ace/Sequel-Ace
#
# Host: 216.219.81.100 (MySQL 5.5.5-10.6.21-MariaDB)
# Database: merhab_cars
# Generation Time: 2025-05-24 23:40:02 +0000
# ************************************************************


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
SET NAMES utf8mb4;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE='NO_AUTO_VALUE_ON_ZERO', SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


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
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table brands
# ------------------------------------------------------------

CREATE TABLE `brands` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `brand` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `brand` (`brand`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table buy_bill
# ------------------------------------------------------------

CREATE TABLE `buy_bill` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_supplier` int(11) DEFAULT NULL,
  `date_buy` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `payed` decimal(10,2) DEFAULT NULL,
  `pi_path` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_supplier` (`id_supplier`),
  CONSTRAINT `buy_bill_ibfk_1` FOREIGN KEY (`id_supplier`) REFERENCES `suppliers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



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
  PRIMARY KEY (`id`),
  KEY `id_car_name` (`id_car_name`),
  KEY `id_color` (`id_color`),
  KEY `id_buy_bill` (`id_buy_bill`),
  CONSTRAINT `buy_details_ibfk_1` FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`),
  CONSTRAINT `buy_details_ibfk_2` FOREIGN KEY (`id_color`) REFERENCES `colors` (`id`),
  CONSTRAINT `buy_details_ibfk_3` FOREIGN KEY (`id_buy_bill`) REFERENCES `buy_bill` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



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
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table cars_names
# ------------------------------------------------------------

CREATE TABLE `cars_names` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_name` varchar(255) DEFAULT NULL,
  `id_brand` int(11) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `car_name` (`car_name`),
  KEY `id_brand` (`id_brand`),
  CONSTRAINT `cars_names_ibfk_1` FOREIGN KEY (`id_brand`) REFERENCES `brands` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



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
  PRIMARY KEY (`id`),
  KEY `id_client` (`id_client`),
  KEY `id_port_loading` (`id_port_loading`),
  KEY `id_port_discharge` (`id_port_discharge`),
  KEY `id_buy_details` (`id_buy_details`),
  CONSTRAINT `cars_stock_ibfk_1` FOREIGN KEY (`id_client`) REFERENCES `clients` (`id`),
  CONSTRAINT `cars_stock_ibfk_2` FOREIGN KEY (`id_port_loading`) REFERENCES `loading_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_3` FOREIGN KEY (`id_port_discharge`) REFERENCES `discharge_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_4` FOREIGN KEY (`id_buy_details`) REFERENCES `buy_details` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table clients
# ------------------------------------------------------------

CREATE TABLE `clients` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobiles` varchar(255) NOT NULL,
  `id_copy_path` varchar(255) DEFAULT NULL,
  `id_no` varchar(255) DEFAULT NULL,
  `is_broker` tinyint(1) DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `client_name_unic` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table colors
# ------------------------------------------------------------

CREATE TABLE `colors` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `color` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `color` (`color`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table discharge_ports
# ------------------------------------------------------------

CREATE TABLE `discharge_ports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `discharge_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discharge_port` (`discharge_port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table loading_ports
# ------------------------------------------------------------

CREATE TABLE `loading_ports` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `loading_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `loading_port` (`loading_port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table permissions
# ------------------------------------------------------------

CREATE TABLE `permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `permission_name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `permission_name` (`permission_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table rates
# ------------------------------------------------------------

CREATE TABLE `rates` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `date_rate` datetime DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table role_permissions
# ------------------------------------------------------------

CREATE TABLE `role_permissions` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role_id` int(11) NOT NULL,
  `permission_id` int(11) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_permission_unique` (`role_id`,`permission_id`),
  KEY `permission_id` (`permission_id`),
  CONSTRAINT `role_permissions_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `role_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table roles
# ------------------------------------------------------------

CREATE TABLE `roles` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `role_name` varchar(255) NOT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_name` (`role_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table sell_bill
# ------------------------------------------------------------

CREATE TABLE `sell_bill` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_client` int(11) DEFAULT NULL,
  `date_sell` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `payed` decimal(10,2) DEFAULT NULL,
  `pi_path` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table sell_payments
# ------------------------------------------------------------

CREATE TABLE `sell_payments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_sell_bill` int(11) DEFAULT NULL,
  `date_payment` datetime DEFAULT NULL,
  `amount_da` decimal(10,2) DEFAULT NULL,
  `amount_usd` decimal(10,2) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `swift_path` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table suppliers
# ------------------------------------------------------------

CREATE TABLE `suppliers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobiles` varchar(255) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `supplier_name_unic` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table transfers
# ------------------------------------------------------------

CREATE TABLE `transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_user_do_transfer` int(11) DEFAULT NULL,
  `id_user_receive_transfer` int(11) DEFAULT NULL,
  `date_transfer` datetime DEFAULT NULL,
  `date_receive` datetime DEFAULT NULL,
  `amount_sending_da` decimal(10,2) DEFAULT NULL,
  `amount_received_usd` decimal(10,2) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `swift_path` varchar(255) DEFAULT NULL,
  KEY `id_user_do_transfer` (`id_user_do_transfer`),
  CONSTRAINT `transfers_ibfk_1` FOREIGN KEY (`id_user_do_transfer`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table transfers_inter
# ------------------------------------------------------------

CREATE TABLE `transfers_inter` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `from_user_id` int(11) DEFAULT NULL,
  `to_user_id` int(11) DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `date_transfer` datetime DEFAULT NULL,
  `notes` text DEFAULT NULL,
  `id_admin_confirm` int(11) DEFAULT NULL,
  `date_received` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table users
# ------------------------------------------------------------

CREATE TABLE `users` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(255) NOT NULL,
  `email` varchar(255) DEFAULT NULL,
  `password` varchar(255) NOT NULL,
  `role_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  KEY `role_id` (`role_id`),
  CONSTRAINT `users_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;



# Dump of table warehouses
# ------------------------------------------------------------

CREATE TABLE `warehouses` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `warhouse_name` varchar(255) DEFAULT NULL,
  `notes` text DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;




/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
