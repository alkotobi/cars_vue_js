-- ============================================
-- Migration: Add alerts permissions
-- ============================================
-- This migration adds:
-- 1. can_c_alerts - General permission to view alerts banner
-- 2. can_c_cars_alerts - Permission to view car-related alerts (unloaded, not arrived, no license, no docs sent)
-- 3. can_c_sell_bill_payemts_alert - Permission to view sell bill payment alerts (unconfirmed, not paid, not fully paid)

SET @db_name = DATABASE();

-- ============================================
-- Add can_c_alerts permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_alerts');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_alerts'', ''Can view alerts banner'')',
  'SELECT ''Permission can_c_alerts already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Add can_c_cars_alerts permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_cars_alerts');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_cars_alerts'', ''Can view car-related alerts (unloaded, not arrived, no license, no docs sent)'')',
  'SELECT ''Permission can_c_cars_alerts already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Add can_c_sell_bill_payemts_alert permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_c_sell_bill_payemts_alert');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_c_sell_bill_payemts_alert'', ''Can view sell bill payment alerts (unconfirmed, not paid, not fully paid)'')',
  'SELECT ''Permission can_c_sell_bill_payemts_alert already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

