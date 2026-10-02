-- ============================================
-- Migration: add foreign key on cars_names.id_brand with ON DELETE RESTRICT
-- ============================================
-- CarModelsView already refuses to delete a brand that has car names (it looks up
-- usage in cars_names WHERE id_brand = ?), but that check is browser-side. The
-- API in this project executes caller-supplied SQL without authentication (see
-- SECURITY.md), so DELETE FROM brands WHERE id = ? can be sent directly and the
-- browser check is skipped.
--
-- This mirrors 028: move the invariant to the schema so the database refuses the
-- delete regardless of caller. Chosen RESTRICT because:
--   - matches the existing UI behaviour,
--   - keeps the action non-destructive, and
--   - the UI does not currently support "models become unassigned" (Add/Edit Car
--     Name requires a brand and the <select> is populated from brands).
--
-- cars_names.id_brand is nullable (LEFT JOIN usage), and on the live database
-- all 35 car names have a brand assigned with 0 dangling references. Other
-- databases may differ, so we defensively null out any dangling values and
-- print how many were touched.
--
-- The FK is added idempotently: existing constraints (if any) are discovered by
-- information_schema rather than dropped by a hardcoded name, and the index on
-- id_brand is added explicitly to keep the schema identical between a fresh
-- install and an upgraded one.

-- ---------------------------------------------------------------------------
-- Step 1: clear dangling id_brand values
-- ---------------------------------------------------------------------------
SELECT COUNT(*) AS `dangling_id_brand_cleared`
FROM `cars_names`
WHERE `id_brand` IS NOT NULL
  AND `id_brand` NOT IN (SELECT `id` FROM `brands`);

UPDATE `cars_names` cn
LEFT JOIN `brands` b ON b.`id` = cn.`id_brand`
SET cn.`id_brand` = NULL
WHERE cn.`id_brand` IS NOT NULL
  AND b.`id` IS NULL;

SELECT COUNT(*) AS `dangling_id_brand_remaining`
FROM `cars_names`
WHERE `id_brand` IS NOT NULL
  AND `id_brand` NOT IN (SELECT `id` FROM `brands`);

-- ---------------------------------------------------------------------------
-- Step 2: ensure index on cars_names.id_brand
-- ---------------------------------------------------------------------------
SET @idx_exists := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'cars_names'
    AND COLUMN_NAME = 'id_brand'
);
SET @sql := IF(
  @idx_exists > 0,
  'SELECT ''cars_names.id_brand index already present'' AS `info`',
  'ALTER TABLE `cars_names` ADD INDEX `idx_cars_names_id_brand` (`id_brand`)'
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- Step 3: drop any existing FK on cars_names.id_brand -> brands(id)
-- ---------------------------------------------------------------------------
SET @fk_name := (
  SELECT CONSTRAINT_NAME FROM information_schema.REFERENTIAL_CONSTRAINTS
  WHERE CONSTRAINT_SCHEMA = DATABASE()
    AND TABLE_NAME = 'cars_names'
    AND REFERENCED_TABLE_NAME = 'brands'
  LIMIT 1
);
SET @sql := IF(
  @fk_name IS NULL,
  'SELECT ''no existing cars_names -> brands foreign key'' AS `info`',
  CONCAT('ALTER TABLE `cars_names` DROP FOREIGN KEY `', @fk_name, '`')
);
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ---------------------------------------------------------------------------
-- Step 4: add FK with ON DELETE RESTRICT
-- ---------------------------------------------------------------------------
ALTER TABLE `cars_names`
  ADD CONSTRAINT `fk_cars_names_brand`
  FOREIGN KEY (`id_brand`) REFERENCES `brands` (`id`)
  ON DELETE RESTRICT;

-- ---------------------------------------------------------------------------
-- Step 5: report the result
-- ---------------------------------------------------------------------------
SELECT rc.TABLE_NAME AS `table_name`,
       rc.CONSTRAINT_NAME AS `constraint_name`,
       rc.DELETE_RULE AS `on_delete`
FROM information_schema.REFERENTIAL_CONSTRAINTS rc
WHERE rc.CONSTRAINT_SCHEMA = DATABASE()
  AND rc.REFERENCED_TABLE_NAME = 'brands'
ORDER BY rc.TABLE_NAME;
