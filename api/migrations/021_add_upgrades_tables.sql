-- ============================================
-- Migration: Add upgrades and car upgrades tables
-- ============================================
-- This migration creates:
-- 1. upgrades table - stores upgrade types/descriptions
-- 2. car_apgrades table - stores car upgrades with values and dates

SET @db_name = DATABASE();

-- ============================================
-- Step 1: Create upgrades table
-- ============================================
SET @table_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades');

SET @sql = IF(@table_exists = 0,
  'CREATE TABLE `upgrades` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `description` varchar(255) DEFAULT NULL,
    `notes` text,
    `id_user_owner` int DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_id_user_owner` (`id_user_owner`),
    CONSTRAINT `fk_upgrades_user_owner` FOREIGN KEY (`id_user_owner`) REFERENCES `users` (`id`) ON DELETE SET NULL
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci',
  'SELECT ''Table upgrades already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 1.1: Add id_user_owner column to upgrades if table exists but column is missing
-- ============================================
-- Re-check if upgrades table exists (it might have been created in Step 1)
SET @upgrades_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades');

-- Add id_user_owner column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades' 
  AND COLUMN_NAME = 'id_user_owner');

SET @sql = IF(@col_exists = 0 AND @upgrades_exists > 0,
  'ALTER TABLE `upgrades` ADD COLUMN `id_user_owner` int DEFAULT NULL AFTER `notes`',
  'SELECT ''Column id_user_owner already exists or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index for id_user_owner if column exists but index doesn't
SET @index_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades' 
  AND INDEX_NAME = 'idx_id_user_owner');

SET @col_exists_for_index = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades' 
  AND COLUMN_NAME = 'id_user_owner');

SET @sql = IF(@index_exists = 0 AND @col_exists_for_index > 0 AND @upgrades_exists > 0,
  'ALTER TABLE `upgrades` ADD KEY `idx_id_user_owner` (`id_user_owner`)',
  'SELECT ''Index idx_id_user_owner already exists, column does not exist, or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraint for id_user_owner if it doesn't exist
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades' 
  AND CONSTRAINT_NAME = 'fk_upgrades_user_owner');

SET @col_exists_for_fk = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'upgrades' 
  AND COLUMN_NAME = 'id_user_owner');

SET @sql = IF(@fk_exists = 0 AND @col_exists_for_fk > 0 AND @upgrades_exists > 0,
  'ALTER TABLE `upgrades` ADD CONSTRAINT `fk_upgrades_user_owner` FOREIGN KEY (`id_user_owner`) REFERENCES `users` (`id`) ON DELETE SET NULL',
  'SELECT ''Foreign key fk_upgrades_user_owner already exists, column does not exist, or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 2: Create car_apgrades table
-- ============================================
SET @table_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades');

SET @sql = IF(@table_exists = 0,
  'CREATE TABLE `car_apgrades` (
    `id` int unsigned NOT NULL AUTO_INCREMENT,
    `id_car` int DEFAULT NULL,
    `id_upgrade` int DEFAULT NULL,
    `value` float NOT NULL,
    `date_done` datetime DEFAULT NULL,
    `id_uder_done` int DEFAULT NULL,
    `id_user_create` int DEFAULT NULL,
    `time_creation` timestamp NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_id_car` (`id_car`),
    KEY `idx_id_upgrade` (`id_upgrade`),
    KEY `idx_id_uder_done` (`id_uder_done`),
    KEY `idx_id_user_create` (`id_user_create`),
    CONSTRAINT `fk_car_apgrades_car` FOREIGN KEY (`id_car`) REFERENCES `cars_stock` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_car_apgrades_upgrade` FOREIGN KEY (`id_upgrade`) REFERENCES `upgrades` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_car_apgrades_user` FOREIGN KEY (`id_uder_done`) REFERENCES `users` (`id`) ON DELETE SET NULL,
    CONSTRAINT `fk_car_apgrades_user_create` FOREIGN KEY (`id_user_create`) REFERENCES `users` (`id`) ON DELETE SET NULL
  ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci',
  'SELECT ''Table car_apgrades already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 3: Add new columns to car_apgrades if table exists but columns are missing
-- ============================================
-- Re-check if car_apgrades table exists (it might have been created in Step 2)
SET @car_apgrades_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades');

-- Add id_user_create column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND COLUMN_NAME = 'id_user_create');

SET @sql = IF(@col_exists = 0 AND @car_apgrades_exists > 0,
  'ALTER TABLE `car_apgrades` ADD COLUMN `id_user_create` int DEFAULT NULL AFTER `id_uder_done`',
  'SELECT ''Column id_user_create already exists or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add index for id_user_create if column exists but index doesn't
SET @index_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND INDEX_NAME = 'idx_id_user_create');

SET @col_exists_for_index = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND COLUMN_NAME = 'id_user_create');

SET @sql = IF(@index_exists = 0 AND @col_exists_for_index > 0 AND @car_apgrades_exists > 0,
  'ALTER TABLE `car_apgrades` ADD KEY `idx_id_user_create` (`id_user_create`)',
  'SELECT ''Index idx_id_user_create already exists, column does not exist, or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add foreign key constraint for id_user_create if it doesn't exist
SET @fk_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND CONSTRAINT_NAME = 'fk_car_apgrades_user_create');

SET @col_exists_for_fk = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND COLUMN_NAME = 'id_user_create');

SET @sql = IF(@fk_exists = 0 AND @col_exists_for_fk > 0 AND @car_apgrades_exists > 0,
  'ALTER TABLE `car_apgrades` ADD CONSTRAINT `fk_car_apgrades_user_create` FOREIGN KEY (`id_user_create`) REFERENCES `users` (`id`) ON DELETE SET NULL',
  'SELECT ''Foreign key fk_car_apgrades_user_create already exists, column does not exist, or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Add time_creation column if it doesn't exist
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'car_apgrades' 
  AND COLUMN_NAME = 'time_creation');

SET @sql = IF(@col_exists = 0 AND @car_apgrades_exists > 0,
  'ALTER TABLE `car_apgrades` ADD COLUMN `time_creation` timestamp NULL DEFAULT NULL AFTER `id_user_create`',
  'SELECT ''Column time_creation already exists or table does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

