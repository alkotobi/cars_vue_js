-- Database Setup SQL File
-- This file contains all CREATE TABLE statements for setting up the databases
-- Execute this file to create all necessary tables

-- ============================================
-- Database: merhab_databases
-- ============================================

-- Login table
CREATE TABLE IF NOT EXISTS `login` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user` varchar(255) DEFAULT NULL,
  `pass` text,
  `active` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Databases table
CREATE TABLE IF NOT EXISTS `dbs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `db_code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `db_host_start` date DEFAULT NULL,
  `db_host_end` date DEFAULT NULL,
  `serv_host_start` date DEFAULT NULL,
  `serv_host_end` date DEFAULT NULL,
  `db_host_cost_per_month` double DEFAULT NULL,
  `serv_host_cost_per_month` double DEFAULT NULL,
  `files_dir` varchar(255) DEFAULT NULL,
  `js_dir` varchar(255) DEFAULT NULL,
  `db_name` varchar(255) DEFAULT NULL,
  `is_created` int DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`db_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Database updates table (for version migration)
CREATE TABLE IF NOT EXISTS `db_updates` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `sql` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `from_version` int NOT NULL,
  `current_version` int NOT NULL,
  `description` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Advanced SQL table
CREATE TABLE IF NOT EXISTS `adv_sql` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `stmt` text,
  `desc` varchar(255) DEFAULT NULL,
  `params` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `param_values` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  CONSTRAINT `adv_sql_chk_1` CHECK (json_valid(`params`)),
  CONSTRAINT `adv_sql_chk_2` CHECK (json_valid(`param_values`))
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Banks table
CREATE TABLE IF NOT EXISTS `banks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `company_name` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_name` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `swift_code` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_account` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_address` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `notes` text COLLATE latin1_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

