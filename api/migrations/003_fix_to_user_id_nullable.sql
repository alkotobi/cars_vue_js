-- ============================================
-- Migration: Fix to_user_id to allow NULL
-- ============================================
-- This migration makes to_user_id nullable in car_file_transfers
-- to support transfers to agents and clients

DELIMITER $$

DROP PROCEDURE IF EXISTS fix_to_user_id_nullable$$

CREATE PROCEDURE fix_to_user_id_nullable()
BEGIN
  DECLARE col_nullable VARCHAR(3);
  DECLARE fk_exists INT DEFAULT 0;
  
  SET @db_name = DATABASE();
  SET @table_name = 'car_file_transfers';
  
  -- Check if to_user_id is currently NOT NULL
  SELECT IS_NULLABLE INTO col_nullable
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
    AND TABLE_NAME = @table_name 
    AND COLUMN_NAME = 'to_user_id';
  
  -- Only alter if it's currently NOT NULL
  IF col_nullable = 'NO' THEN
    SET @sql = 'ALTER TABLE `car_file_transfers` MODIFY COLUMN `to_user_id` int(11) DEFAULT NULL COMMENT ''FK to users - who received (NULL if transferred to agent or client)''';
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
  
  -- Update foreign key constraint to allow NULL (drop and recreate)
  SELECT COUNT(*) INTO fk_exists
  FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
    AND TABLE_NAME = @table_name 
    AND CONSTRAINT_NAME = 'fk_transfers_to_user';
  
  IF fk_exists > 0 THEN
    SET @sql = 'ALTER TABLE `car_file_transfers` DROP FOREIGN KEY `fk_transfers_to_user`';
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
    
    SET @sql = 'ALTER TABLE `car_file_transfers` ADD CONSTRAINT `fk_transfers_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL';
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$

DELIMITER ;

-- Execute the procedure
CALL fix_to_user_id_nullable();

-- Drop the procedure
DROP PROCEDURE IF EXISTS fix_to_user_id_nullable;

