<?php
require_once 'config.php';

try {
    $pdo = new PDO("mysql:host=$host;dbname=$dbname;charset=utf8", $username, $password);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    echo "Starting batch cars update...\n";
    
    // First, let's see how many cars will be affected
    $countQuery = "
        SELECT COUNT(*) as cars_to_update
        FROM cars_stock cs
        INNER JOIN sell_bill sb ON cs.id_sell = sb.id
        WHERE sb.is_batch_sell = 1
        AND cs.hidden = 0
    ";
    
    $countStmt = $pdo->prepare($countQuery);
    $countStmt->execute();
    $countResult = $countStmt->fetch(PDO::FETCH_ASSOC);
    
    echo "Cars that will be updated: " . $countResult['cars_to_update'] . "\n";
    
    if ($countResult['cars_to_update'] > 0) {
        // Show details of cars that will be updated
        $detailsQuery = "
            SELECT 
                cs.id as car_id,
                cs.vin,
                sb.id as sell_bill_id,
                sb.bill_ref,
                sb.is_batch_sell,
                cs.is_batch as current_is_batch
            FROM cars_stock cs
            INNER JOIN sell_bill sb ON cs.id_sell = sb.id
            WHERE sb.is_batch_sell = 1
            AND cs.hidden = 0
            LIMIT 10
        ";
        
        $detailsStmt = $pdo->prepare($detailsQuery);
        $detailsStmt->execute();
        $details = $detailsStmt->fetchAll(PDO::FETCH_ASSOC);
        
        echo "\nSample cars that will be updated:\n";
        foreach ($details as $car) {
            echo "Car ID: {$car['car_id']}, VIN: {$car['vin']}, Sell Bill: {$car['bill_ref']}, Current is_batch: {$car['current_is_batch']}\n";
        }
        
        if ($countResult['cars_to_update'] > 10) {
            echo "... and " . ($countResult['cars_to_update'] - 10) . " more cars\n";
        }
        
        echo "\nProceeding with update...\n";
        
        // Perform the update
        $updateQuery = "
            UPDATE cars_stock cs
            INNER JOIN sell_bill sb ON cs.id_sell = sb.id
            SET cs.is_batch = 1
            WHERE sb.is_batch_sell = 1
            AND cs.hidden = 0
        ";
        
        $updateStmt = $pdo->prepare($updateQuery);
        $updateStmt->execute();
        
        $affectedRows = $updateStmt->rowCount();
        echo "Update completed successfully!\n";
        echo "Rows affected: $affectedRows\n";
        
    } else {
        echo "No cars found that need to be updated.\n";
    }
    
} catch (PDOException $e) {
    echo "Database error: " . $e->getMessage() . "\n";
} catch (Exception $e) {
    echo "Error: " . $e->getMessage() . "\n";
}
?> 