-- Brands table
CREATE TABLE IF NOT EXISTS `brands` (
  `id` int NOT NULL AUTO_INCREMENT,
  `brand` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `brand` (`brand`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default brands
INSERT IGNORE INTO `brands` (`brand`) VALUES
('AUDI'),
('CHANGAN'),
('CHERRY'),
('FREIGHT'),
('GEELY'),
('JETA'),
('JETOUR'),
('KIA'),
('MG'),
('PEUGEOT'),
('SKODA'),
('VW');

-- Buy bill table
CREATE TABLE IF NOT EXISTS `buy_bill` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_supplier` int DEFAULT NULL,
  `date_buy` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `payed` decimal(10,2) DEFAULT NULL,
  `pi_path` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  `is_stock_updated` tinyint(1) DEFAULT 0,
  `is_ordered` tinyint(1) DEFAULT '1',
  `notes` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `bill_ref` (`bill_ref`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Buy details table
CREATE TABLE IF NOT EXISTS `buy_details` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_car_name` int DEFAULT NULL,
  `id_color` int DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `notes` text,
  `QTY` int DEFAULT NULL,
  `year` int DEFAULT NULL,
  `month` int DEFAULT NULL,
  `is_used_car` tinyint(1) DEFAULT '0',
  `id_buy_bill` int DEFAULT NULL,
  `price_sell` decimal(10,2) DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Buy payments table
CREATE TABLE IF NOT EXISTS `buy_payments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_buy_bill` int DEFAULT NULL,
  `date_payment` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `swift_path` varchar(255) DEFAULT NULL,
  `notes` text,
  `id_user` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Cars names table
CREATE TABLE IF NOT EXISTS `cars_names` (
  `id` int NOT NULL AUTO_INCREMENT,
  `car_name` varchar(255) DEFAULT NULL,
  `id_brand` int DEFAULT NULL,
  `notes` text,
  `is_big_car` tinyint(1) DEFAULT '0',
  `cbm` decimal(10,0) DEFAULT NULL,
  `gw` decimal(10,0) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `car_name` (`car_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default car names (all uppercase)
INSERT IGNORE INTO `cars_names` (`car_name`) VALUES
('2008'),
('A3'),
('CHANGAN CS75 PLUS'),
('CHERRY TIGO 7'),
('CHERY TIGGO 3X'),
('COOLRAY FULL NOT SPORT'),
('COOLRAY SPORT'),
('COOLRAY SUPER AUTO NO SUNROOF'),
('COOLRAY SUPER AUTO WITH SUNROOF'),
('COOLRAY SUPER MANUAL'),
('DASHING PRO 1.6 DCT'),
('EMGRAND MAN'),
('EMGRAND USED CAR'),
('FREIGHT'),
('GOLF 300TSI R-LINE'),
('GOLF R-LINE (FULL OPTION)'),
('JETTA VS5'),
('K3'),
('KAMIQ GT'),
('KX1 20251.4LCVT SUNROOFEDITION'),
('LIVAN AUTO'),
('LIVAN MAN'),
('MG5 BASE AUTO'),
('MG5 MAN'),
('SELTOS LUXERY BLACK ROOF'),
('SONET BLACK ROOF'),
('T-CROSS'),
('T-ROC STARLIGHT'),
('T-ROC WITH PACKAKE'),
('TACOUA'),
('THARU BASIC'),
('THARU MID'),
('THARU TOP'),
('TIGO 3'),
('TIGUAN L');

-- Cars stock table
CREATE TABLE IF NOT EXISTS `cars_stock` (
  `id` int NOT NULL AUTO_INCREMENT,
  `notes` text,
  `id_buy_details` int DEFAULT NULL,
  `date_sell` datetime DEFAULT NULL,
  `id_client` int DEFAULT NULL,
  `price_cell` decimal(10,2) DEFAULT NULL,
  `cfr_da` decimal(10,2) DEFAULT NULL,
  `freight` decimal(10,2) DEFAULT NULL,
  `id_port_loading` int DEFAULT NULL,
  `id_port_discharge` int DEFAULT NULL,
  `vin` varchar(255) DEFAULT NULL,
  `path_documents` varchar(255) DEFAULT NULL,
  `date_loding` datetime DEFAULT NULL,
  `date_send_documents` datetime DEFAULT NULL,
  `hidden` tinyint(1) DEFAULT '0',
  `id_sell_pi` varchar(255) DEFAULT NULL,
  `sell_pi_path` varchar(255) DEFAULT NULL,
  `buy_pi_path` varchar(255) DEFAULT NULL,
  `id_sell` int DEFAULT NULL,
  `export_lisence_ref` varchar(255) DEFAULT NULL,
  `id_warehouse` int DEFAULT NULL,
  `in_wharhouse_date` date DEFAULT NULL,
  `date_get_documents_from_supp` date DEFAULT NULL,
  `date_get_keys_from_supp` date DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `date_get_bl` date DEFAULT NULL,
  `date_pay_freight` date DEFAULT NULL,
  `is_used_car` tinyint(1) DEFAULT NULL,
  `is_big_car` tinyint(1) DEFAULT '0',
  `container_ref` varchar(255) DEFAULT NULL,
  `is_tmp_client` tinyint(1) DEFAULT '0',
  `id_loaded_container` int DEFAULT NULL,
  `is_batch` tinyint(1) DEFAULT '0',
  `is_loading_inquiry_sent` tinyint(1) DEFAULT '0',
  `date_assigned` timestamp NULL DEFAULT NULL,
  `id_color` int DEFAULT NULL,
  `hidden_by_user_id` int DEFAULT NULL,
  `hidden_time_stamp` timestamp NULL DEFAULT NULL,
  `path_coo` varchar(255) DEFAULT NULL,
  `path_coc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `vin` (`vin`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat groups table
CREATE TABLE IF NOT EXISTS `chat_groups` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `id_user_owner` int DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Chat last read message table
CREATE TABLE IF NOT EXISTS `chat_last_read_message` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_group` int DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `id_last_read_message` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `id_group` (`id_group`,`id_user`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat messages table
CREATE TABLE IF NOT EXISTS `chat_messages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_group` int DEFAULT NULL,
  `message_from_user_id` int DEFAULT NULL,
  `chat_replay_to_message_id` int DEFAULT NULL,
  `message` text COLLATE utf8mb4_general_ci,
  `time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Chat read by table
CREATE TABLE IF NOT EXISTS `chat_read_by` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_message` int DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `read_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat users table
CREATE TABLE IF NOT EXISTS `chat_users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_user` int DEFAULT NULL,
  `id_chat_group` int DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Clients table
CREATE TABLE IF NOT EXISTS `clients` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobiles` varchar(255) DEFAULT 'please provide mobile',
  `id_copy_path` varchar(255) DEFAULT NULL,
  `id_no` varchar(255) DEFAULT NULL,
  `nin` varchar(40) DEFAULT NULL,
  `is_broker` tinyint(1) DEFAULT '0',
  `is_client` tinyint(1) DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `client_name_unic` (`name`,`id_no`),
  UNIQUE KEY `id_no` (`id_no`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Colors table
CREATE TABLE IF NOT EXISTS `colors` (
  `id` int NOT NULL AUTO_INCREMENT,
  `color` varchar(255) DEFAULT NULL,
  `hexa` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `color` (`color`),
  UNIQUE KEY `hexa` (`hexa`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default colors
INSERT IGNORE INTO `colors` ( `color`, `hexa`) VALUES
('WHITE', '#ffffff'),
( 'BLACK', '#050505'),
( 'GRINARDO', '#828580'),
( 'GREY', '#8b8989'),
( 'CHAMILION', '#7b647d'),
( 'PEARLY WHITE', '#fbf9f9');

-- Containers table
CREATE TABLE IF NOT EXISTS `containers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `nm` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default containers
INSERT IGNORE INTO `containers` (`name`) VALUES
('40 FEET HQ'),
('40 FEET'),
('20 FEET');

-- Defaults table
CREATE TABLE IF NOT EXISTS `defaults` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `rate` decimal(10,2) DEFAULT NULL,
  `freight_small` decimal(10,2) DEFAULT NULL,
  `freight_big` decimal(10,2) DEFAULT NULL,
  `alert_unloaded_after_days` int DEFAULT NULL,
  `alert_not_arrived_after_days` int DEFAULT NULL,
  `alert_no_licence_after_days` int DEFAULT NULL,
  `alert_no_docs_sent_after_days` int DEFAULT NULL,
  `max_unpayed_sell_bills` int DEFAULT '3',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

-- Insert default values
INSERT IGNORE INTO `defaults` (`id`, `rate`, `freight_small`, `freight_big`, `alert_unloaded_after_days`, `alert_not_arrived_after_days`, `alert_no_licence_after_days`, `alert_no_docs_sent_after_days`, `max_unpayed_sell_bills`) VALUES
(1, 240.00, 1600.00, 2500.00, 10, 7, 9, 20, 5);

-- Discharge ports table
CREATE TABLE IF NOT EXISTS `discharge_ports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `discharge_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discharge_port` (`discharge_port`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default discharge ports
INSERT IGNORE INTO `discharge_ports` (`discharge_port`) VALUES
('ALGIERS'),
('ANNABA'),
('MOSTAGANEM'),
('ORAN'),
('SKIKDA');

-- Loaded containers table
CREATE TABLE IF NOT EXISTS `loaded_containers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_loading` int DEFAULT NULL,
  `id_container` int DEFAULT NULL,
  `ref_container` varchar(255) DEFAULT NULL,
  `date_departed` date DEFAULT NULL,
  `note` text,
  `date_loaded` date DEFAULT NULL,
  `date_on_board` date DEFAULT NULL,
  `so` varchar(255) DEFAULT NULL,
  `is_released` tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Loading table
CREATE TABLE IF NOT EXISTS `loading` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `date_loading` date DEFAULT NULL,
  `id_shipping_line` int DEFAULT NULL,
  `freight` decimal(10,0) DEFAULT NULL,
  `id_loading_port` int DEFAULT NULL,
  `id_discharge_port` int DEFAULT NULL,
  `EDD` date DEFAULT NULL,
  `date_loaded` date DEFAULT NULL,
  `note` text COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Loading ports table
CREATE TABLE IF NOT EXISTS `loading_ports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `loading_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `loading_port` (`loading_port`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default loading ports
INSERT IGNORE INTO `loading_ports` (`loading_port`) VALUES
('NANSHA'),
('SHANGHAI');

-- Permissions table
CREATE TABLE IF NOT EXISTS `permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `permission_name` varchar(50) NOT NULL,
  `description` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `permission_name` (`permission_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default permissions
INSERT IGNORE INTO `permissions` (`permission_name`, `description`) VALUES
('can_manage_users', 'Can create, edit, and delete users'),
('can_manage_roles', 'Can create, edit, and delete roles'),
('can_manage_permissions', 'Can manage role permissions'),
('is_exchange_sender', 'Can send exchange requests'),
('is_exchange_receiver', 'Can receive exchange requests'),
('can_manage_cars', 'Can manage cars inventory'),
('can_edit_cars_prop', 'Can edit car properties'),
('can_edit_vin', 'Can edit vehicle identification number (VIN)'),
('can_edit_car_client_name', 'Can edit car client name'),
('can_edit_cars_sell_price', 'Can edit car sell price'),
('can_edit_cars_sell_rate', 'Can edit car sell rate'),
('can_edit_cars_discharge_port', 'Can edit car discharge port'),
('can_upload_car_files', 'Can upload car files'),
('can_edit_cars_ports', 'Can edit car ports'),
('can_edit_car_money', 'Can edit car money/financial information'),
('can_receive_car', 'Can receive cars'),
('can_edit_car_documents', 'Can edit car documents'),
('can_load_car', 'Can load cars'),
('can_edit_sell_payments', 'Can edit sell payments'),
('can_delete_sell_paymets', 'Can delete sell payments'),
('can_c_sell_payments', 'Can create sell payments'),
('can_delete_sell_bill', 'Can delete sell bills'),
('can_edit_sell_bill', 'Can edit sell bills'),
('can_access_cashier', 'Can access cashier functions'),
('can_purchase_cars', 'Can purchase cars'),
('can_sell_cars', 'Can sell cars'),
('can_c_car_stock', 'Can create car stock entries'),
('can_assign_to_tmp_clients', 'Can assign cars to temporary clients'),
('can_change_car_color', 'Can change car color'),
('can_c_other_users_sells', 'Can create other users sells'),
('can_unassign_cars', 'Can unassign cars'),
('can_hide_car', 'Can hide cars');

-- Priorities table
CREATE TABLE IF NOT EXISTS `priorities` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `priority` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_color` int DEFAULT NULL,
  `power` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `priority` (`priority`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Insert default priorities
INSERT IGNORE INTO `priorities` (`priority`, `id_color`, `power`) VALUES
('very important', NULL, 3),
('urgent', NULL, 4),
('important', NULL, 2);

-- Rates table
CREATE TABLE IF NOT EXISTS `rates` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `rate` decimal(10,2) DEFAULT NULL,
  `created_on` timestamp NULL DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Roles table
CREATE TABLE IF NOT EXISTS `roles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `role_name` varchar(50) NOT NULL,
  `description` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_name` (`role_name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default role
INSERT IGNORE INTO `roles` (`role_name`, `description`) VALUES
('admin', 'Administrator role with full system access');

-- Role permissions table
CREATE TABLE IF NOT EXISTS `role_permissions` (
  `role_id` int NOT NULL,
  `permission_id` int NOT NULL,
  PRIMARY KEY (`role_id`,`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Sell bill table
CREATE TABLE IF NOT EXISTS `sell_bill` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_broker` int DEFAULT NULL,
  `date_sell` date DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `id_user` int DEFAULT NULL,
  `path_pi` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  `is_batch_sell` tinyint(1) DEFAULT '0',
  `time_created` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Sell payments table
CREATE TABLE IF NOT EXISTS `sell_payments` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_sell_bill` int DEFAULT NULL,
  `amount_usd` decimal(10,2) DEFAULT NULL,
  `amount_da` decimal(10,2) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `date` date DEFAULT NULL,
  `path_swift` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Shipping lines table
CREATE TABLE IF NOT EXISTS `shipping_lines` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default shipping lines
INSERT IGNORE INTO `shipping_lines` (`name`) VALUES
('AKK'),
('CMA'),
('COSCO'),
('MSC'),
('MEARSK');

-- Suppliers table
CREATE TABLE IF NOT EXISTS `suppliers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `contact_info` text,
  `notes` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `supplier_name_unic` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Tasks table
CREATE TABLE IF NOT EXISTS `tasks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_user_create` int DEFAULT NULL,
  `id_user_receive` int DEFAULT NULL,
  `date_create` datetime DEFAULT NULL,
  `date_declare_done` datetime DEFAULT NULL,
  `date_confirm_done` datetime DEFAULT NULL,
  `id_user_confirm_done` int DEFAULT NULL,
  `id_priority` int DEFAULT NULL,
  `title` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `desciption` text COLLATE utf8mb4_general_ci,
  `notes` text COLLATE utf8mb4_general_ci,
  `is_task_for_car` tinyint(1) DEFAULT '0',
  `is_task_for_transfer` tinyint(1) DEFAULT '0',
  `is_task_for_supplier` tinyint(1) DEFAULT '0',
  `is_task_for_client` tinyint(1) DEFAULT '0',
  `is_task_for_user` tinyint(1) DEFAULT '0',
  `id_chat_grroup` int DEFAULT NULL,
  `assigned_users_ids` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `subject_ids` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  PRIMARY KEY (`id`),
  CONSTRAINT `tasks_chk_1` CHECK (json_valid(`assigned_users_ids`)),
  CONSTRAINT `tasks_chk_2` CHECK (json_valid(`subject_ids`))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Tracking table
CREATE TABLE IF NOT EXISTS `tracking` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `container_ref` varchar(255) DEFAULT NULL,
  `tracking` varchar(255) DEFAULT NULL,
  `time` timestamp NULL DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `container_ref` (`container_ref`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Transfer details table
CREATE TABLE IF NOT EXISTS `transfer_details` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_transfer` int DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `date` datetime DEFAULT NULL,
  `client_name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `client_mobile` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `description` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `notes` text COLLATE utf8mb4_general_ci,
  `id_client` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Users table
CREATE TABLE IF NOT EXISTS `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role_id` int NOT NULL,
  `max_unpayed_created_bills` int DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default admin user
INSERT IGNORE INTO `users` (`username`, `email`, `password`, `role_id`, `max_unpayed_created_bills`) VALUES
('admin', 'admin@example.com', '$2y$10$xk9Kh/WwDsRGOBvSQG4jO.Lwvwfnl1wXFmjCdvrD2ahCq5oGHYGSO', 1, 5);

-- Transfers table
CREATE TABLE IF NOT EXISTS `transfers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_user_do_transfer` int NOT NULL,
  `date_do_transfer` datetime NOT NULL,
  `amount_sending_da` decimal(10,2) NOT NULL,
  `rate` decimal(10,2) NOT NULL,
  `id_user_receive_transfer` int DEFAULT NULL,
  `amount_received_usd` decimal(10,2) DEFAULT '0.00',
  `date_receive` datetime DEFAULT NULL,
  `details_transfer` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_bin,
  `notes` text,
  `receiver_notes` text,
  `id_bank` int DEFAULT NULL,
  `ref_pi_transfer` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_user_do_transfer` (`id_user_do_transfer`),
  CONSTRAINT `transfers_ibfk_1` FOREIGN KEY (`id_user_do_transfer`) REFERENCES `users` (`id`),
  CONSTRAINT `transfers_chk_1` CHECK (json_valid(`details_transfer`))
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Transfers inter table
CREATE TABLE IF NOT EXISTS `transfers_inter` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `amount` decimal(10,2) DEFAULT NULL,
  `date_transfer` date DEFAULT NULL,
  `from_user_id` int DEFAULT NULL,
  `to_user_id` int DEFAULT NULL,
  `id_admin_confirm` int DEFAULT NULL,
  `notes` text,
  `date_received` date DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Versions table
CREATE TABLE IF NOT EXISTS `versions` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `version` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default version
INSERT IGNORE INTO `versions` (`id`, `version`) VALUES
(1, 19);

-- Warehouses table
CREATE TABLE IF NOT EXISTS `warehouses` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `warhouse_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- Add more CREATE TABLE statements below
-- ============================================


