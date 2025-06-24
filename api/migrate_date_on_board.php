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

    // Check if date_on_board column exists
    $stmt = $pdo->prepare("SHOW COLUMNS FROM loaded_containers LIKE 'date_on_board'");
    $stmt->execute();
    $columnExists = $stmt->rowCount() > 0;

    if (!$columnExists) {
        // Add date_on_board column
        $pdo->exec("ALTER TABLE loaded_containers ADD COLUMN date_on_board date DEFAULT NULL");
        echo json_encode([
            'success' => true,
            'message' => 'date_on_board column added successfully to loaded_containers table'
        ]);
    } else {
        echo json_encode([
            'success' => true,
            'message' => 'date_on_board column already exists in loaded_containers table'
        ]);
    }

} catch (Exception $e) {
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage()
    ]);
}
?> 