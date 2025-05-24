CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `password` varchar(255) COLLATE utf8mb4_general_ci NOT NULL,
  `email` varchar(100) COLLATE utf8mb4_general_ci NOT NULL,
  `role_id` int DEFAULT NULL,
  `status` enum('active','inactive') COLLATE utf8mb4_general_ci DEFAULT 'active',
  `is_broker` boolean DEFAULT false,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`),
  UNIQUE KEY `email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Insert admin user with bcrypt hashed password for '123'
INSERT INTO `users` (`username`, `password`, `email`, `role_id`, `status`) VALUES
('admin', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin@example.com', 1, 'active');

-- Create roles table
CREATE TABLE `roles` (
  `id` int NOT NULL AUTO_INCREMENT,
  `role_name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `role_name` (`role_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Insert admin and guest roles
INSERT INTO `roles` (`role_name`, `description`) VALUES
('admin', 'Administrator with full system access'),
('guest', 'Guest user with limited access');

CREATE TABLE `permissions` (
  `id` int NOT NULL AUTO_INCREMENT,
  `permission_name` varchar(50) COLLATE utf8mb4_general_ci NOT NULL,
  `description` text COLLATE utf8mb4_general_ci,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `permission_name` (`permission_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Insert permissions
INSERT INTO `permissions` (`permission_name`, `description`) VALUES
('can_manage_users', 'Can create, edit, and delete user accounts'),
('can_manage_roles', 'Can create, edit, and delete roles'),
('can_manage_permissions', 'Can manage role permissions'),
('is_exchange_sender', 'Can send exchange requests'),
('is_exchange_receiver', 'Can receive exchange requests'),
('can_manage_cars', 'Can manage cars inventory'),
('can_edit_cars_prop', ''),
('can_edit_vin', ''),
('can_edit_car_client_name', ''),
('can_edit_cars_sell_price', ''),
('can_edit_cars_sell_rate', ''),
('can_edit_cars_discharge_port', ''),
('can_receive_car', 'Can mark cars as received in warehouse'),
('can_edit_car_documents', 'Can manage car documents and update document dates'),
('can_load_car', 'Can mark cars as loaded and set loading date');

CREATE TABLE `role_permissions` (
  `role_id` int NOT NULL,
  `permission_id` int NOT NULL,
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`role_id`,`permission_id`),
  KEY `permission_id` (`permission_id`),
  CONSTRAINT `role_permissions_ibfk_1` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `role_permissions_ibfk_2` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create warehouses table
CREATE TABLE `warehouses` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `warhouse_name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `notes` text COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create cars_stock table
CREATE TABLE `cars_stock` (
  `id` int NOT NULL AUTO_INCREMENT,
  `notes` text,
  `id_buy_details` int DEFAULT NULL,
  `date_sell` datetime DEFAULT NULL,
  `id_client` int DEFAULT NULL,
  `price_cell` decimal(10,2) DEFAULT NULL,
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
  `rate` decimal(10,0) DEFAULT NULL,
  `date_pay_freight` date DEFAULT NULL,
  `date_get_bl` date DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_client` (`id_client`),
  KEY `id_port_loading` (`id_port_loading`),
  KEY `id_port_discharge` (`id_port_discharge`),
  KEY `id_buy_details` (`id_buy_details`),
  KEY `id_warehouse` (`id_warehouse`),
  CONSTRAINT `cars_stock_ibfk_1` FOREIGN KEY (`id_client`) REFERENCES `clients` (`id`),
  CONSTRAINT `cars_stock_ibfk_2` FOREIGN KEY (`id_port_loading`) REFERENCES `loading_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_3` FOREIGN KEY (`id_port_discharge`) REFERENCES `discharge_ports` (`id`),
  CONSTRAINT `cars_stock_ibfk_4` FOREIGN KEY (`id_buy_details`) REFERENCES `buy_details` (`id`),
  CONSTRAINT `cars_stock_ibfk_5` FOREIGN KEY (`id_warehouse`) REFERENCES `warehouses` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create brands table
CREATE TABLE `brands` (
  `id` int NOT NULL AUTO_INCREMENT,
  `brand` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `brand` (`brand`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create colors table
CREATE TABLE `colors` (
  `id` int NOT NULL AUTO_INCREMENT,
  `color` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `color` (`color`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create discharge ports table
CREATE TABLE `discharge_ports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `discharge_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discharge_port` (`discharge_port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create loading ports table
CREATE TABLE `loading_ports` (
  `id` int NOT NULL AUTO_INCREMENT,
  `loading_port` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `loading_port` (`loading_port`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create clients table
CREATE TABLE `clients` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL,
  `address` varchar(255) DEFAULT NULL,
  `email` varchar(255) DEFAULT NULL,
  `mobiles` varchar(255) NOT NULL,
  `id_copy_path` varchar(255) DEFAULT NULL,
  `id_no` varchar(255) DEFAULT NULL,
  `is_broker` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create suppliers table
CREATE TABLE `suppliers` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(255) DEFAULT NULL,
  `contact_info` text,
  `notes` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create cars_names table
CREATE TABLE `cars_names` (
  `id` int NOT NULL AUTO_INCREMENT,
  `car_name` varchar(255) DEFAULT NULL,
  `id_brand` int DEFAULT NULL,
  `notes` text,
  `is_big_car` tinyint(1) DEFAULT '0',
  PRIMARY KEY (`id`),
  UNIQUE KEY `car_name` (`car_name`),
  KEY `id_brand` (`id_brand`),
  CONSTRAINT `cars_names_ibfk_1` FOREIGN KEY (`id_brand`) REFERENCES `brands` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create buy_bill table
CREATE TABLE `buy_bill` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_supplier` int DEFAULT NULL,
  `date_buy` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `payed` decimal(10,2) DEFAULT NULL,
  `pi_path` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_supplier` (`id_supplier`),
  CONSTRAINT `buy_bill_ibfk_1` FOREIGN KEY (`id_supplier`) REFERENCES `suppliers` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create buy_details table
CREATE TABLE `buy_details` (
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
  PRIMARY KEY (`id`),
  KEY `id_car_name` (`id_car_name`),
  KEY `id_color` (`id_color`),
  KEY `id_buy_bill` (`id_buy_bill`),
  CONSTRAINT `buy_details_ibfk_1` FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`),
  CONSTRAINT `buy_details_ibfk_2` FOREIGN KEY (`id_color`) REFERENCES `colors` (`id`),
  CONSTRAINT `buy_details_ibfk_3` FOREIGN KEY (`id_buy_bill`) REFERENCES `buy_bill` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create buy_payments table
CREATE TABLE `buy_payments` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_buy_bill` int DEFAULT NULL,
  `date_payment` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `swift_path` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `id_buy_bill` (`id_buy_bill`),
  CONSTRAINT `buy_payments_ibfk_1` FOREIGN KEY (`id_buy_bill`) REFERENCES `buy_bill` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create sell_bill table
CREATE TABLE `sell_bill` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `id_broker` int(11) DEFAULT NULL,
  `date_sell` date DEFAULT NULL,
  `notes` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_user` int(11) DEFAULT NULL,
  `path_pi` varchar(255) DEFAULT NULL,
  `bill_ref` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create sell_payments table
CREATE TABLE `sell_payments` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_sell_bill` int DEFAULT NULL,
  `amount-usd` decimal(10,2) DEFAULT NULL,
  `amount_da` decimal(10,2) DEFAULT NULL,
  `rate` decimal(10,2) DEFAULT NULL,
  `date` date DEFAULT NULL,
  `path_swift` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
  `id_user` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Create banks table
CREATE TABLE `banks` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `company_name` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_name` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `swift_code` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_account` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `bank_address` varchar(255) COLLATE latin1_general_ci DEFAULT NULL,
  `notes` text COLLATE latin1_general_ci,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci;

-- Insert default bank
INSERT INTO `banks` (`company_name`, `bank_name`, `swift_code`, `bank_account`, `bank_address`) VALUES
('GROUP MERHAB LIMITED', 'GROUP MERHAB BANK', 'MERHAB123', '123456789', 'GUANGZHOU, CHINA');