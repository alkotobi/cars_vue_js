-- ============================================
-- Migration: enforce car-name delete guards in the database
-- ============================================
-- CarModelsView refuses to delete a brand or car name that is still referenced,
-- and lists what is referencing it. That check lives in the browser, which is not
-- a trust boundary here: api/api.php runs caller-supplied SQL with no
-- authentication (see SECURITY.md), so DELETE FROM cars_names WHERE id = ? can be
-- sent directly and the browser check is simply skipped.
--
-- Observed on a scratch copy with the client check bypassed:
--   DELETE FROM cars_names WHERE id = 4  ->  succeeded, no error
--   buy_details rows still pointing at id 4 afterwards: 1
--   car_name_media rows for id 4: 0   (silently cascaded away, files orphaned)
--
-- This migration moves both guarantees into the schema so they hold regardless of
-- who calls the API:
--
--   1. buy_details.id_car_name gains a foreign key with ON DELETE RESTRICT. It
--      currently has none, which is why the delete above left a dangling row.
--   2. car_name_media.car_name_id changes from ON DELETE CASCADE to RESTRICT, so
--      deleting a car name can no longer wipe media rows and strand the files.
--
-- RESTRICT rather than CASCADE on purpose. Cascading a car-name delete destroys
-- purchase history and media references, and the files on disk would survive the
-- rows either way, so the correct action is to refuse and make a human decide.
--
-- Idempotent, and safe to run on a database that already has these constraints:
-- the existing foreign keys are discovered from information_schema rather than
-- dropped by a hardcoded name, because a database built through
-- migration_v19_v25.sql may carry different constraint names.

-- ---------------------------------------------------------------------------
-- Step 1: clear orphaned buy_details.id_car_name values
--
-- Required before the foreign key can be added. The column is nullable, so
-- clearing the link keeps the purchase row and loses only a reference that was
-- already pointing at nothing. The SELECTs report what was touched so this is
-- visible in the migration output instead of happening silently.
-- ---------------------------------------------------------------------------
SELECT COUNT(*) AS `orphaned_buy_details_cleared`
FROM `buy_details`
WHERE `id_car_name` IS NOT NULL
  AND `id_car_name` NOT IN (SELECT `id` FROM `cars_names`);

UPDATE `buy_details`
SET `id_car_name` = NULL
WHERE `id_car_name` IS NOT NULL
  AND `id_car_name` NOT IN (SELECT `id` FROM `cars_names`);

SELECT COUNT(*) AS `orphaned_buy_details_remaining`
FROM `buy_details`
WHERE `id_car_name` IS NOT NULL
  AND `id_car_name` NOT IN (SELECT `id` FROM `cars_names`);

-- ---------------------------------------------------------------------------
-- Step 2: buy_details.id_car_name -> cars_names(id) ON DELETE RESTRICT
--
-- The child column needs an index. MySQL would create one implicitly, but doing
-- it explicitly keeps the schema identical whether or not it did.
-- ---------------------------------------------------------------------------
SET @idx_exists := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'buy_details'
    AND COLUMN_NAME = 'id_car_name'
);
SET @sql := IF(
  @idx_exists > 0,
  'SELECT ''buy_details.id_car_name index already present'' AS `info`',
  'ALTER TABLE `buy_details` ADD INDEX `idx_buy_details_id_car_name` (`id_car_name`)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Drop any existing foreign key on this column, whatever it is named.
SET @fk_name := (
  SELECT CONSTRAINT_NAME FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'buy_details'
    AND REFERENCED_TABLE_NAME = 'cars_names'
  LIMIT 1
);
SET @sql := IF(
  @fk_name IS NULL,
  'SELECT ''no existing buy_details -> cars_names foreign key'' AS `info`',
  CONCAT('ALTER TABLE `buy_details` DROP FOREIGN KEY `', @fk_name, '`')
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `buy_details`
  ADD CONSTRAINT `fk_buy_details_car_name`
  FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`)
  ON DELETE RESTRICT;

-- ---------------------------------------------------------------------------
-- Step 3: car_name_media.car_name_id CASCADE -> RESTRICT
--
-- Same discovery pattern. The column is NOT NULL and already validated by the
-- existing constraint, so there is nothing to clean up first.
-- ---------------------------------------------------------------------------
SET @fk_name := (
  SELECT CONSTRAINT_NAME FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'car_name_media'
    AND REFERENCED_TABLE_NAME = 'cars_names'
  LIMIT 1
);
SET @sql := IF(
  @fk_name IS NULL,
  'SELECT ''no existing car_name_media -> cars_names foreign key'' AS `info`',
  CONCAT('ALTER TABLE `car_name_media` DROP FOREIGN KEY `', @fk_name, '`')
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE `car_name_media`
  ADD CONSTRAINT `fk_car_name_media_car_name`
  FOREIGN KEY (`car_name_id`) REFERENCES `cars_names` (`id`)
  ON DELETE RESTRICT;

-- ---------------------------------------------------------------------------
-- Step 4: report the result
-- ---------------------------------------------------------------------------
SELECT rc.TABLE_NAME AS `table_name`,
       rc.CONSTRAINT_NAME AS `constraint_name`,
       rc.DELETE_RULE AS `on_delete`
FROM information_schema.REFERENTIAL_CONSTRAINTS rc
WHERE rc.CONSTRAINT_SCHEMA = DATABASE()
  AND rc.REFERENCED_TABLE_NAME = 'cars_names'
ORDER BY rc.TABLE_NAME;
