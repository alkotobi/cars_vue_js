-- ============================================
-- Migration: API tokens on the DB-manager login table
-- ============================================
-- TARGET DATABASE: the registry (merhab_databases), NOT the app/tenant database.
-- Every other migration in this folder targets the tenant database, so this one is
-- the exception:
--
--   mysql -u USER -p merhab_databases < api/migrations/031_login_api_token.sql
--
-- `login` lives in the registry, alongside `dbs`. That is the whole point of it:
-- the DB manager is a separate security realm from the app, holding host-level
-- credentials, and authenticates against this table rather than the tenant's
-- `users` table (see the gate in db_manager_api.php).
--
-- Why this is needed at all: db_manager_api.php used to have no authentication,
-- so run_sql, delete_database, backup_databases, prepare_upload_folder and the
-- rest were reachable by an anonymous POST. The obvious fix was to gate them on
-- the tenant admin token from lib/auth.php, but that put `login` and `signup`
-- behind the gate too - and those are the only way to obtain the credential the
-- UI requires. The DB manager became unreachable for everyone, including admins.
-- So it needed a credential of its own, and this is the column that carries it.
--
-- One token per user, stored in the clear, same reasoning as
-- 021_users_api_token.sql: it is only as sensitive as the `pass` column beside
-- it, it is never rendered in the UI, and hashing it buys nothing without an
-- expiry to enforce.
--
-- Idempotent: re-running on a server that already has the column is a no-op.

SET @db_name = DATABASE();

-- ============================================
-- Step 1: Add api_token
-- ============================================
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'login'
  AND COLUMN_NAME = 'api_token');

SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `login` ADD COLUMN `api_token` VARCHAR(64) DEFAULT NULL COMMENT ''Token sent by the db-manager UI; checked server-side in db_manager_api.php'' AFTER `pass`',
  'SELECT ''Column api_token already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ============================================
-- Step 2: Index it (token lookups are by value)
-- ============================================
SET @index_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.STATISTICS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'login'
  AND INDEX_NAME = 'idx_login_api_token');

SET @sql = IF(@index_exists = 0,
  'ALTER TABLE `login` ADD INDEX `idx_login_api_token` (`api_token`)',
  'SELECT ''Index idx_login_api_token already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;