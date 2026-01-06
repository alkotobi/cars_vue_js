-- ============================================
-- Migration: Add edited_by JSON column to sell_bill
-- ============================================
-- This migration adds:
-- 1. edited_by (JSON) - Tracks who edited the sell bill and when
--    Format: [{"id_user": 1, "timestamp": "2025-01-15 10:30:00"}, ...]

SET @db_name = DATABASE();

-- ============================================
-- Add edited_by column
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'sell_bill' 
  AND COLUMN_NAME = 'edited_by');

SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `sell_bill` ADD COLUMN `edited_by` JSON DEFAULT NULL COMMENT ''JSON array tracking edits: [{"id_user": int, "timestamp": "datetime"}]''',
  'SELECT ''Column edited_by already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

