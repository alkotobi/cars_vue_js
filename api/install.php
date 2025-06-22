<?php
header('Content-Type: application/json');
require_once 'config.php';

try {
    $pdo = new PDO(
        "mysql:host=$db_host;charset=utf8",
        $db_user,
        $db_pass
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    // Create database if not exists
    $pdo->exec("CREATE DATABASE IF NOT EXISTS $db_name CHARACTER SET utf8 COLLATE utf8_general_ci");
    $pdo->exec("USE $db_name");

    // Create banks table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `banks` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `company_name` varchar(255) DEFAULT NULL,
        `bank_name` varchar(255) DEFAULT NULL,
        `swift_code` varchar(255) DEFAULT NULL,
        `bank_account` varchar(255) DEFAULT NULL,
        `bank_address` varchar(255) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci");

    // Create brands table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `brands` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `brand` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `brand` (`brand`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create suppliers table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `suppliers` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `name` varchar(255) NOT NULL,
        `address` varchar(255) DEFAULT NULL,
        `email` varchar(255) DEFAULT NULL,
        `mobiles` varchar(255) NOT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `supplier_name_unic` (`name`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create cars_names table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `cars_names` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `car_name` varchar(255) DEFAULT NULL,
        `id_brand` int(11) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `is_big_car` tinyint(1) DEFAULT 0,
        PRIMARY KEY (`id`),
        UNIQUE KEY `car_name` (`car_name`),
        KEY `id_brand` (`id_brand`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create colors table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `colors` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `color` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `color` (`color`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create discharge_ports table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `discharge_ports` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `discharge_port` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `discharge_port` (`discharge_port`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create loading_ports table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `loading_ports` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `loading_port` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `loading_port` (`loading_port`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create buy_bill table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `buy_bill` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_supplier` int(11) DEFAULT NULL,
        `date_buy` datetime DEFAULT NULL,
        `amount` decimal(10,2) DEFAULT NULL,
        `payed` decimal(10,2) DEFAULT NULL,
        `pi_path` varchar(255) DEFAULT NULL,
        `bill_ref` varchar(255) DEFAULT NULL,
        `is_stock_updated` tinyint(1) DEFAULT NULL,
        `is_ordered` tinyint(1) DEFAULT 1,
        PRIMARY KEY (`id`),
        KEY `id_supplier` (`id_supplier`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");


    $pdo->exec("CREATE TABLE `adv_sql` (
    `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `stmt` text DEFAULT NULL,
    `desc` varchar(255) DEFAULT NULL,
    `params` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`params`)),
    `param_values` text CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT NULL CHECK (json_valid(`param_values`)),
    PRIMARY KEY (`id`)
  ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;");
    // Create buy_details table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `buy_details` (
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
        KEY `id_buy_bill` (`id_buy_bill`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create buy_payments table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `buy_payments` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_buy_bill` int(11) DEFAULT NULL,
        `date_payment` datetime DEFAULT NULL,
        `amount` decimal(10,2) DEFAULT NULL,
        `swift_path` varchar(255) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `id_user` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_buy_bill` (`id_buy_bill`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create clients table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `clients` (
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
        UNIQUE KEY `client_name_unic` (`name`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create cars_stock table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `cars_stock` (
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
        PRIMARY KEY (`id`),
        KEY `id_client` (`id_client`),
        KEY `id_port_loading` (`id_port_loading`),
        KEY `id_port_discharge` (`id_port_discharge`),
        KEY `id_buy_details` (`id_buy_details`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create roles table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `roles` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `role_name` varchar(255) NOT NULL,
        `description` text DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `role_name` (`role_name`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create users table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `users` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `username` varchar(255) NOT NULL,
        `email` varchar(255) DEFAULT NULL,
        `password` varchar(255) NOT NULL,
        `role_id` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `username` (`username`),
        KEY `role_id` (`role_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create permissions table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `permissions` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `permission_name` varchar(255) NOT NULL,
        `description` text DEFAULT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `permission_name` (`permission_name`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create rates table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `rates` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `rate` decimal(10,2) DEFAULT NULL,
        `created_on` timestamp NULL DEFAULT NULL,
        `id_user` int DEFAULT NULL,
        `notes` text,
        PRIMARY KEY (`id`),
        KEY `id_user` (`id_user`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create role_permissions table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `role_permissions` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `role_id` int(11) NOT NULL,
        `permission_id` int(11) NOT NULL,
        PRIMARY KEY (`id`),
        UNIQUE KEY `role_permission_unique` (`role_id`,`permission_id`),
        KEY `permission_id` (`permission_id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create defaults table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `defaults` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `rate` decimal(10,2) DEFAULT NULL,
        `freight_small` decimal(10,2) DEFAULT NULL,
        `freight_big` decimal(10,2) DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_general_ci");

    // Create warehouses table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `warehouses` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `name` varchar(255) NOT NULL,
        `address` text DEFAULT NULL,
        `notes` text DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create sell_bill table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `sell_bill` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_client` int(11) DEFAULT NULL,
        `date_sell` datetime DEFAULT NULL,
        `amount` decimal(10,2) DEFAULT NULL,
        `received` decimal(10,2) DEFAULT NULL,
        `bill_ref` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_client` (`id_client`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create sell_payments table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `sell_payments` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_sell_bill` int(11) DEFAULT NULL,
        `date_payment` datetime DEFAULT NULL,
        `amount` decimal(10,2) DEFAULT NULL,
        `swift_path` varchar(255) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `id_user` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_sell_bill` (`id_sell_bill`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create transfers table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `transfers` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_user_do_transfer` int(11) DEFAULT NULL,
        `id_user_receive_transfer` int(11) DEFAULT NULL,
        `date_do_transfer` datetime DEFAULT NULL,
        `date_receive` datetime DEFAULT NULL,
        `amount_sending_da` decimal(10,2) DEFAULT NULL,
        `rate` decimal(10,2) DEFAULT NULL,
        `amount_received_usd` decimal(10,2) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `id_bank` int(11) DEFAULT NULL,
        `ref_pi_transfer` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_user_do_transfer` (`id_user_do_transfer`),
        KEY `id_user_receive_transfer` (`id_user_receive_transfer`),
        KEY `id_bank` (`id_bank`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create transfer_details table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `transfer_details` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_transfer` int(11) DEFAULT NULL,
        `amount` decimal(10,2) DEFAULT NULL,
        `date` datetime DEFAULT NULL,
        `client_name` varchar(255) DEFAULT NULL,
        `client_mobile` varchar(255) DEFAULT NULL,
        `rate` decimal(10,2) DEFAULT NULL,
        `description` text DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `id_client` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_transfer` (`id_transfer`),
        KEY `id_client` (`id_client`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Create transfers_inter table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `transfers_inter` (
        `id` int(11) NOT NULL AUTO_INCREMENT,
        `id_user_do_transfer` int(11) DEFAULT NULL,
        `id_user_receive_transfer` int(11) DEFAULT NULL,
        `date_do_transfer` datetime DEFAULT NULL,
        `date_receive` datetime DEFAULT NULL,
        `amount_sending_da` decimal(10,2) DEFAULT NULL,
        `rate` decimal(10,2) DEFAULT NULL,
        `amount_received_usd` decimal(10,2) DEFAULT NULL,
        `notes` text DEFAULT NULL,
        `id_bank` int(11) DEFAULT NULL,
        `ref_pi_transfer` varchar(255) DEFAULT NULL,
        PRIMARY KEY (`id`),
        KEY `id_user_do_transfer` (`id_user_do_transfer`),
        KEY `id_user_receive_transfer` (`id_user_receive_transfer`),
        KEY `id_bank` (`id_bank`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    // Insert default roles if not exist
    $stmt = $pdo->prepare("INSERT IGNORE INTO roles (role_name, description) VALUES 
        ('admin', 'Administrator with full access'),
        ('user', 'Regular user with limited access')
    ");
    $stmt->execute();

    // Insert default permissions if not exist
    $stmt = $pdo->prepare("INSERT IGNORE INTO permissions (permission_name, description) VALUES 
        ('can_manage_users', 'Can create, edit, and delete users'),
        ('can_manage_roles', 'Can create, edit, and delete roles'),
        ('can_manage_permissions', 'Can manage role permissions'),
        ('is_exchange_sender', 'Can send exchange requests'),
        ('is_exchange_receiver', 'Can receive exchange requests'),
        ('can_manage_cars', 'Can manage cars inventory'),
        ('can_edit_cars_prop', 'Can edit car properties like brand, model, year, etc.'),
        ('can_edit_vin', 'Can edit vehicle identification number'),
        ('can_edit_car_client_name', 'Can modify client information for cars'),
        ('can_edit_cars_sell_price', 'Can modify selling price of cars'),
        ('can_edit_cars_sell_rate', 'Can modify exchange rates for car sales'),
        ('can_edit_cars_discharge_port', 'Can modify discharge port information'),
        ('can_upload_car_files', 'Can upload documents and files related to cars'),
        ('can_edit_cars_ports', 'Can modify loading and discharge ports for cars'),
        ('can_edit_car_money', 'Can modify financial information related to cars'),
        ('can_receive_car', 'Can mark cars as received in warehouse'),
        ('can_edit_car_documents', 'Can modify and manage car documentation'),
        ('can_load_car', 'Can manage car loading operations'),
        ('can_edit_sell_payments', 'Can modify payment information for sales'),
        ('can_delete_sell_paymets', 'Can delete payment records for sales'),
        ('can_c_sell_payments', 'Can create new payment records for sales'),
        ('can_delete_sell_bill', 'Can delete sales invoices'),
        ('can_edit_sell_bill', 'Can modify sales invoice information'),
        ('can_access_cashier', 'Can access cashier functions and operations'),
        ('can_purchase_cars', 'Can create and manage purchase orders for cars'),
        ('can_sell_cars', 'Can create and manage sales invoices for cars'),
        ('can_c_car_stock', 'Can view and manage car stock inventory'),
        ('can_assign_to_tmp_clients', 'Can assign cars to temporary clients')
    ");
    $stmt->execute();

    // Create default admin user if not exists
    $defaultPassword = password_hash('123', PASSWORD_DEFAULT);
    $stmt = $pdo->prepare("INSERT IGNORE INTO users (username, email, password, role_id) VALUES 
        ('admin', 'admin@example.com', ?, (SELECT id FROM roles WHERE role_name = 'admin'))
    ");
    $stmt->execute([$defaultPassword]);

    // Assign all permissions to admin role
    $stmt = $pdo->prepare("INSERT IGNORE INTO role_permissions (role_id, permission_id)
        SELECT r.id, p.id 
        FROM roles r 
        CROSS JOIN permissions p 
        WHERE r.role_name = 'admin'
    ");
    $stmt->execute();

    echo json_encode(['success' => true, 'message' => 'Database and tables created successfully']);
} catch (PDOException $e) {
    echo json_encode(['success' => false, 'error' => $e->getMessage()]);
}
