-- ============================================
-- Migration: Custom Clearance Agents
-- ============================================
-- This migration adds support for custom clearance agents
-- and updates the physical tracking system to always have an owner

-- ============================================
-- 1. Create custom_clearance_agents table
-- ============================================
CREATE TABLE IF NOT EXISTS `custom_clearance_agents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Agent name or company name',
  `contact_person` varchar(255) DEFAULT NULL COMMENT 'Contact person name',
  `phone` varchar(50) DEFAULT NULL COMMENT 'Phone number',
  `email` varchar(255) DEFAULT NULL COMMENT 'Email address',
  `address` text DEFAULT NULL COMMENT 'Physical address',
  `license_number` varchar(100) DEFAULT NULL COMMENT 'License or registration number',
  `notes` text DEFAULT NULL COMMENT 'Additional notes',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- 2. Update car_file_physical_tracking table
-- ============================================
-- Add fields to support custom clearance agents and ensure always has owner
-- Check if columns exist before adding them

SET @db_name = DATABASE();
SET @table_name = 'car_file_physical_tracking';

-- Add custom_clearance_agent_id column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'custom_clearance_agent_id');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD COLUMN `custom_clearance_agent_id` int(11) DEFAULT NULL COMMENT ''FK to custom_clearance_agents - if checked out to agent'' AFTER `current_holder_id`',
  'SELECT ''Column custom_clearance_agent_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add checkout_type column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'checkout_type');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD COLUMN `checkout_type` enum(''user'', ''client'', ''custom_clearance_agent'') NOT NULL DEFAULT ''user'' COMMENT ''Type of checkout: user, client, or custom clearance agent'' AFTER `custom_clearance_agent_id`',
  'SELECT ''Column checkout_type already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add client_id column if it doesn't exist
-- Use a stored procedure to handle the conditional logic
DELIMITER $$

DROP PROCEDURE IF EXISTS add_client_id_column$$

CREATE PROCEDURE add_client_id_column()
BEGIN
  DECLARE col_exists INT DEFAULT 0;
  DECLARE checkout_type_exists INT DEFAULT 0;
  DECLARE agent_id_exists INT DEFAULT 0;
  DECLARE sql_stmt TEXT;
  
  SET @db_name = DATABASE();
  SET @table_name = 'car_file_physical_tracking';
  
  -- Check if client_id exists
  SELECT COUNT(*) INTO col_exists
  FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
    AND TABLE_NAME = @table_name 
    AND COLUMN_NAME = 'client_id';
  
  -- Only proceed if column doesn't exist
  IF col_exists = 0 THEN
    -- Check if checkout_type exists
    SELECT COUNT(*) INTO checkout_type_exists
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = @db_name 
      AND TABLE_NAME = @table_name 
      AND COLUMN_NAME = 'checkout_type';
    
    -- Check if custom_clearance_agent_id exists
    SELECT COUNT(*) INTO agent_id_exists
    FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_SCHEMA = @db_name 
      AND TABLE_NAME = @table_name 
      AND COLUMN_NAME = 'custom_clearance_agent_id';
    
    -- Build SQL based on what exists
    IF checkout_type_exists > 0 THEN
      SET sql_stmt = 'ALTER TABLE `car_file_physical_tracking` ADD COLUMN `client_id` int(11) DEFAULT NULL COMMENT ''FK to clients - if checked out to client'' AFTER `checkout_type`';
    ELSEIF agent_id_exists > 0 THEN
      SET sql_stmt = 'ALTER TABLE `car_file_physical_tracking` ADD COLUMN `client_id` int(11) DEFAULT NULL COMMENT ''FK to clients - if checked out to client'' AFTER `custom_clearance_agent_id`';
    ELSE
      SET sql_stmt = 'ALTER TABLE `car_file_physical_tracking` ADD COLUMN `client_id` int(11) DEFAULT NULL COMMENT ''FK to clients - if checked out to client'' AFTER `current_holder_id`';
    END IF;
    
    SET @sql = sql_stmt;
    PREPARE stmt FROM @sql;
    EXECUTE stmt;
    DEALLOCATE PREPARE stmt;
  END IF;
END$$

DELIMITER ;

-- Execute the procedure
CALL add_client_id_column();

-- Drop the procedure
DROP PROCEDURE IF EXISTS add_client_id_column;

-- Add indexes if they don't exist (checking by trying to add and ignoring errors is not ideal, but MySQL doesn't support IF NOT EXISTS for indexes)
-- We'll add them conditionally by checking if they exist first
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_agent_id');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD INDEX `idx_agent_id` (`custom_clearance_agent_id`)',
  'SELECT ''Index idx_agent_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_client_id');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD INDEX `idx_client_id` (`client_id`)',
  'SELECT ''Index idx_client_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_checkout_type');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD INDEX `idx_checkout_type` (`checkout_type`)',
  'SELECT ''Index idx_checkout_type already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraints if they don't exist
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND CONSTRAINT_NAME = 'fk_tracking_agent');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD CONSTRAINT `fk_tracking_agent` FOREIGN KEY (`custom_clearance_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL',
  'SELECT ''Constraint fk_tracking_agent already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND CONSTRAINT_NAME = 'fk_tracking_client');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `car_file_physical_tracking` ADD CONSTRAINT `fk_tracking_client` FOREIGN KEY (`client_id`) REFERENCES `clients` (`id`) ON DELETE SET NULL',
  'SELECT ''Constraint fk_tracking_client already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- 3. Update car_file_transfers table
