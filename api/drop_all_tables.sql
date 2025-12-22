-- ============================================
-- Drop All Tables Script
-- ============================================
-- This script safely drops all tables in the database
-- It handles foreign key constraints by temporarily disabling them
-- 
-- WARNING: This will delete ALL data in ALL tables!
-- Use with extreme caution!
-- ============================================

-- Step 1: Disable foreign key checks
SET FOREIGN_KEY_CHECKS = 0;

-- Step 2: Drop all tables
-- Get all table names and drop them dynamically
SET @tables = NULL;
SELECT GROUP_CONCAT('`', table_schema, '`.`', table_name, '`') 
INTO @tables
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_type = 'BASE TABLE';

SET @tables = CONCAT('DROP TABLE IF EXISTS ', @tables);

-- Execute the drop statement
PREPARE stmt FROM @tables;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Step 3: Re-enable foreign key checks
SET FOREIGN_KEY_CHECKS = 1;

-- ============================================
-- Alternative: Manual drop (if dynamic doesn't work)
-- ============================================
-- Uncomment the section below if you prefer to manually specify tables
-- Make sure to drop child tables before parent tables

/*
SET FOREIGN_KEY_CHECKS = 0;

-- Drop tables with foreign keys first (child tables)
DROP TABLE IF EXISTS `car_file_transfers`;
DROP TABLE IF EXISTS `car_file_physical_tracking`;
DROP TABLE IF EXISTS `car_files`;
DROP TABLE IF EXISTS `selection_comments`;
DROP TABLE IF EXISTS `selection_ownership_history`;
DROP TABLE IF EXISTS `car_selections`;
DROP TABLE IF EXISTS `team_members`;
DROP TABLE IF EXISTS `jobs`;
DROP TABLE IF EXISTS `transfers_inter`;
DROP TABLE IF EXISTS `transfers`;
DROP TABLE IF EXISTS `transfer_details`;
DROP TABLE IF EXISTS `tasks`;
DROP TABLE IF EXISTS `tracking`;
DROP TABLE IF EXISTS `chat_read_by`;
DROP TABLE IF EXISTS `chat_messages`;
DROP TABLE IF EXISTS `chat_last_read_message`;
DROP TABLE IF EXISTS `chat_users`;
DROP TABLE IF EXISTS `chat_groups`;
DROP TABLE IF EXISTS `sell_payments`;
DROP TABLE IF EXISTS `buy_payments`;
DROP TABLE IF EXISTS `buy_details`;
DROP TABLE IF EXISTS `cars_stock`;
DROP TABLE IF EXISTS `role_permissions`;
DROP TABLE IF EXISTS `rates`;
DROP TABLE IF EXISTS `loaded_containers`;
DROP TABLE IF EXISTS `loading`;

-- Drop parent/reference tables
DROP TABLE IF EXISTS `sell_bill`;
DROP TABLE IF EXISTS `buy_bill`;
DROP TABLE IF EXISTS `clients`;
DROP TABLE IF EXISTS `suppliers`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `roles`;
DROP TABLE IF EXISTS `permissions`;
DROP TABLE IF EXISTS `cars_names`;
DROP TABLE IF EXISTS `colors`;
DROP TABLE IF EXISTS `loading_ports`;
DROP TABLE IF EXISTS `discharge_ports`;
DROP TABLE IF EXISTS `warehouses`;
DROP TABLE IF EXISTS `containers`;
DROP TABLE IF EXISTS `shipping_lines`;
DROP TABLE IF EXISTS `car_file_categories`;
DROP TABLE IF EXISTS `custom_clearance_agents`;
DROP TABLE IF EXISTS `teams`;
DROP TABLE IF EXISTS `priorities`;
DROP TABLE IF EXISTS `defaults`;
DROP TABLE IF EXISTS `banks`;
DROP TABLE IF EXISTS `brands`;
DROP TABLE IF EXISTS `versions`;
DROP TABLE IF EXISTS `db_updates`;
DROP TABLE IF EXISTS `adv_sql`;
DROP TABLE IF EXISTS `dbs`;
DROP TABLE IF EXISTS `login`;

SET FOREIGN_KEY_CHECKS = 1;
*/

-- ============================================
-- Verification: Check remaining tables
-- ============================================
-- Run this query after execution to verify all tables are dropped:
-- SELECT COUNT(*) as remaining_tables 
-- FROM information_schema.tables 
-- WHERE table_schema = DATABASE() 
--   AND table_type = 'BASE TABLE';

