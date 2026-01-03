-- ============================================
-- Migration: Add payment_confirmed_by_user_id to sell_bill table
-- ============================================
-- This migration adds:
-- 1. payment_confirmed_by_user_id field to track which user confirmed the payment

SET @db_name = DATABASE();

-- ============================================
-- Add payment_confirmed_by_user_id column to sell_bill table
-- ============================================

-- Add payment_confirmed_by_user_id column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'sell_bill' 
  AND COLUMN_NAME = 'payment_confirmed_by_user_id');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `sell_bill` ADD COLUMN `payment_confirmed_by_user_id` int(11) DEFAULT NULL COMMENT ''ID of user who confirmed the payment'' AFTER `payment_confirmed`',
  'SELECT ''Column payment_confirmed_by_user_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraint to users table
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'sell_bill' 
  AND CONSTRAINT_NAME = 'fk_sell_bill_payment_confirmed_by_user');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `sell_bill` ADD CONSTRAINT `fk_sell_bill_payment_confirmed_by_user` FOREIGN KEY (`payment_confirmed_by_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL',
  'SELECT ''Foreign key fk_sell_bill_payment_confirmed_by_user already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index for better query performance
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'sell_bill' 
  AND INDEX_NAME = 'idx_payment_confirmed_by_user_id');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `sell_bill` ADD INDEX `idx_payment_confirmed_by_user_id` (`payment_confirmed_by_user_id`)',
  'SELECT ''Index idx_payment_confirmed_by_user_id already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

