<?php
require_once 'config.php';

try {
    $pdo = new PDO("mysql:host=$db_host;dbname=$db_name;charset=utf8", $db_user, $db_pass);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    echo "Starting migration: Add unique constraint to chat_last_read_message table\n";

    // First, let's check if there are any duplicate records and clean them up
    echo "Checking for duplicate records...\n";
    
    $stmt = $pdo->query("
        SELECT id_group, id_user, COUNT(*) as count, MAX(id) as max_id
        FROM chat_last_read_message 
        GROUP BY id_group, id_user 
        HAVING COUNT(*) > 1
    ");
    
    $duplicates = $stmt->fetchAll(PDO::FETCH_ASSOC);
    
    if (!empty($duplicates)) {
        echo "Found " . count($duplicates) . " duplicate group/user combinations. Cleaning up...\n";
        
        foreach ($duplicates as $duplicate) {
            $group_id = $duplicate['id_group'];
            $user_id = $duplicate['id_user'];
            $max_id = $duplicate['max_id'];
            
            // Keep only the record with the highest ID (most recent) and delete the rest
            $deleteStmt = $pdo->prepare("
                DELETE FROM chat_last_read_message 
                WHERE id_group = ? AND id_user = ? AND id != ?
            ");
            $deleteStmt->execute([$group_id, $user_id, $max_id]);
            
            $deletedCount = $deleteStmt->rowCount();
            echo "Deleted $deletedCount duplicate records for group $group_id, user $user_id\n";
        }
    } else {
        echo "No duplicate records found.\n";
    }

    // Now add the unique constraint
    echo "Adding unique constraint...\n";
    
    // Check if the unique constraint already exists
    $stmt = $pdo->query("
        SELECT CONSTRAINT_NAME 
        FROM information_schema.TABLE_CONSTRAINTS 
        WHERE TABLE_SCHEMA = '$db_name' 
        AND TABLE_NAME = 'chat_last_read_message' 
        AND CONSTRAINT_TYPE = 'UNIQUE'
        AND CONSTRAINT_NAME LIKE '%id_group%'
    ");
    
    $existingConstraint = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if ($existingConstraint) {
        echo "Unique constraint already exists: " . $existingConstraint['CONSTRAINT_NAME'] . "\n";
    } else {
        // Add unique constraint
        $pdo->exec("
            ALTER TABLE chat_last_read_message 
            ADD UNIQUE KEY unique_group_user (id_group, id_user)
        ");
        echo "Unique constraint 'unique_group_user' added successfully.\n";
    }

    // Verify the constraint was added
    $stmt = $pdo->query("
        SELECT CONSTRAINT_NAME 
        FROM information_schema.TABLE_CONSTRAINTS 
        WHERE TABLE_SCHEMA = '$db_name' 
        AND TABLE_NAME = 'chat_last_read_message' 
        AND CONSTRAINT_TYPE = 'UNIQUE'
        AND CONSTRAINT_NAME LIKE '%id_group%'
    ");
    
    $constraint = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if ($constraint) {
        echo "✓ Unique constraint verified: " . $constraint['CONSTRAINT_NAME'] . "\n";
    } else {
        echo "✗ Error: Unique constraint not found after creation.\n";
    }

    // Show final table structure
    echo "\nFinal table structure:\n";
    $stmt = $pdo->query("DESCRIBE chat_last_read_message");
    $columns = $stmt->fetchAll(PDO::FETCH_ASSOC);
    
    foreach ($columns as $column) {
        echo "- {$column['Field']}: {$column['Type']} {$column['Key']} {$column['Null']} {$column['Default']}\n";
    }

    // Show indexes
    echo "\nTable indexes:\n";
    $stmt = $pdo->query("SHOW INDEX FROM chat_last_read_message");
    $indexes = $stmt->fetchAll(PDO::FETCH_ASSOC);
    
    foreach ($indexes as $index) {
        echo "- {$index['Key_name']}: {$index['Column_name']}\n";
    }

    echo "\n✓ Migration completed successfully!\n";
    echo "Now the table will only allow one record per group/user combination.\n";
    echo "The INSERT ... ON DUPLICATE KEY UPDATE statement in the code will work correctly.\n";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage() . "\n";
    exit(1);
}
?> 