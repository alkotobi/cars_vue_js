-- ============================================
-- Rollback Script: Car Files Management System
-- Description: Remove the new file management tables
-- WARNING: This will delete all data in the new tables!
-- ============================================

-- Drop views first
DROP VIEW IF EXISTS `v_car_file_transfers`;
DROP VIEW IF EXISTS `v_car_files_with_tracking`;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS `car_file_transfers`;
DROP TABLE IF EXISTS `car_file_physical_tracking`;
DROP TABLE IF EXISTS `car_files`;
DROP TABLE IF EXISTS `car_file_categories`;

-- Note: The old columns in cars_stock table are preserved
-- (path_documents, sell_pi_path, buy_pi_path, path_coo, path_coc)

