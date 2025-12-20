-- ============================================
-- Migration: Add deadline, status change tracking, and comments for selections
-- ============================================
-- This migration adds:
-- 1. Deadline field (renaming/keeping due_date as deadline)
-- 2. Status change tracking (status_changed_at, previous_status)
-- 3. Selection comments table for holders to share thoughts

SET @db_name = DATABASE();

-- ============================================
-- 1. Add status change tracking fields to car_selections
-- ============================================
ALTER TABLE `car_selections`
ADD COLUMN `status_changed_at` datetime DEFAULT NULL COMMENT 'When the status was last changed' AFTER `status`,
ADD COLUMN `previous_status` enum('pending', 'in_progress', 'completed', 'cancelled') DEFAULT NULL COMMENT 'Previous status before change' AFTER `status_changed_at`,
ADD COLUMN `deadline` datetime DEFAULT NULL COMMENT 'Deadline for completing the selection' AFTER `due_date`,
ADD KEY `idx_status_changed_at` (`status_changed_at`),
ADD KEY `idx_deadline` (`deadline`);

-- Update existing records: set status_changed_at to created_at for records with pending status
UPDATE `car_selections` 
SET `status_changed_at` = `created_at` 
WHERE `status_changed_at` IS NULL;

-- ============================================
-- 2. Create selection_comments table (if it doesn't exist)
-- ============================================
CREATE TABLE IF NOT EXISTS `selection_comments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `selection_id` int(11) NOT NULL COMMENT 'FK to car_selections - the selection being commented on',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - who wrote the comment',
  `comment` text NOT NULL COMMENT 'The comment text',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_selection_id` (`selection_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_comments_selection` FOREIGN KEY (`selection_id`) REFERENCES `car_selections` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Comments on selections - allows holders to share thoughts';

