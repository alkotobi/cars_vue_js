-- ============================================
-- Migration: Add pay purchases permission
-- ============================================
-- This migration adds:
-- 1. can_pay_purchases - Permission to record payments to suppliers (buy bills)
--
-- The app's BuyBillPaymentsView gates on admin only; this dedicated permission
-- lets a non-admin role (and the AI user) be granted just the supplier-payment
-- area instead of full admin.

SET @db_name = DATABASE();

-- ============================================
-- Add can_pay_purchases permission
-- ============================================
SET @perm_exists = (SELECT COUNT(*) FROM `permissions` WHERE `permission_name` = 'can_pay_purchases');
SET @sql = IF(@perm_exists = 0, 
  'INSERT INTO `permissions` (`permission_name`, `description`) VALUES (''can_pay_purchases'', ''Can record payments to suppliers for buy bills'')',
  'SELECT ''Permission can_pay_purchases already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
