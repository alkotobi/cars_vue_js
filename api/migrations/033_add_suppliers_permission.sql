-- ============================================
-- Migration: Add suppliers permission
-- ============================================
-- This migration adds:
-- 1. can_c_suppliers - Permission to view suppliers information

SET @db_name = DATABASE();

-- ============================================
-- Add can_c_suppliers permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_suppliers');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_suppliers'', ''Can view suppliers information'')',
  'SELECT ''Permission can_c_suppliers already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;