-- ============================================
-- Migration: Car Selections System
-- ============================================
-- This migration creates tables for saving car selections,
-- team management, jobs, and selection tracking

SET @db_name = DATABASE();

-- ============================================
-- 1. Create jobs table
-- ============================================
CREATE TABLE IF NOT EXISTS `jobs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Job name',
  `description` text DEFAULT NULL COMMENT 'Job description',
  `category` varchar(100) DEFAULT NULL COMMENT 'Job category (loading, delivery, inspection, documentation, etc.)',
  `estimated_duration_hours` int(11) DEFAULT NULL COMMENT 'Estimated duration in hours',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `team_id` int(11) DEFAULT NULL COMMENT 'FK to teams - each job belongs to one team (one team can have multiple jobs)',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_category` (`category`),
  KEY `idx_name` (`name`),
  KEY `idx_team_id` (`team_id`),
  CONSTRAINT `fk_jobs_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Jobs assigned to teams - one team can have multiple jobs, each job belongs to one team';

-- ============================================
-- 2. Create teams table
-- ============================================
CREATE TABLE IF NOT EXISTS `teams` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Team name',
  `team_leader_id` int(11) NOT NULL COMMENT 'FK to users - team leader',
  `description` text DEFAULT NULL COMMENT 'Team description',
  `jobs_completed_count` int(11) NOT NULL DEFAULT 0 COMMENT 'Incremental count of completed jobs',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_team_leader_id` (`team_leader_id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_jobs_completed_count` (`jobs_completed_count`),
  CONSTRAINT `fk_teams_team_leader` FOREIGN KEY (`team_leader_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Teams with team leaders';

-- ============================================
-- 3. Create team_members table (users can only be in one team)
-- ============================================
CREATE TABLE IF NOT EXISTS `team_members` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `team_id` int(11) NOT NULL COMMENT 'FK to teams',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - UNIQUE constraint ensures one team per user',
  `role` enum('member', 'deputy_leader') NOT NULL DEFAULT 'member' COMMENT 'Role in the team',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'When user joined the team',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_user_team` (`user_id`),
  KEY `idx_team_id` (`team_id`),
  KEY `idx_user_id` (`user_id`),
  CONSTRAINT `fk_team_members_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_team_members_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Team members - users can only be in one team at a time';

-- ============================================
-- 4. Create car_selections table
-- ============================================
CREATE TABLE IF NOT EXISTS `car_selections` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Selection name',
  `description` text DEFAULT NULL COMMENT 'Selection description',
  `user_create_selection` int(11) NOT NULL COMMENT 'FK to users - who created the selection',
  `selection_data` json DEFAULT NULL COMMENT 'JSON array of car IDs that were selected',
  `assigned_to_team` int(11) DEFAULT NULL COMMENT 'FK to teams - team assigned to work on this',
  `assigned_to_team_from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who assigned to team',
  `assigned_at` datetime DEFAULT NULL COMMENT 'When assigned to team',
  `job_id` int(11) DEFAULT NULL COMMENT 'FK to jobs - what job the team needs to do',
  `job_done_on` datetime DEFAULT NULL COMMENT 'DateTime when job was completed',
  `owned_by` json DEFAULT NULL COMMENT 'JSON array of user IDs - users who received this selection',
  `sent_by_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who sent the selection',
  `status` enum('pending', 'in_progress', 'completed', 'cancelled') NOT NULL DEFAULT 'pending' COMMENT 'Selection status',
  `status_changed_at` datetime DEFAULT NULL COMMENT 'When the status was last changed',
  `previous_status` enum('pending', 'in_progress', 'completed', 'cancelled') DEFAULT NULL COMMENT 'Previous status before change',
  `priority` enum('low', 'medium', 'high', 'urgent') NOT NULL DEFAULT 'medium' COMMENT 'Priority level',
  `due_date` datetime DEFAULT NULL COMMENT 'Optional deadline',
  `deadline` datetime DEFAULT NULL COMMENT 'Deadline for completing the selection',
  `notes` text DEFAULT NULL COMMENT 'Additional notes/comments',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_create_selection` (`user_create_selection`),
  KEY `idx_assigned_to_team` (`assigned_to_team`),
  KEY `idx_assigned_to_team_from_user_id` (`assigned_to_team_from_user_id`),
  KEY `idx_job_id` (`job_id`),
  KEY `idx_sent_by_user_id` (`sent_by_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_status_changed_at` (`status_changed_at`),
  KEY `idx_priority` (`priority`),
  KEY `idx_due_date` (`due_date`),
  KEY `idx_deadline` (`deadline`),
  KEY `idx_assigned_at` (`assigned_at`),
  KEY `idx_job_done_on` (`job_done_on`),
  CONSTRAINT `fk_car_selections_creator` FOREIGN KEY (`user_create_selection`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_team` FOREIGN KEY (`assigned_to_team`) REFERENCES `teams` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_assigner` FOREIGN KEY (`assigned_to_team_from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_job` FOREIGN KEY (`job_id`) REFERENCES `jobs` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_car_selections_sender` FOREIGN KEY (`sent_by_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Car selections with team assignments and job tracking';

-- ============================================
-- 5. Create selection_ownership_history table
-- ============================================
CREATE TABLE IF NOT EXISTS `selection_ownership_history` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `selection_id` int(11) NOT NULL COMMENT 'FK to car_selections',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - user involved in the action',
  `action` enum('sent', 'received', 'transferred') NOT NULL COMMENT 'Type of action',
  `from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who sent (nullable)',
  `to_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who received (nullable)',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_selection_id` (`selection_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_from_user_id` (`from_user_id`),
  KEY `idx_to_user_id` (`to_user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_ownership_history_selection` FOREIGN KEY (`selection_id`) REFERENCES `car_selections` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE,
  CONSTRAINT `fk_ownership_history_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Track ownership changes and transfers of selections';

-- ============================================
-- 6. Create selection_comments table
-- ============================================
CREATE TABLE IF NOT EXISTS `selection_comments` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `selection_id` int(11) NOT NULL COMMENT 'FK to car_selections',
  `user_id` int(11) NOT NULL COMMENT 'FK to users - who commented',
  `comment` text NOT NULL COMMENT 'Comment text',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_selection_id` (`selection_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_created_at` (`created_at`),
  CONSTRAINT `fk_comments_selection` FOREIGN KEY (`selection_id`) REFERENCES `car_selections` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_comments_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='Comments and updates on selections';

-- ============================================
-- 7. Create trigger to increment jobs_completed_count
-- ============================================
DELIMITER $$

DROP TRIGGER IF EXISTS trg_increment_team_jobs_completed$$

CREATE TRIGGER trg_increment_team_jobs_completed
AFTER UPDATE ON `car_selections`
FOR EACH ROW
BEGIN
  -- Only increment if status changed to 'completed' and job_done_on is set
  IF NEW.status = 'completed' 
     AND OLD.status != 'completed' 
     AND NEW.job_done_on IS NOT NULL
     AND NEW.assigned_to_team IS NOT NULL THEN
    UPDATE `teams`
    SET `jobs_completed_count` = `jobs_completed_count` + 1
    WHERE `id` = NEW.assigned_to_team;
  END IF;
END$$

DELIMITER ;

-- ============================================
-- 8. Insert default jobs
-- ============================================
INSERT IGNORE INTO `jobs` (`name`, `description`, `category`, `is_active`) VALUES
('Loading', 'Load cars onto container/ship', 'loading', 1),
('Delivery', 'Deliver cars to destination', 'delivery', 1),
('Inspection', 'Inspect cars for quality/condition', 'inspection', 1),
('Documentation', 'Prepare and process documents', 'documentation', 1),
('Warehouse Management', 'Manage warehouse operations', 'warehouse', 1),
('Custom Clearance', 'Handle custom clearance procedures', 'custom_clearance', 1);

