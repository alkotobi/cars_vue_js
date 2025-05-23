<?php
header('Content-Type: application/json');
require_once 'config.php';

try {
    $pdo = new PDO(
        "mysql:host=$db_host;charset=utf8mb4",
        $db_user,
        $db_pass
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    // Create database if not exists
    $pdo->exec("CREATE DATABASE IF NOT EXISTS $db_name");
    $pdo->exec("USE $db_name");

    // Create roles table
    $pdo->exec("CREATE TABLE IF NOT EXISTS roles (
        id INT PRIMARY KEY AUTO_INCREMENT,
        role_name VARCHAR(50) NOT NULL UNIQUE,
        description TEXT
    )");

    // Create permissions table
    $pdo->exec("CREATE TABLE IF NOT EXISTS permissions (
        id INT PRIMARY KEY AUTO_INCREMENT,
        permission_name VARCHAR(50) NOT NULL UNIQUE,
        description TEXT
    )");

    // Create users table
    $pdo->exec("CREATE TABLE IF NOT EXISTS users (
        id INT PRIMARY KEY AUTO_INCREMENT,
        username VARCHAR(50) NOT NULL UNIQUE,
        email VARCHAR(100) NOT NULL UNIQUE,
        password VARCHAR(255) NOT NULL,
        role_id INT,
        FOREIGN KEY (role_id) REFERENCES roles(id)
    )");

    // Create role_permissions table
    $pdo->exec("CREATE TABLE IF NOT EXISTS role_permissions (
        role_id INT,
        permission_id INT,
        PRIMARY KEY (role_id, permission_id),
        FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
        FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
    )");

    // Create transfers table
    $pdo->exec("CREATE TABLE IF NOT EXISTS transfers (
        id INT PRIMARY KEY AUTO_INCREMENT,
        id_user_do_transfer INT NOT NULL,
        date_do_transfer DATETIME NOT NULL,
        amount_sending_da DECIMAL(10,2) NOT NULL,
        rate DECIMAL(10,2) NOT NULL,
        id_user_receive_transfer INT ,
        amount_received_usd DECIMAL(10,2) default 0.00,
        date_receive DATETIME,
        details_transfer JSON,
        notes TEXT,
        receiver_notes TEXT,
        FOREIGN KEY (id_user_do_transfer) REFERENCES users(id)
    )");

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
        ('can_edit_cars_prop',''),
        ('can_edit_vin',''),
        ('can_edit_car_client_name',''),
        ('can_edit_cars_sell_price',''),
        ('can_edit_cars_sell_rate',''),
        ('can_edit_cars_discharge_port','')
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

    // Create brands table
    $pdo->exec("CREATE TABLE IF NOT EXISTS brands (
        id INT PRIMARY KEY AUTO_INCREMENT,
        brand VARCHAR(255) DEFAULT NULL,
        UNIQUE KEY brand (brand)
    )");

    // Create colors table
    $pdo->exec("CREATE TABLE IF NOT EXISTS colors (
        id INT PRIMARY KEY AUTO_INCREMENT,
        color VARCHAR(255) DEFAULT NULL,
        UNIQUE KEY color (color)
    )");

    // Create discharge ports table
    $pdo->exec("CREATE TABLE IF NOT EXISTS discharge_ports (
        id INT PRIMARY KEY AUTO_INCREMENT,
        discharge_port VARCHAR(255) DEFAULT NULL,
        UNIQUE KEY discharge_port (discharge_port)
    )");

    // Create loading ports table
    $pdo->exec("CREATE TABLE IF NOT EXISTS loading_ports (
        id INT PRIMARY KEY AUTO_INCREMENT,
        loading_port VARCHAR(255) DEFAULT NULL,
        UNIQUE KEY loading_port (loading_port)
    )");

    // Create clients table
    $pdo->exec("CREATE TABLE IF NOT EXISTS clients (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) NOT NULL,
        address VARCHAR(255) DEFAULT NULL,
        email VARCHAR(255) DEFAULT NULL,
        mobiles VARCHAR(255) NOT NULL,
        id_copy_path VARCHAR(255) DEFAULT NULL,
        id_no VARCHAR(255) DEFAULT NULL,
        is_broker TINYINT(1) DEFAULT 0
    )");

    // Create suppliers table
    $pdo->exec("CREATE TABLE IF NOT EXISTS suppliers (
        id INT PRIMARY KEY AUTO_INCREMENT,
        name VARCHAR(255) DEFAULT NULL,
        contact_info TEXT DEFAULT NULL,
        notes TEXT DEFAULT NULL
    )");

    // Create cars_names table
    $pdo->exec("CREATE TABLE IF NOT EXISTS cars_names (
        id INT NOT NULL AUTO_INCREMENT,
        car_name VARCHAR(255) DEFAULT NULL,
        id_brand INT DEFAULT NULL,
        notes TEXT,
        is_big_car TINYINT(1) DEFAULT '0',
        PRIMARY KEY (id),
        UNIQUE KEY car_name (car_name),
        FOREIGN KEY (id_brand) REFERENCES brands(id)
    )");

    // Create buy_bill table
    $pdo->exec("CREATE TABLE IF NOT EXISTS buy_bill (
        id INT PRIMARY KEY AUTO_INCREMENT,
        id_supplier INT DEFAULT NULL,
        date_buy DATETIME DEFAULT NULL,
        amount DECIMAL(10,2) DEFAULT NULL,
        payed DECIMAL(10,2) DEFAULT NULL,
        pi_path VARCHAR(255) DEFAULT NULL,
        FOREIGN KEY (id_supplier) REFERENCES suppliers(id)
    )");

    // Create buy_details table
    $pdo->exec("CREATE TABLE IF NOT EXISTS buy_details (
        id INT PRIMARY KEY AUTO_INCREMENT,
        id_car_name INT DEFAULT NULL,
        id_color INT DEFAULT NULL,
        amount DECIMAL(10,2) DEFAULT NULL,
        notes TEXT DEFAULT NULL,
        QTY INT DEFAULT NULL,
        year INT DEFAULT NULL,
        month INT DEFAULT NULL,
        is_used_car TINYINT(1) DEFAULT 0,
        id_buy_bill INT DEFAULT NULL,
        price_sell DECIMAL(10,2) DEFAULT NULL,
        FOREIGN KEY (id_car_name) REFERENCES cars_names(id),
        FOREIGN KEY (id_color) REFERENCES colors(id),
        FOREIGN KEY (id_buy_bill) REFERENCES buy_bill(id)
    )");

    // Create warhouse table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `warehouses` ( 
       `id` int unsigned NOT NULL AUTO_INCREMENT, 
       `warhouse_name` varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL, 
       `notes` text COLLATE utf8mb4_general_ci, 
       PRIMARY KEY (`id`) 
     );"); 

    // Update cars_stock table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `cars_stock` ( 
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
        `rate` decimal(10,2) DEFAULT NULL,
       PRIMARY KEY (`id`), 
       KEY `id_client` (`id_client`), 
       KEY `id_port_loading` (`id_port_loading`), 
       KEY `id_port_discharge` (`id_port_discharge`), 
       KEY `id_buy_details` (`id_buy_details`), 
       CONSTRAINT `cars_stock_ibfk_1` FOREIGN KEY (`id_client`) REFERENCES `clients` (`id`), 
       CONSTRAINT `cars_stock_ibfk_2` FOREIGN KEY (`id_port_loading`) REFERENCES `loading_ports` (`id`), 
       CONSTRAINT `cars_stock_ibfk_3` FOREIGN KEY (`id_port_discharge`) REFERENCES `discharge_ports` (`id`), 
       CONSTRAINT `cars_stock_ibfk_4` FOREIGN KEY (`id_buy_details`) REFERENCES `buy_details` (`id`) 
     );"); 

    // Create buy_payments table
    $pdo->exec("CREATE TABLE IF NOT EXISTS buy_payments (
        id INT PRIMARY KEY AUTO_INCREMENT,
        id_buy_bill INT DEFAULT NULL,
        date_payment DATETIME DEFAULT NULL,
        amount DECIMAL(10,2) DEFAULT NULL,
        swift_path VARCHAR(255) DEFAULT NULL,
        FOREIGN KEY (id_buy_bill) REFERENCES buy_bill(id)
    )");

    // Create sell_bill table
    $pdo->exec("CREATE TABLE IF NOT EXISTS sell_bill (
        id INT UNSIGNED NOT NULL AUTO_INCREMENT,
        id_broker INT DEFAULT NULL,
        date_sell DATE DEFAULT NULL,
        notes TEXT COLLATE utf8mb4_general_ci,
        id_user INT UNSIGNED ,
        path_pi VARCHAR(255) DEFAULT NULL,
        PRIMARY KEY (id)
    )");

    // Create sell_payments table
    $pdo->exec("CREATE TABLE IF NOT EXISTS sell_payments (
        id INT UNSIGNED NOT NULL AUTO_INCREMENT,
        id_sell_bill INT DEFAULT NULL,
        `amount-usd` DECIMAL(10,2) DEFAULT NULL,
        amount_da DECIMAL(10,2) DEFAULT NULL,
        rate DECIMAL(10,2) DEFAULT NULL,
        date DATE DEFAULT NULL,
        path_swift VARCHAR(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
        id_user INT DEFAULT NULL,
        PRIMARY KEY (id)
    )");

    echo json_encode(['success' => true, 'message' => 'Database setup completed successfully']);
} catch (PDOException $e) {
    echo json_encode(['success' => false, 'error' => $e->getMessage()]);
}
