-- ============================================
-- Migration: Convert cars_stock notes column to JSON
-- ============================================
-- This migration converts the notes column from TEXT to JSON
-- Format: [{"id_user": int, "note": "text", "timestamp": "datetime"}, ...]
-- Preserves all existing notes data by wrapping them in the new JSON structure

SET @db_name = DATABASE();

-- ============================================
-- Step 1: Add temporary column for JSON notes
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'cars_stock' 
  AND COLUMN_NAME = 'notes_json');

SET @sql = IF(@col_exists = 0, 
  'ALTER TABLE `cars_stock` ADD COLUMN `notes_json` JSON DEFAULT NULL COMMENT ''JSON array of notes: [{"id_user": int, "note": "text", "timestamp": "datetime"}]'' AFTER `notes`',
  'SELECT ''Column notes_json already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 2: Migrate existing notes data to JSON format
-- ============================================
-- Convert existing notes to JSON array format
-- If notes exist, create a JSON array with one entry (id_user will be 1 for old notes)
UPDATE `cars_stock`
SET `notes_json` = CASE
  WHEN `notes` IS NOT NULL AND `notes` != '' THEN
    JSON_ARRAY(
      JSON_OBJECT(
        'id_user', 1,
        'note', `notes`,
        'timestamp', NOW()
      )
    )
  ELSE
    NULL
END
WHERE `notes_json` IS NULL;

-- ============================================
-- Step 3: Drop old notes column
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'cars_stock' 
  AND COLUMN_NAME = 'notes');

SET @sql = IF(@col_exists > 0, 
  'ALTER TABLE `cars_stock` DROP COLUMN `notes`',
  'SELECT ''Column notes does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 4: Rename notes_json to notes
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS 
  WHERE TABLE_SCHEMA = @db_name 
  AND TABLE_NAME = 'cars_stock' 
  AND COLUMN_NAME = 'notes_json');

SET @sql = IF(@col_exists > 0, 
  'ALTER TABLE `cars_stock` CHANGE COLUMN `notes_json` `notes` JSON DEFAULT NULL COMMENT ''JSON array of notes: [{"id_user": int, "note": "text", "timestamp": "datetime"}]''',
  'SELECT ''Column notes_json does not exist'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

