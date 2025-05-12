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
    // $pdo->exec("CREATE DATABASE IF NOT EXISTS $db_name");
    // $pdo->exec("USE $db_name");

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
        ('can_manage_cars', 'Can manage cars inventory')
    ");
    $stmt->execute();

    // Create default admin user if not exists
    $defaultPassword = password_hash('admin123', PASSWORD_DEFAULT);
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
        id INT PRIMARY KEY AUTO_INCREMENT,
        car_name VARCHAR(255) DEFAULT NULL,
        notes TEXT DEFAULT NULL,
        UNIQUE KEY car_name (car_name)
    )");

    // Create cars_stock table
    $pdo->exec("CREATE TABLE IF NOT EXISTS cars_stock (
        id INT PRIMARY KEY AUTO_INCREMENT,
        id_car_name INT NOT NULL,
        notes TEXT DEFAULT NULL,
        id_brand INT DEFAULT NULL,
        id_color INT NOT NULL,
        id_supplier INT NOT NULL,
        date_buy DATETIME NOT NULL,
        price_buy DECIMAL(10,2) NOT NULL,
        date_sell DATETIME DEFAULT NULL,
        id_client INT DEFAULT NULL,
        price_cell DECIMAL(10,2) DEFAULT NULL,
        id_port_loading INT DEFAULT NULL,
        id_port_discharge INT DEFAULT NULL,
        vin VARCHAR(255) DEFAULT NULL,
        path_documents VARCHAR(255) DEFAULT NULL,
        date_loding DATETIME DEFAULT NULL,
        date_send_documents DATETIME DEFAULT NULL,
        deposit DECIMAL(10,2) DEFAULT NULL,
        balance DECIMAL(10,2) DEFAULT NULL,
        date_balance DATETIME DEFAULT NULL,
        hidden TINYINT(1) DEFAULT 0,
        id_buy_pi INT DEFAULT NULL,
        id_sell_pi VARCHAR(255) DEFAULT NULL,
        sell_pi_path VARCHAR(255) DEFAULT NULL,
        buy_pi_path VARCHAR(255) DEFAULT NULL,
        FOREIGN KEY (id_car_name) REFERENCES cars_names(id),
        FOREIGN KEY (id_brand) REFERENCES brands(id),
        FOREIGN KEY (id_color) REFERENCES colors(id),
        FOREIGN KEY (id_supplier) REFERENCES suppliers(id),
        FOREIGN KEY (id_client) REFERENCES clients(id),
        FOREIGN KEY (id_port_loading) REFERENCES loading_ports(id),
        FOREIGN KEY (id_port_discharge) REFERENCES discharge_ports(id)
    )");

    echo json_encode([
        'success' => true,
        'message' => 'Database initialized successfully',
        'default_credentials' => [
            'username' => 'admin',
            'password' => 'admin123'
        ]
    ]);

} catch (PDOException $e) {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage()
    ]);
}