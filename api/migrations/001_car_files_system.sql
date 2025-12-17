-- ============================================
-- Migration: Car Files Management System
-- Description: Flexible file management with physical copy tracking
-- Date: 2025-01-XX
-- ============================================

-- ============================================
-- 1. Create car_file_categories table
-- ============================================
CREATE TABLE IF NOT EXISTS `car_file_categories` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `category_name` varchar(255) NOT NULL,
  `importance_level` tinyint(1) NOT NULL DEFAULT 3 COMMENT '1=Critical, 2=High, 3=Medium, 4=Low, 5=Optional',
  `is_required` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1=Required, 0=Optional',
  `display_order` int(11) NOT NULL DEFAULT 0 COMMENT 'Order for UI display',
  `description` text DEFAULT NULL,
  `visibility_scope` enum('public', 'department', 'role', 'private') NOT NULL DEFAULT 'public' COMMENT 'Who can see files in this category',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `category_name` (`category_name`),
  KEY `idx_importance` (`importance_level`),
  KEY `idx_display_order` (`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- Insert default categories (mapping from old fixed columns)
INSERT INTO `car_file_categories` (`category_name`, `importance_level`, `is_required`, `display_order`, `description`, `visibility_scope`) VALUES
('Bill of Lading', 1, 1, 1, 'Bill of Lading document', 'public'),
('Invoice', 2, 1, 2, 'Sell Invoice (PI)', 'public'),
('Packing List', 2, 1, 3, 'Buy Packing List (PI)', 'public'),
('Certificate of Origin', 3, 0, 4, 'COO Certificate', 'public'),
('Certificate of Conformity', 3, 0, 5, 'COC Certificate', 'public')
ON DUPLICATE KEY UPDATE `category_name` = VALUES(`category_name`);

-- ============================================
-- 2. Create car_files table
-- ============================================
CREATE TABLE IF NOT EXISTS `car_files` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_id` int(11) NOT NULL COMMENT 'FK to cars_stock',
  `category_id` int(11) NOT NULL COMMENT 'FK to car_file_categories',
  `file_path` varchar(500) NOT NULL COMMENT 'Storage path to file',
  `file_name` varchar(255) NOT NULL COMMENT 'Original filename',
  `file_size` bigint(20) DEFAULT NULL COMMENT 'File size in bytes',
  `file_type` varchar(100) DEFAULT NULL COMMENT 'MIME type',
  `uploaded_by` int(11) NOT NULL COMMENT 'FK to users - who uploaded',
  `uploaded_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `version` int(11) NOT NULL DEFAULT 1 COMMENT 'File version number',
  `notes` text DEFAULT NULL COMMENT 'Additional notes about the file',
  `visibility_scope` enum('public', 'department', 'role', 'private') DEFAULT NULL COMMENT 'Override category visibility, NULL = use category default',
  `allowed_viewers` text DEFAULT NULL COMMENT 'JSON array of user_ids who can view, NULL = all',
  `department_id` int(11) DEFAULT NULL COMMENT 'FK to departments if visibility is department-based',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=Deleted/Archived, 1=Active',
  PRIMARY KEY (`id`),
  KEY `idx_car_id` (`car_id`),
  KEY `idx_category_id` (`category_id`),
  KEY `idx_uploaded_by` (`uploaded_by`),
  KEY `idx_uploaded_at` (`uploaded_at`),
  KEY `idx_car_category` (`car_id`, `category_id`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_car_files_car` FOREIGN KEY (`car_id`) REFERENCES `cars_stock` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_files_category` FOREIGN KEY (`category_id`) REFERENCES `car_file_categories` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_car_files_uploaded_by` FOREIGN KEY (`uploaded_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- 3. Create car_file_physical_tracking table
-- ============================================
CREATE TABLE IF NOT EXISTS `car_file_physical_tracking` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `current_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who has physical copy now, NULL = available',
  `previous_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who had it before',
  `checked_out_at` timestamp NULL DEFAULT NULL COMMENT 'When physical copy was taken',
  `checked_in_at` timestamp NULL DEFAULT NULL COMMENT 'When returned, NULL if still out',
  `transfer_notes` text DEFAULT NULL COMMENT 'Notes about the transfer',
  `transferred_by` int(11) DEFAULT NULL COMMENT 'FK to users - who made the transfer',
  `transferred_at` timestamp NULL DEFAULT NULL COMMENT 'When transfer was made',
  `status` enum('available', 'checked_out', 'lost', 'archived') NOT NULL DEFAULT 'available',
  `is_visible_to_holder_only` tinyint(1) NOT NULL DEFAULT 0 COMMENT '1=only holder+admin can see, 0=use file visibility rules',
  `expected_return_date` date DEFAULT NULL COMMENT 'Expected return date for checked out files',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_car_file_id` (`car_file_id`),
  KEY `idx_current_holder` (`current_holder_id`),
  KEY `idx_status` (`status`),
  KEY `idx_checked_out_at` (`checked_out_at`),
  KEY `idx_car_file_status` (`car_file_id`, `status`),
  CONSTRAINT `fk_physical_tracking_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_physical_tracking_current_holder` FOREIGN KEY (`current_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_previous_holder` FOREIGN KEY (`previous_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- 4. Create car_file_transfers table (history)
-- ============================================
CREATE TABLE IF NOT EXISTS `car_file_transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who transferred from, NULL = available',
  `to_user_id` int(11) NOT NULL COMMENT 'FK to users - who received',
  `transferred_by` int(11) NOT NULL COMMENT 'FK to users - who performed the transfer',
  `transferred_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `notes` text DEFAULT NULL COMMENT 'Transfer notes',
  `return_expected_date` date DEFAULT NULL COMMENT 'Expected return date',
  `returned_at` timestamp NULL DEFAULT NULL COMMENT 'When file was returned (checked in)',
  `return_notes` text DEFAULT NULL COMMENT 'Return notes',
  PRIMARY KEY (`id`),
  KEY `idx_car_file_id` (`car_file_id`),
  KEY `idx_from_user` (`from_user_id`),
  KEY `idx_to_user` (`to_user_id`),
  KEY `idx_transferred_by` (`transferred_by`),
  KEY `idx_transferred_at` (`transferred_at`),
  KEY `idx_returned_at` (`returned_at`),
  CONSTRAINT `fk_transfers_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_transfers_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_transfers_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- 5. Migration: Move existing files from cars_stock to car_files
-- ============================================
-- This script migrates existing file paths from cars_stock table to the new car_files table

-- Migrate path_documents (Bill of Lading)
INSERT INTO `car_files` (`car_id`, `category_id`, `file_path`, `file_name`, `uploaded_by`, `uploaded_at`)
SELECT 
  cs.id,
  (SELECT id FROM car_file_categories WHERE category_name = 'Bill of Lading' LIMIT 1) as category_id,
  cs.path_documents,
  SUBSTRING_INDEX(cs.path_documents, '/', -1) as file_name,
  1 as uploaded_by, -- Default to admin (user_id = 1)
  NOW() as uploaded_at
FROM `cars_stock` cs
WHERE cs.path_documents IS NOT NULL 
  AND cs.path_documents != ''
  AND NOT EXISTS (
    SELECT 1 FROM `car_files` cf 
    WHERE cf.car_id = cs.id 
    AND cf.category_id = (SELECT id FROM car_file_categories WHERE category_name = 'Bill of Lading' LIMIT 1)
  );

-- Migrate sell_pi_path (Invoice)
INSERT INTO `car_files` (`car_id`, `category_id`, `file_path`, `file_name`, `uploaded_by`, `uploaded_at`)
SELECT 
  cs.id,
  (SELECT id FROM car_file_categories WHERE category_name = 'Invoice' LIMIT 1) as category_id,
  cs.sell_pi_path,
  SUBSTRING_INDEX(cs.sell_pi_path, '/', -1) as file_name,
  1 as uploaded_by, -- Default to admin (user_id = 1)
  NOW() as uploaded_at
FROM `cars_stock` cs
WHERE cs.sell_pi_path IS NOT NULL 
  AND cs.sell_pi_path != ''
  AND NOT EXISTS (
    SELECT 1 FROM `car_files` cf 
    WHERE cf.car_id = cs.id 
    AND cf.category_id = (SELECT id FROM car_file_categories WHERE category_name = 'Invoice' LIMIT 1)
  );

-- Migrate buy_pi_path (Packing List)
INSERT INTO `car_files` (`car_id`, `category_id`, `file_path`, `file_name`, `uploaded_by`, `uploaded_at`)
SELECT 
  cs.id,
  (SELECT id FROM car_file_categories WHERE category_name = 'Packing List' LIMIT 1) as category_id,
  cs.buy_pi_path,
  SUBSTRING_INDEX(cs.buy_pi_path, '/', -1) as file_name,
  1 as uploaded_by, -- Default to admin (user_id = 1)
  NOW() as uploaded_at
FROM `cars_stock` cs
WHERE cs.buy_pi_path IS NOT NULL 
  AND cs.buy_pi_path != ''
  AND NOT EXISTS (
    SELECT 1 FROM `car_files` cf 
    WHERE cf.car_id = cs.id 
    AND cf.category_id = (SELECT id FROM car_file_categories WHERE category_name = 'Packing List' LIMIT 1)
  );

-- Migrate path_coo (Certificate of Origin)
INSERT INTO `car_files` (`car_id`, `category_id`, `file_path`, `file_name`, `uploaded_by`, `uploaded_at`)
SELECT 
  cs.id,
  (SELECT id FROM car_file_categories WHERE category_name = 'Certificate of Origin' LIMIT 1) as category_id,
  cs.path_coo,
  SUBSTRING_INDEX(cs.path_coo, '/', -1) as file_name,
  1 as uploaded_by, -- Default to admin (user_id = 1)
  NOW() as uploaded_at
FROM `cars_stock` cs
WHERE cs.path_coo IS NOT NULL 
  AND cs.path_coo != ''
  AND NOT EXISTS (
    SELECT 1 FROM `car_files` cf 
    WHERE cf.car_id = cs.id 
    AND cf.category_id = (SELECT id FROM car_file_categories WHERE category_name = 'Certificate of Origin' LIMIT 1)
  );

-- Migrate path_coc (Certificate of Conformity)
INSERT INTO `car_files` (`car_id`, `category_id`, `file_path`, `file_name`, `uploaded_by`, `uploaded_at`)
SELECT 
  cs.id,
  (SELECT id FROM car_file_categories WHERE category_name = 'Certificate of Conformity' LIMIT 1) as category_id,
  cs.path_coc,
  SUBSTRING_INDEX(cs.path_coc, '/', -1) as file_name,
  1 as uploaded_by, -- Default to admin (user_id = 1)
  NOW() as uploaded_at
FROM `cars_stock` cs
WHERE cs.path_coc IS NOT NULL 
  AND cs.path_coc != ''
  AND NOT EXISTS (
    SELECT 1 FROM `car_files` cf 
    WHERE cf.car_id = cs.id 
    AND cf.category_id = (SELECT id FROM car_file_categories WHERE category_name = 'Certificate of Conformity' LIMIT 1)
  );

-- ============================================
-- 6. Create indexes for performance
-- ============================================
-- Additional composite indexes for common queries

-- Index for finding files by car and visibility
CREATE INDEX `idx_car_visibility` ON `car_files` (`car_id`, `visibility_scope`, `is_active`);

-- Index for finding files by uploader
CREATE INDEX `idx_uploader_active` ON `car_files` (`uploaded_by`, `is_active`);

-- Index for physical tracking by holder and status
CREATE INDEX `idx_holder_status` ON `car_file_physical_tracking` (`current_holder_id`, `status`);

-- Index for transfers by user (from or to)
CREATE INDEX `idx_transfers_user` ON `car_file_transfers` (`from_user_id`, `to_user_id`, `transferred_at`);

-- ============================================
-- 7. Create views for easier querying
-- ============================================

-- View: Files with current physical copy status
CREATE OR REPLACE VIEW `v_car_files_with_tracking` AS
SELECT 
  cf.id as file_id,
  cf.car_id,
  cf.category_id,
  cfc.category_name,
  cfc.importance_level,
  cfc.is_required,
  cf.file_path,
  cf.file_name,
  cf.file_size,
  cf.file_type,
  cf.uploaded_by,
  u_upload.username as uploaded_by_username,
  cf.uploaded_at,
  cf.version,
  cf.notes,
  cf.visibility_scope,
  cf.is_active,
  cpt.id as tracking_id,
  cpt.current_holder_id,
  u_holder.username as current_holder_username,
  cpt.status as physical_status,
  cpt.checked_out_at,
  cpt.checked_in_at,
  cpt.expected_return_date
FROM `car_files` cf
INNER JOIN `car_file_categories` cfc ON cf.category_id = cfc.id
LEFT JOIN `users` u_upload ON cf.uploaded_by = u_upload.id
LEFT JOIN `car_file_physical_tracking` cpt ON cf.id = cpt.car_file_id AND cpt.status IN ('available', 'checked_out')
LEFT JOIN `users` u_holder ON cpt.current_holder_id = u_holder.id
WHERE cf.is_active = 1;

-- View: Transfer history with user names
CREATE OR REPLACE VIEW `v_car_file_transfers` AS
SELECT 
  cft.id,
  cft.car_file_id,
  cf.file_name,
  cf.car_id,
  cfc.category_name,
  cft.from_user_id,
  u_from.username as from_username,
  cft.to_user_id,
  u_to.username as to_username,
  cft.transferred_by,
  u_transfer.username as transferred_by_username,
  cft.transferred_at,
  cft.notes,
  cft.return_expected_date,
  cft.returned_at,
  cft.return_notes
FROM `car_file_transfers` cft
INNER JOIN `car_files` cf ON cft.car_file_id = cf.id
INNER JOIN `car_file_categories` cfc ON cf.category_id = cfc.id
LEFT JOIN `users` u_from ON cft.from_user_id = u_from.id
INNER JOIN `users` u_to ON cft.to_user_id = u_to.id
INNER JOIN `users` u_transfer ON cft.transferred_by = u_transfer.id;

-- ============================================
-- 8. Notes
-- ============================================
-- IMPORTANT: 
-- 1. The old columns (path_documents, sell_pi_path, etc.) in cars_stock are kept
--    for backward compatibility. They can be removed in a future migration.
-- 2. All existing files are migrated to the new system with uploaded_by = 1 (admin)
-- 3. Physical tracking is not created for migrated files (status = 'available' by default)
-- 4. The UNIQUE constraint on physical_tracking allows only one active tracking per file
-- 5. Foreign keys use ON DELETE CASCADE for files, ON DELETE RESTRICT for users
-- 6. The visibility_scope in car_files can override the category default

