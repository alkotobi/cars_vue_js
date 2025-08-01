<?php
require_once 'config.php';

try {
    $pdo = new PDO("mysql:host={$db_config['host']};dbname={$db_config['database']}", $db_config['username'], $db_config['password']);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    echo "Starting CFR DA migration...\n";
    
    // Get all records that have price_cell, rate, and freight but no cfr_da
    $stmt = $pdo->prepare("
        SELECT id, price_cell, rate, freight 
        FROM cars_stock 
        WHERE price_cell IS NOT NULL 
        AND rate IS NOT NULL 
        AND cfr_da IS NULL
    ");
    $stmt->execute();
    $records = $stmt->fetchAll(PDO::FETCH_ASSOC);
    
    echo "Found " . count($records) . " records to migrate\n";
    
    $updated = 0;
    foreach ($records as $record) {
        $price_cell = floatval($record['price_cell']);
        $rate = floatval($record['rate']);
        $freight = floatval($record['freight'] ?? 0);
        
        // Calculate CFR DA: (price_cell + freight) * rate
        $cfr_da = ($price_cell + $freight) * $rate;
        
        // Update the record
        $updateStmt = $pdo->prepare("
            UPDATE cars_stock 
            SET cfr_da = ? 
            WHERE id = ?
        ");
        $updateStmt->execute([$cfr_da, $record['id']]);
        
        $updated++;
        echo "Updated record ID {$record['id']}: price_cell={$price_cell}, rate={$rate}, freight={$freight} -> cfr_da={$cfr_da}\n";
    }
    
    echo "Migration completed! Updated {$updated} records.\n";
    
} catch (PDOException $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
?> 