-- ============================================
-- Migration: store the model's court-records answer
-- ============================================
-- The credibility check asks two questions, and the second one is whether the
-- supplier has any court cases. The answer is a paragraph, so it needs its own
-- column: folding it into the summary field would lose the ability to show the
-- two questions side by side, and would make the history list unreadable.
--
-- There is no data source behind this column. The model has no court register
-- access, so what lands here is either its own labelled recollection or an
-- explicit "I cannot check this", never a lookup. The column comment says so, so
-- the next person to read the schema is not misled into thinking otherwise.
--
-- Idempotent: re-running on a server that already has the column is a no-op.

SET @db_name = DATABASE();

SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'supplier_credibility_checks'
  AND COLUMN_NAME = 'court_records');

SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `supplier_credibility_checks` ADD COLUMN `court_records` text DEFAULT NULL AFTER `summary`',
  'SELECT ''Column court_records already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Which basis the model claimed for its court answer. Only 'recollection' and
-- 'no_information' ever reach this table; anything else is normalised away
-- first. The modal labels a claimed recollection, so a reader never mistakes it
-- for a lookup result.
SET @col_exists = (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
  WHERE TABLE_SCHEMA = @db_name
  AND TABLE_NAME = 'supplier_credibility_checks'
  AND COLUMN_NAME = 'courts_basis');

SET @sql = IF(@col_exists = 0,
  'ALTER TABLE `supplier_credibility_checks` ADD COLUMN `courts_basis` varchar(20) DEFAULT NULL AFTER `court_records`',
  'SELECT ''Column courts_basis already exists'' AS message');

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