-- ============================================
-- Add fields to track transfers to/from agents and clients
SET @table_name = 'car_file_transfers';

-- Add from_agent_id column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'from_agent_id');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD COLUMN `from_agent_id` int(11) DEFAULT NULL COMMENT ''FK to custom_clearance_agents - if transferred from agent'' AFTER `from_user_id`',
  'SELECT ''Column from_agent_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add to_agent_id column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'to_agent_id');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD COLUMN `to_agent_id` int(11) DEFAULT NULL COMMENT ''FK to custom_clearance_agents - if transferred to agent'' AFTER `to_user_id`',
  'SELECT ''Column to_agent_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add from_client_name column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'from_client_name');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD COLUMN `from_client_name` varchar(255) DEFAULT NULL COMMENT ''Client name if transferred from client'' AFTER `to_agent_id`',
  'SELECT ''Column from_client_name already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add to_client_name column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'to_client_name');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD COLUMN `to_client_name` varchar(255) DEFAULT NULL COMMENT ''Client name if transferred to client'' AFTER `from_client_name`',
  'SELECT ''Column to_client_name already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add transfer_type column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND COLUMN_NAME = 'transfer_type');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD COLUMN `transfer_type` enum(''user_to_user'', ''user_to_agent'', ''agent_to_user'', ''user_to_client'', ''client_to_user'') NOT NULL DEFAULT ''user_to_user'' AFTER `to_client_name`',
  'SELECT ''Column transfer_type already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add indexes if they don't exist
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_from_agent');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD INDEX `idx_from_agent` (`from_agent_id`)',
  'SELECT ''Index idx_from_agent already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_to_agent');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD INDEX `idx_to_agent` (`to_agent_id`)',
  'SELECT ''Index idx_to_agent already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND INDEX_NAME = 'idx_transfer_type');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD INDEX `idx_transfer_type` (`transfer_type`)',
  'SELECT ''Index idx_transfer_type already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraints if they don't exist
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND CONSTRAINT_NAME = 'fk_transfers_from_agent');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD CONSTRAINT `fk_transfers_from_agent` FOREIGN KEY (`from_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL',
  'SELECT ''Constraint fk_transfers_from_agent already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = @table_name 
  AND CONSTRAINT_NAME = 'fk_transfers_to_agent');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `car_file_transfers` ADD CONSTRAINT `fk_transfers_to_agent` FOREIGN KEY (`to_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL',
  'SELECT ''Constraint fk_transfers_to_agent already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- 4. Set initial owners for existing files
-- ============================================
-- Files that don't have a tracking record should have the uploader as owner
UPDATE `car_files` cf
LEFT JOIN `car_file_physical_tracking` cpt ON cf.id = cpt.car_file_id AND cpt.status IN ('available', 'checked_out')
SET cpt.current_holder_id = cf.uploaded_by,
    cpt.status = 'checked_out',
    cpt.checked_out_at = cf.uploaded_at,
    cpt.checkout_type = 'user'
WHERE cpt.id IS NULL AND cf.is_active = 1;

-- Insert tracking records for files that don't have any
INSERT INTO `car_file_physical_tracking` 
  (car_file_id, current_holder_id, checked_out_at, status, checkout_type)
SELECT 
  cf.id,
  cf.uploaded_by,
  cf.uploaded_at,
  'checked_out',
  'user'
FROM `car_files` cf
WHERE cf.is_active = 1
  AND NOT EXISTS (
    SELECT 1 FROM `car_file_physical_tracking` cpt 
    WHERE cpt.car_file_id = cf.id
  );

