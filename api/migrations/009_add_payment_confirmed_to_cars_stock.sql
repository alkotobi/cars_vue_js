-- ============================================
-- Migration: Add payment_confirmed field to cars_stock table
-- ============================================
-- This migration adds:
-- 1. payment_confirmed field (tinyint(1), default 0/false)
--    Only admins or users with permission 'can_confirm_payment' can set this to true

SET @db_name = DATABASE();

-- ============================================
-- Add payment_confirmed field to cars_stock
-- ============================================
ALTER TABLE `cars_stock`
ADD COLUMN `payment_confirmed` tinyint(1) DEFAULT 0 COMMENT 'Payment confirmed status - only admins or users with can_confirm_payment permission can set to true' AFTER `id_color`,
ADD KEY `idx_payment_confirmed` (`payment_confirmed`);

-- Set all existing cars to payment_confirmed = 0 (false) by default
UPDATE `cars_stock` 
SET `payment_confirmed` = 0 
WHERE `payment_confirmed` IS NULL;

