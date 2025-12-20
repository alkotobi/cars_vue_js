-- ============================================
-- Migration: Add team_id to jobs table
-- ============================================
-- This migration adds team_id to jobs table to establish
-- one-to-many relationship: one team can have multiple jobs,
-- but each job belongs to only one team

SET @db_name = DATABASE();

-- ============================================
-- 1. Add team_id column to jobs table
-- ============================================
ALTER TABLE `jobs` 
ADD COLUMN `team_id` int(11) DEFAULT NULL COMMENT 'FK to teams - each job belongs to one team' AFTER `is_active`,
ADD KEY `idx_team_id` (`team_id`);

-- ============================================
-- 2. Add foreign key constraint
-- ============================================
ALTER TABLE `jobs`
ADD CONSTRAINT `fk_jobs_team` FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE SET NULL ON UPDATE CASCADE;

-- ============================================
-- 3. Add UNIQUE constraint to ensure each job belongs to only one team
-- Note: This allows NULL values (jobs without teams), but if team_id is set,
-- it must be unique per job. Since job.id is already unique, we don't need
-- a separate unique constraint on team_id alone. The relationship is:
-- - One team can have many jobs (team_id can repeat)
-- - One job belongs to one team (job.id is unique, team_id is just a FK)
-- ============================================

-- ============================================
-- 4. Update existing jobs to be team-specific (optional)
-- If you want to assign existing jobs to teams, uncomment and modify:
-- UPDATE `jobs` SET `team_id` = 1 WHERE `id` IN (1, 2, 3);
-- ============================================

