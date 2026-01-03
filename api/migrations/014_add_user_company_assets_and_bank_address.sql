-- ============================================
-- Migration: Add company assets and bank account to users table, and company_address to banks table
-- ============================================
-- This migration adds:
-- 1. Multi-tenancy support fields to users table:
--    - is_diffrent_company (boolean flag)
--    - path_logo (path to company logo)
--    - path_letter_head (path to letterhead)
--    - path_stamp (path to stamp/gml2)
--    - path_contract_terms (path to contract terms JSON)
--    - id_bank_account (foreign key to banks table)
-- 2. company_address field to banks table

SET @db_name = DATABASE();

-- ============================================
-- 1. Add columns to users table
-- ============================================

-- Add is_diffrent_company column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'is_diffrent_company');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `is_diffrent_company` tinyint(1) DEFAULT 0 COMMENT ''Flag to indicate if user has different company assets'' AFTER `max_unpayed_created_bills`',
  'SELECT ''Column is_diffrent_company already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add path_logo column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'path_logo');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `path_logo` varchar(500) DEFAULT NULL COMMENT ''Path to company logo file'' AFTER `is_diffrent_company`',
  'SELECT ''Column path_logo already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add path_letter_head column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'path_letter_head');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `path_letter_head` varchar(500) DEFAULT NULL COMMENT ''Path to letterhead file'' AFTER `path_logo`',
  'SELECT ''Column path_letter_head already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add path_stamp column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'path_stamp');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `path_stamp` varchar(500) DEFAULT NULL COMMENT ''Path to stamp/gml2 file'' AFTER `path_letter_head`',
  'SELECT ''Column path_stamp already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add path_contract_terms column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'path_contract_terms');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `path_contract_terms` varchar(500) DEFAULT NULL COMMENT ''Path to contract terms JSON file'' AFTER `path_stamp`',
  'SELECT ''Column path_contract_terms already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add id_bank_account column
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'id_bank_account');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `users` ADD COLUMN `id_bank_account` int unsigned DEFAULT NULL COMMENT ''Foreign key to banks table'' AFTER `path_contract_terms`',
  'SELECT ''Column id_bank_account already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraint for id_bank_account if it doesn't exist
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND COLUMN_NAME = 'id_bank_account'
  AND REFERENCED_TABLE_NAME = 'banks');
SET @sql = IF(@fk_exists = 0, 
  'ALTER TABLE `users` ADD CONSTRAINT `fk_users_bank_account` FOREIGN KEY (`id_bank_account`) REFERENCES `banks` (`id`) ON DELETE SET NULL',
  'SELECT ''Foreign key fk_users_bank_account already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index on id_bank_account for faster lookups
SET @idx_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'users' 
  AND INDEX_NAME = 'idx_id_bank_account');
SET @sql = IF(@idx_exists = 0, 
  'ALTER TABLE `users` ADD INDEX `idx_id_bank_account` (`id_bank_account`)',
  'SELECT ''Index idx_id_bank_account already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- 2. Add company_address to banks table
-- ============================================

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'banks' 
  AND COLUMN_NAME = 'company_address');
SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `banks` ADD COLUMN `company_address` text COLLATE latin1_general_ci DEFAULT NULL COMMENT ''Company address for bank account'' AFTER `bank_address`',
  'SELECT ''Column company_address already exists'' AS message');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

