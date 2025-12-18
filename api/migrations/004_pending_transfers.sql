-- ============================================
-- Migration: Pending Transfer Approval System
-- ============================================
-- This migration adds transfer_status to car_file_transfers
-- to support pending/approved/rejected transfer workflow

DELIMITER $$

DROP PROCEDURE IF EXISTS add_transfer_status_column$$

CREATE PROCEDURE add_transfer_status_column()
BEGIN
  DECLARE col_exists INT DEFAULT 0;
  
  SET @db_name = DATABASE();
  SET @table_name = 'car_file_transfers';
  
  -- Check if transfer_status column exists
  SELECT COUNT(*) INTO col_exists
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
    AND TABLE_NAME = @table_name 
    AND COLUMN_NAME = 'transfer_status';
  
  -- Add column if it doesn't exist
  IF col_exists = 0 THEN
    SET @sql = 'ALTER TABLE `car_file_transfers` ADD COLUMN `transfer_status` enum(''pending'', ''approved'', ''rejected'') NOT NULL DEFAULT ''approved'' COMMENT ''Transfer status: pending (waiting approval), approved (completed), rejected (cancelled)'' AFTER `transfer_type`';
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
    
    -- Update existing transfers to approved status
    SET @sql = 'UPDATE `car_file_transfers` SET `transfer_status` = ''approved'' WHERE `transfer_status` IS NULL';
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
    
    -- Add index for faster queries
    SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
      WHERE TABLE_SCHEMA = @db_name 
      AND TABLE_NAME = @table_name 
      AND INDEX_NAME = 'idx_transfer_status');
    
    IF @idx_exists = 0 THEN
      SET @sql = 'ALTER TABLE `car_file_transfers` ADD INDEX `idx_transfer_status` (`transfer_status`)';
      PREPARE stmt FROM @sql;
      EXECUTE stmt;
      DEALLOCATE PREPARE stmt;
    END IF;
  END IF;
END$$

DELIMITER ;

-- Execute the procedure
CALL add_transfer_status_column();

-- Drop the procedure
DROP PROCEDURE IF EXISTS add_transfer_status_column;

