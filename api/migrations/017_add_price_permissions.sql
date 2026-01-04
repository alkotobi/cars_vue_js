-- ============================================
-- Migration: Add price-related permissions
-- ============================================
-- This migration adds:
-- 1. can_c_freight - Permission to view freight information
-- 2. can_c_fob_price - Permission to view FOB price information
-- 3. can_c_cfr_prices - Permission to view CFR prices information

SET @db_name = DATABASE();

-- ============================================
-- Add can_c_freight permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_freight');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_freight'', ''Can view freight information'')',
  'SELECT ''Permission can_c_freight already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Add can_c_fob_price permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_fob_price');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_fob_price'', ''Can view FOB price information'')',
  'SELECT ''Permission can_c_fob_price already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Add can_c_cfr_prices permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_cfr_prices');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_cfr_prices'', ''Can view CFR prices information'')',
  'SELECT ''Permission can_c_cfr_prices already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

