<?php
header('Content-Type: application/json');
require_once 'config.php';

try {
    $pdo = new PDO(
        "mysql:host=$db_host;dbname=$db_name;charset=utf8",
        $db_user,
        $db_pass
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    // Create priorities table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `priorities` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `priority` varchar(255) DEFAULT NULL,
        `id_color` int(11) DEFAULT NULL,
        `power` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci");

    // Create tasks table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `tasks` (
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
        `is_task_for_transfer` int(11) DEFAULT 0,
        `is_task_for_supplier` tinyint(1) DEFAULT 0,
        `is_task_for_client` tinyint(1) DEFAULT 0,
        `is_task_for_user` tinyint(1) DEFAULT 0,
        `id_chat_grroup` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci");

    // Create chat_groups table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `chat_groups` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `name` varchar(255) DEFAULT NULL,
        `description` text DEFAULT NULL,
        `id_user_owner` int(11) DEFAULT NULL,
        `is_active` tinyint(1) DEFAULT 1,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci");

    // Create chat_messages table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `chat_messages` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `id_chat_group` int(11) DEFAULT NULL,
        `message_from_user_id` int(11) DEFAULT NULL,
        `chat_replay_to_message_id` int(11) DEFAULT NULL,
        `message` text DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci");

    // Create chat_read_by table
    $pdo->exec("CREATE TABLE IF NOT EXISTS `chat_read_by` (
        `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
        `id_chat_message` int(11) DEFAULT NULL,
        `id_user` int(11) DEFAULT NULL,
        PRIMARY KEY (`id`)
    ) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci");

    echo json_encode(['success' => true, 'message' => 'New tables created successfully']);
} catch (PDOException $e) {
    echo json_encode(['success' => false, 'error' => $e->getMessage()]);
}
?> 