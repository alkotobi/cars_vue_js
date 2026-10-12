-- ============================================
-- Migration: Add buy price permission
-- ============================================
-- This migration adds:
-- 1. can_c_buy_price - Permission to view buy price information

SET @db_name = DATABASE();

-- ============================================
-- Add can_c_buy_price permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_buy_price');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_buy_price'', ''Can view buy price information'')',
  'SELECT ''Permission can_c_buy_price already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;