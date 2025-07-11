-- SQL Script to Remove Purchase Bill and Associated Cars
-- This script will:
-- 1. Delete cars from cars_stock that are associated with the purchase bill
-- 2. Delete buy_details records associated with the purchase bill
-- 3. Delete buy_payments records associated with the purchase bill
-- 4. Delete the purchase bill itself

-- Replace @BUY_BILL_ID with the actual purchase bill ID you want to remove
-- Example: SET @BUY_BILL_ID = 123;

SET @BUY_BILL_ID = ?; -- Replace with actual purchase bill ID

-- Start transaction for data integrity
START TRANSACTION;

-- Step 1: Delete cars from cars_stock that are associated with this purchase bill
-- This deletes cars that have buy_details linked to this purchase bill
DELETE cs FROM cars_stock cs
INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
WHERE bd.id_buy_bill = @BUY_BILL_ID;

-- Step 2: Delete buy_details records associated with this purchase bill
DELETE FROM buy_details 
WHERE id_buy_bill = @BUY_BILL_ID;

-- Step 3: Delete buy_payments records associated with this purchase bill
DELETE FROM buy_payments 
WHERE id_buy_bill = @BUY_BILL_ID;

-- Step 4: Delete the purchase bill itself
DELETE FROM buy_bill 
WHERE id = @BUY_BILL_ID;

-- Commit the transaction
COMMIT;

-- Verification queries (optional - run these to verify the deletion)
-- Check if any cars remain associated with this purchase bill
SELECT COUNT(*) as remaining_cars 
FROM cars_stock cs
INNER JOIN buy_details bd ON cs.id_buy_details = bd.id
WHERE bd.id_buy_bill = @BUY_BILL_ID;

-- Check if any buy_details remain for this purchase bill
SELECT COUNT(*) as remaining_buy_details 
FROM buy_details 
WHERE id_buy_bill = @BUY_BILL_ID;

-- Check if any payments remain for this purchase bill
SELECT COUNT(*) as remaining_payments 
FROM buy_payments 
WHERE id_buy_bill = @BUY_BILL_ID;

-- Check if the purchase bill still exists
SELECT COUNT(*) as purchase_bill_exists 
FROM buy_bill 
WHERE id = @BUY_BILL_ID; 