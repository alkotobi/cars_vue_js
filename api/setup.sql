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
  `company_address` text COLLATE latin1_general_ci DEFAULT NULL COMMENT 'Company address for bank account',
  `notes` text COLLATE latin1_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

-- Brands table
CREATE TABLE IF NOT EXISTS `brands` (
  `id` int NOT NULL AUTO_INCREMENT,
  `brand` varchar(255) DEFAULT NULL,
  `logo_path` varchar(500) DEFAULT NULL COMMENT 'Path to brand logo image file',
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

-- Car name media table (photos and videos)
CREATE TABLE IF NOT EXISTS `car_name_media` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_name_id` int(11) NOT NULL COMMENT 'FK to cars_names',
  `file_path` varchar(500) NOT NULL COMMENT 'Storage path to media file',
  `file_name` varchar(255) NOT NULL COMMENT 'Original filename',
  `file_size` bigint(20) DEFAULT NULL COMMENT 'File size in bytes',
  `file_type` varchar(100) DEFAULT NULL COMMENT 'MIME type (image/* or video/*)',
  `media_type` enum('photo', 'video') NOT NULL COMMENT 'Type: photo or video',
  `uploaded_by` int(11) NOT NULL COMMENT 'FK to users - who uploaded',
  `uploaded_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=Deleted, 1=Active',
  PRIMARY KEY (`id`),
  KEY `idx_car_name_id` (`car_name_id`),
  KEY `idx_uploaded_by` (`uploaded_by`),
  KEY `idx_uploaded_at` (`uploaded_at`),
  KEY `idx_media_type` (`media_type`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_car_name_media_car_name` FOREIGN KEY (`car_name_id`) REFERENCES `cars_names` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_name_media_uploaded_by` FOREIGN KEY (`uploaded_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Cars stock table
CREATE TABLE IF NOT EXISTS `cars_stock` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
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
  `payment_confirmed` tinyint(1) DEFAULT 0 COMMENT 'Payment confirmed status - only admins or users with can_confirm_payment permission can set to true',
  `hidden_by_user_id` int DEFAULT NULL,
  `hidden_time_stamp` timestamp NULL DEFAULT NULL,
  `path_coo` varchar(255) DEFAULT NULL,
  `path_coc` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `vin` (`vin`),
  KEY `idx_payment_confirmed` (`payment_confirmed`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat groups table
CREATE TABLE IF NOT EXISTS `chat_groups` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `id_user_owner` int DEFAULT NULL,
  `id_client_owner` int DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_id_client_owner` (`id_client_owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Chat last read message table
CREATE TABLE IF NOT EXISTS `chat_last_read_message` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_group` int DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `id_client` int DEFAULT NULL,
  `id_last_read_message` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_group_user` (`id_group`,`id_user`),
  KEY `idx_group_client` (`id_group`,`id_client`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat messages table
CREATE TABLE IF NOT EXISTS `chat_messages` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_group` int DEFAULT NULL,
  `message_from_user_id` int DEFAULT NULL,
  `message_from_client_id` int DEFAULT NULL,
  `chat_replay_to_message_id` int DEFAULT NULL,
  `message` text COLLATE utf8mb4_general_ci,
  `time` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_message_from_client_id` (`message_from_client_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Chat read by table
CREATE TABLE IF NOT EXISTS `chat_read_by` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_chat_message` int DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  `id_client` int DEFAULT NULL,
  `read_at` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id_client` (`id_client`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Chat users table (supports both users and clients)
CREATE TABLE IF NOT EXISTS `chat_users` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_user` int DEFAULT NULL,
  `id_client` int DEFAULT NULL,
  `id_chat_group` int DEFAULT NULL,
  `is_active` tinyint(1) DEFAULT '1',
  PRIMARY KEY (`id`),
  KEY `idx_id_client` (`id_client`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Clients table
CREATE TABLE IF NOT EXISTS `clients` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `share_token` varchar(64) DEFAULT NULL,
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
  UNIQUE KEY `id_no` (`id_no`),
  UNIQUE KEY `idx_share_token` (`share_token`)
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
('can_hide_car', 'Can hide cars'),
('can_confirm_payment', 'Can confirm payment for sell bills and cars');

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
  `payment_confirmed` tinyint(1) DEFAULT 0 COMMENT 'Payment confirmed status - only admins or users with can_confirm_payment permission can set to true',
  `time_created` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_payment_confirmed` (`payment_confirmed`)
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
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `email` varchar(100) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role_id` int(11) NOT NULL,
  `max_unpayed_created_bills` int(11) DEFAULT '0',
  `is_diffrent_company` tinyint(1) DEFAULT 0 COMMENT 'Flag to indicate if user has different company assets',
  `path_logo` varchar(500) DEFAULT NULL COMMENT 'Path to company logo file',
  `path_letter_head` varchar(500) DEFAULT NULL COMMENT 'Path to letterhead file',
  `path_stamp` varchar(500) DEFAULT NULL COMMENT 'Path to stamp/gml2 file',
  `path_contract_terms` varchar(500) DEFAULT NULL COMMENT 'Path to contract terms JSON file',
  `id_bank_account` int unsigned DEFAULT NULL COMMENT 'Foreign key to banks table for user''s bank account',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`),
  KEY `idx_id_bank_account` (`id_bank_account`),
  CONSTRAINT `fk_users_bank_account` FOREIGN KEY (`id_bank_account`) REFERENCES `banks` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default admin user
INSERT IGNORE INTO `users` (`username`, `email`, `password`, `role_id`, `max_unpayed_created_bills`) VALUES
('admin', 'admin@example.com', '$2y$10$xk9Kh/WwDsRGOBvSQG4jO.Lwvwfnl1wXFmjCdvrD2ahCq5oGHYGSO', 1, 5);

-- Transfers table
CREATE TABLE IF NOT EXISTS `transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `id_user_do_transfer` int(11) NOT NULL,
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
-- Car Files Management System Tables
-- ============================================

-- Car file categories table
CREATE TABLE IF NOT EXISTS `car_file_categories` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category_name` varchar(255) NOT NULL,
  `importance_level` tinyint(1) NOT NULL DEFAULT 3 COMMENT '1=Critical, 2=High, 3=Medium, 4=Low, 5=Optional',
  `is_required` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1=Required, 0=Optional',
  `display_order` int(11) NOT NULL DEFAULT 0 COMMENT 'Order for UI display',
  `description` text DEFAULT NULL,
  `visibility_scope` enum('public', 'department', 'role', 'private') NOT NULL DEFAULT 'public' COMMENT 'Who can see files in this category',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `category_name` (`category_name`),
  KEY `idx_importance` (`importance_level`),
  KEY `idx_display_order` (`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default categories
INSERT IGNORE INTO `car_file_categories` (`category_name`, `importance_level`, `is_required`, `display_order`, `description`, `visibility_scope`) VALUES
('Bill of Lading', 1, 1, 1, 'Bill of Lading document', 'public'),
('Invoice', 2, 1, 2, 'Sell Invoice (PI)', 'public'),
('Packing List', 2, 1, 3, 'Buy Packing List (PI)', 'public'),
('Certificate of Origin', 3, 0, 4, 'COO Certificate', 'public'),
('Certificate of Conformity', 3, 0, 5, 'COC Certificate', 'public');

-- Car files table
CREATE TABLE IF NOT EXISTS `car_files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_id` int(11) NOT NULL COMMENT 'FK to cars_stock',
  `category_id` int(11) NOT NULL COMMENT 'FK to car_file_categories',
  `file_path` varchar(500) NOT NULL COMMENT 'Storage path to file',
  `file_name` varchar(255) NOT NULL COMMENT 'Original filename',
  `file_size` bigint(20) DEFAULT NULL COMMENT 'File size in bytes',
  `file_type` varchar(100) DEFAULT NULL COMMENT 'MIME type',
  `uploaded_by` int(11) NOT NULL COMMENT 'FK to users - who uploaded',
  `uploaded_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `version` int(11) NOT NULL DEFAULT 1 COMMENT 'File version number',
  `notes` text DEFAULT NULL COMMENT 'Additional notes about the file',
  `visibility_scope` enum('public', 'department', 'role', 'private') DEFAULT NULL COMMENT 'Override category visibility, NULL = use category default',
  `allowed_viewers` text DEFAULT NULL COMMENT 'JSON array of user_ids who can view, NULL = all',
  `department_id` int(11) DEFAULT NULL COMMENT 'FK to departments if visibility is department-based',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=Deleted/Archived, 1=Active',
  PRIMARY KEY (`id`),
  KEY `idx_car_id` (`car_id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_uploaded_by` (`uploaded_by`),
  KEY `idx_uploaded_at` (`uploaded_at`),
  KEY `idx_car_category` (`car_id`, `category_id`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_car_files_car` FOREIGN KEY (`car_id`) REFERENCES `cars_stock` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_files_category` FOREIGN KEY (`category_id`) REFERENCES `car_file_categories` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_car_files_uploaded_by` FOREIGN KEY (`uploaded_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Custom clearance agents table (Transiteurs)
CREATE TABLE IF NOT EXISTS `custom_clearance_agents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Agent name or company name',
  `contact_person` varchar(255) DEFAULT NULL COMMENT 'Contact person name',
  `phone` varchar(50) DEFAULT NULL COMMENT 'Phone number',
  `email` varchar(255) DEFAULT NULL COMMENT 'Email address',
  `address` text DEFAULT NULL COMMENT 'Physical address',
  `license_number` varchar(100) DEFAULT NULL COMMENT 'License or registration number',
  `notes` text DEFAULT NULL COMMENT 'Additional notes',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Car file physical tracking table
CREATE TABLE IF NOT EXISTS `car_file_physical_tracking` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `current_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who has physical copy now, NULL = available',
  `previous_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who had it before',
  `custom_clearance_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if checked out to agent',
  `checkout_type` enum('user', 'client', 'custom_clearance_agent') NOT NULL DEFAULT 'user' COMMENT 'Type of checkout: user, client, or custom clearance agent',
  `client_id` int(11) DEFAULT NULL COMMENT 'FK to clients - if checked out to client',
  `checked_out_at` timestamp NULL DEFAULT NULL COMMENT 'When physical copy was taken',
  `checked_in_at` timestamp NULL DEFAULT NULL COMMENT 'When returned, NULL if still out',
  `transfer_notes` text DEFAULT NULL COMMENT 'Notes about the transfer',
  `transferred_by` int(11) DEFAULT NULL COMMENT 'FK to users - who made the transfer',
  `transferred_at` timestamp NULL DEFAULT NULL COMMENT 'When transfer was made',
  `status` enum('available', 'checked_out', 'lost', 'archived') NOT NULL DEFAULT 'available',
  `is_visible_to_holder_only` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1=only holder+admin can see, 0=use file visibility rules',
  `expected_return_date` date DEFAULT NULL COMMENT 'Expected return date for checked out files',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_car_file_id` (`car_file_id`),
  KEY `idx_current_holder` (`current_holder_id`),
  KEY `idx_status` (`status`),
  KEY `idx_checked_out_at` (`checked_out_at`),
  KEY `idx_car_file_status` (`car_file_id`, `status`),
  KEY `idx_agent_id` (`custom_clearance_agent_id`),
  KEY `idx_client_id` (`client_id`),
  KEY `idx_checkout_type` (`checkout_type`),
  CONSTRAINT `fk_physical_tracking_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_physical_tracking_current_holder` FOREIGN KEY (`current_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_previous_holder` FOREIGN KEY (`previous_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_tracking_agent` FOREIGN KEY (`custom_clearance_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_tracking_client` FOREIGN KEY (`client_id`) REFERENCES `clients` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Car file transfers table (history)
CREATE TABLE IF NOT EXISTS `car_file_transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who transferred from, NULL = available',
  `to_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who received (NULL if transferred to agent or client)',
  `from_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if transferred from agent',
  `to_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if transferred to agent',
  `from_client_name` varchar(255) DEFAULT NULL COMMENT 'Client name if transferred from client',
  `to_client_name` varchar(255) DEFAULT NULL COMMENT 'Client name if transferred to client',
  `transfer_type` enum('user_to_user', 'user_to_agent', 'agent_to_user', 'user_to_client', 'client_to_user') NOT NULL DEFAULT 'user_to_user',
  `transfer_status` enum('pending', 'approved', 'rejected') NOT NULL DEFAULT 'approved' COMMENT 'Transfer status: pending (waiting approval), approved (completed), rejected (cancelled)',
  `transferred_by` int(11) NOT NULL COMMENT 'FK to users - who performed the transfer',
  `transferred_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `notes` text DEFAULT NULL COMMENT 'Transfer notes',
  `return_expected_date` date DEFAULT NULL COMMENT 'Expected return date',
  `returned_at` timestamp NULL DEFAULT NULL COMMENT 'When file was returned (checked in)',
  `return_notes` text DEFAULT NULL COMMENT 'Return notes',
  PRIMARY KEY (`id`),
  KEY `idx_car_file_id` (`car_file_id`),
  KEY `idx_from_user` (`from_user_id`),
  KEY `idx_to_user` (`to_user_id`),
  KEY `idx_transferred_by` (`transferred_by`),
  KEY `idx_transferred_at` (`transferred_at`),
  KEY `idx_returned_at` (`returned_at`),
  KEY `idx_from_agent` (`from_agent_id`),
  KEY `idx_to_agent` (`to_agent_id`),
  KEY `idx_transfer_type` (`transfer_type`),
  KEY `idx_transfer_status` (`transfer_status`),
  CONSTRAINT `fk_transfers_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_transfers_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_transfers_from_agent` FOREIGN KEY (`from_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_to_agent` FOREIGN KEY (`to_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- Car Selections System Tables
-- ============================================

-- Teams table (must be created before jobs since jobs references teams)
CREATE TABLE IF NOT EXISTS `teams` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Team name',
  `team_leader_id` int(11) NOT NULL COMMENT 'FK to users - team leader',
  `description` text DEFAULT NULL COMMENT 'Team description',
  `jobs_completed_count` int(11) NOT NULL DEFAULT 0 COMMENT 'Incremental count of completed jobs',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_team_leader_id` (`team_leader_id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_jobs_completed_count` (`jobs_completed_count`),
  CONSTRAINT `fk_teams_team_leader` FOREIGN KEY (`team_leader_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Teams with team leaders';

-- Team members table (users can only be in one team)
CREATE TABLE IF NOT EXISTS `team_members` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `team_id` int(11) NOT NULL COMMENT 'FK to teams',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - UNIQUE constraint ensures one team per user',
  `role` enum('member', 'deputy_leader') NOT NULL DEFAULT 'member' COMMENT 'Role in the team',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When user joined the team',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_team` (`user_id`),
  KEY `idx_team_id` (`team_id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_team_members_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_team_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Team members - users can only be in one team at a time';

-- Jobs table (must be created after teams since jobs references teams)
CREATE TABLE IF NOT EXISTS `jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Job name',
  `description` text DEFAULT NULL COMMENT 'Job description',
  `category` varchar(100) DEFAULT NULL COMMENT 'Job category (loading, delivery, inspection, documentation, etc.)',
  `estimated_duration_hours` int(11) DEFAULT NULL COMMENT 'Estimated duration in hours',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `team_id` int(11) DEFAULT NULL COMMENT 'FK to teams - each job belongs to one team (one team can have multiple jobs)',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_category` (`category`),
  KEY `idx_name` (`name`),
  KEY `idx_team_id` (`team_id`),
  CONSTRAINT `fk_jobs_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Jobs assigned to teams - one team can have multiple jobs, each job belongs to one team';

-- Insert default jobs
INSERT IGNORE INTO `jobs` (`name`, `description`, `category`, `is_active`) VALUES
('Loading', 'Load cars onto container/ship', 'loading', 1),
('Delivery', 'Deliver cars to destination', 'delivery', 1),
('Inspection', 'Inspect cars for quality/condition', 'inspection', 1),
('Documentation', 'Prepare and process documents', 'documentation', 1),
('Warehouse Management', 'Manage warehouse operations', 'warehouse', 1),
('Custom Clearance', 'Handle custom clearance procedures', 'custom_clearance', 1);

-- Car selections table
CREATE TABLE IF NOT EXISTS `car_selections` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Selection name',
  `description` text DEFAULT NULL COMMENT 'Selection description',
  `user_create_selection` int(11) NOT NULL COMMENT 'FK to users - who created the selection',
  `selection_data` json DEFAULT NULL COMMENT 'JSON array of car IDs that were selected',
  `assigned_to_team` int(11) DEFAULT NULL COMMENT 'FK to teams - team assigned to work on this',
  `assigned_to_team_from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who assigned to team',
  `assigned_at` datetime DEFAULT NULL COMMENT 'When assigned to team',
  `job_id` int(11) DEFAULT NULL COMMENT 'FK to jobs - what job the team needs to do',
  `job_done_on` datetime DEFAULT NULL COMMENT 'DateTime when job was completed',
  `owned_by` json DEFAULT NULL COMMENT 'JSON array of user IDs - users who received this selection',
  `sent_by_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who sent the selection',
  `status` enum('pending', 'in_progress', 'completed', 'cancelled') NOT NULL DEFAULT 'pending' COMMENT 'Selection status',
  `status_changed_at` datetime DEFAULT NULL COMMENT 'When the status was last changed',
  `previous_status` enum('pending', 'in_progress', 'completed', 'cancelled') DEFAULT NULL COMMENT 'Previous status before change',
  `priority` enum('low', 'medium', 'high', 'urgent') NOT NULL DEFAULT 'medium' COMMENT 'Priority level',
  `due_date` datetime DEFAULT NULL COMMENT 'Optional deadline',
  `deadline` datetime DEFAULT NULL COMMENT 'Deadline for completing the selection',
  `notes` text DEFAULT NULL COMMENT 'Additional notes/comments',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_create_selection` (`user_create_selection`),
  KEY `idx_assigned_to_team` (`assigned_to_team`),
  KEY `idx_assigned_to_team_from_user_id` (`assigned_to_team_from_user_id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_sent_by_user_id` (`sent_by_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_status_changed_at` (`status_changed_at`),
  KEY `idx_priority` (`priority`),
  KEY `idx_due_date` (`due_date`),
  KEY `idx_deadline` (`deadline`),
  KEY `idx_assigned_at` (`assigned_at`),
  KEY `idx_job_done_on` (`job_done_on`),
  CONSTRAINT `fk_car_selections_creator` FOREIGN KEY (`user_create_selection`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_team` FOREIGN KEY (`assigned_to_team`) REFERENCES `teams` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_assigner` FOREIGN KEY (`assigned_to_team_from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_sender` FOREIGN KEY (`sent_by_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Car selections with team assignments and job tracking';

-- Selection ownership history table
CREATE TABLE IF NOT EXISTS `selection_ownership_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `selection_id` int(11) NOT NULL COMMENT 'FK to car_selections',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - user involved in the action',
  `action` enum('sent', 'received', 'transferred') NOT NULL COMMENT 'Type of action',
  `from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who sent (nullable)',
  `to_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who received (nullable)',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_selection_id` (`selection_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_to_user_id` (`to_user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_ownership_history_selection` FOREIGN KEY (`selection_id`) REFERENCES `car_selections` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Track ownership changes and transfers of selections';

-- Selection comments table
CREATE TABLE IF NOT EXISTS `selection_comments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `selection_id` int(11) NOT NULL COMMENT 'FK to car_selections',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - who commented',
  `comment` text NOT NULL COMMENT 'Comment text',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_selection_id` (`selection_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_comments_selection` FOREIGN KEY (`selection_id`) REFERENCES `car_selections` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Comments and updates on selections';

-- ============================================
-- Trigger to increment jobs_completed_count
-- ============================================
DELIMITER $$

DROP TRIGGER IF EXISTS trg_increment_team_jobs_completed$$

CREATE TRIGGER trg_increment_team_jobs_completed
AFTER UPDATE ON `car_selections`
FOR EACH ROW
BEGIN
  -- Only increment if status changed to 'completed' and job_done_on is set
  IF NEW.status = 'completed' 
     AND OLD.status != 'completed' 
     AND NEW.job_done_on IS NOT NULL
     AND NEW.assigned_to_team IS NOT NULL THEN
    UPDATE `teams`
    SET `jobs_completed_count` = `jobs_completed_count` + 1
    WHERE `id` = NEW.assigned_to_team;
  END IF;
END$$

DELIMITER ;

-- ============================================
-- Trigger to auto-generate share_token for new clients
-- ============================================
DELIMITER $$

DROP TRIGGER IF EXISTS `generate_client_share_token`$$

CREATE TRIGGER `generate_client_share_token` 
BEFORE INSERT ON `clients`
FOR EACH ROW
BEGIN
  IF NEW.share_token IS NULL OR NEW.share_token = '' THEN
    SET NEW.share_token = SUBSTRING(
      SHA2(CONCAT(NEW.id, NEW.name, COALESCE(NEW.id_no, ''), NOW(), RAND()), 256), 
      1, 
      64
    );
  END IF;
END$$

DELIMITER ;

-- ============================================
-- Add more CREATE TABLE statements below
-- ============================================


