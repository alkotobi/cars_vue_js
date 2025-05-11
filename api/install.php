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
        ('is_exchange_receiver', 'Can receive exchange requests')
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