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

    // Get the purchase bill ID from POST data
    $buy_bill_id = $_POST['buy_bill_id'] ?? null;

    if (!$buy_bill_id) {
        throw new Exception('Purchase bill ID is required');
    }

    // Validate that the purchase bill exists
    $stmt = $pdo->prepare("SELECT id, bill_ref FROM buy_bill WHERE id = ?");
    $stmt->execute([$buy_bill_id]);
    $purchase_bill = $stmt->fetch(PDO::FETCH_ASSOC);

    if (!$purchase_bill) {
        throw new Exception('Purchase bill not found');
    }

    // Start transaction for data integrity
    $pdo->beginTransaction();

    try {
        // Step 1: Get count of cars that will be deleted for reporting
        $stmt = $pdo->prepare("
            SELECT COUNT(*) as car_count 
            FROM cars_stock cs
            INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
            WHERE bd.id_buy_bill = ?
        ");
        $stmt->execute([$buy_bill_id]);
        $car_count = $stmt->fetch(PDO::FETCH_ASSOC)['car_count'];

        // Step 2: Delete cars from cars_stock that are associated with this purchase bill
        $stmt = $pdo->prepare("
            DELETE cs FROM cars_stock cs
            INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
            WHERE bd.id_buy_bill = ?
        ");
        $stmt->execute([$buy_bill_id]);
        $deleted_cars = $stmt->rowCount();

        // Step 3: Get count of buy_details that will be deleted
        $stmt = $pdo->prepare("SELECT COUNT(*) as detail_count FROM buy_details WHERE id_buy_bill = ?");
        $stmt->execute([$buy_bill_id]);
        $detail_count = $stmt->fetch(PDO::FETCH_ASSOC)['detail_count'];

        // Step 4: Delete buy_details records associated with this purchase bill
        $stmt = $pdo->prepare("DELETE FROM buy_details WHERE id_buy_bill = ?");
        $stmt->execute([$buy_bill_id]);
        $deleted_details = $stmt->rowCount();

        // Step 5: Get count of payments that will be deleted
        $stmt = $pdo->prepare("SELECT COUNT(*) as payment_count FROM buy_payments WHERE id_buy_bill = ?");
        $stmt->execute([$buy_bill_id]);
        $payment_count = $stmt->fetch(PDO::FETCH_ASSOC)['payment_count'];

        // Step 6: Delete buy_payments records associated with this purchase bill
        $stmt = $pdo->prepare("DELETE FROM buy_payments WHERE id_buy_bill = ?");
        $stmt->execute([$buy_bill_id]);
        $deleted_payments = $stmt->rowCount();

        // Step 7: Delete the purchase bill itself
        $stmt = $pdo->prepare("DELETE FROM buy_bill WHERE id = ?");
        $stmt->execute([$buy_bill_id]);
        $deleted_bill = $stmt->rowCount();

        // Commit the transaction
        $pdo->commit();

        // Return success response with details
        echo json_encode([
            'success' => true,
            'message' => 'Purchase bill and associated data removed successfully',
            'data' => [
                'purchase_bill_id' => $buy_bill_id,
                'purchase_bill_ref' => $purchase_bill['bill_ref'],
                'deleted_cars' => $deleted_cars,
                'deleted_details' => $deleted_details,
                'deleted_payments' => $deleted_payments,
                'deleted_bill' => $deleted_bill,
                'total_records_deleted' => $deleted_cars + $deleted_details + $deleted_payments + $deleted_bill
            ]
        ]);

    } catch (Exception $e) {
        // Rollback transaction on error
        $pdo->rollBack();
        throw $e;
    }

} catch (PDOException $e) {
    echo json_encode([
        'success' => false, 
        'error' => 'Database error: ' . $e->getMessage()
    ]);
} catch (Exception $e) {
    echo json_encode([
        'success' => false, 
        'error' => $e->getMessage()
    ]);
}
?> 