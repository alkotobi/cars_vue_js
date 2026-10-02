-- ============================================
-- Migration: API tokens on users
-- ============================================
-- The app had no server-side credential: LoginView fetched the user row
-- (password hash included) through the generic query endpoint and verified the
-- password in the browser, so api.php could not tell a real user from anyone
-- posting { user_id: 1, is_admin: true }.
--
-- This adds the one column the login action needs to hand out a token the
-- server can then check. One token per user, stored in the clear: it is only as
-- sensitive as a password, it is never rendered in the UI, and storing a hash
-- would buy nothing here without an expiry to enforce.
--
-- Idempotent: re-running on a server that already has the column is a no-op.

SET @db_name = DATABASE();

-- ============================================
-- Step 1: Add api_token
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'users'
  AND COLUMN_NAME = 'api_token');

SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `users` ADD COLUMN `api_token` VARCHAR(64) DEFAULT NULL COMMENT ''Token sent by useApi; checked server-side by lib/auth.php'' AFTER `password`',
  'SELECT ''Column api_token already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 2: Index it (token lookups are by value)
-- ============================================
SET @index_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'users'
  AND INDEX_NAME = 'idx_api_token');

SET @sql = IF(@index_exists = 0,
  'ALTER TABLE `users` ADD INDEX `idx_api_token` (`api_token`)',
  'SELECT ''Index idx_api_token already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
