SET FOREIGN_KEY_CHECKS = 0;
START TRANSACTION;

SET time_zone = "+00:00";
SET NAMES utf8mb4;

-- 1) Adjust existing tables to match v25

-- 1.1 banks: add new columns used in v25
ALTER TABLE `banks`
  ADD COLUMN `mobile` varchar(255) DEFAULT NULL AFTER `company_address`,
  ADD COLUMN `email` varchar(255) DEFAULT NULL AFTER `mobile`,
  ADD COLUMN `website` varchar(500) DEFAULT NULL AFTER `email`,
  ADD COLUMN `logo_path` varchar(500) DEFAULT NULL COMMENT 'Path to bank/company logo image file' AFTER `website`,
  ADD COLUMN `path_letter_head` varchar(500) DEFAULT NULL COMMENT 'Path to letterhead image file' AFTER `logo_path`,
  ADD COLUMN `path_stamp` varchar(500) DEFAULT NULL COMMENT 'Path to stamp image file' AFTER `path_letter_head`,
  ADD COLUMN `is_active` tinyint(1) DEFAULT 1 COMMENT '1=active, 0=inactive' AFTER `path_stamp`;

-- 1.2 brands: add logo_path
ALTER TABLE `brands`
  ADD COLUMN `logo_path` varchar(500) DEFAULT NULL COMMENT 'Path to brand logo image file' AFTER `brand`;

-- 1.3 buy_bill: align defaults
ALTER TABLE `buy_bill`
  MODIFY `is_stock_updated` tinyint(1) DEFAULT 0,
  MODIFY `is_ordered` tinyint(1) DEFAULT 1;

-- 1.4 cars_stock: add payment_confirmed + vin unique index
ALTER TABLE `cars_stock`
  ADD COLUMN `payment_confirmed` tinyint(1) DEFAULT 0 COMMENT 'Payment confirmed status - only admins or users with can_confirm_payment permission can set to true' AFTER `id_color`,
  ADD UNIQUE KEY `vin` (`vin`),
  ADD KEY `idx_payment_confirmed` (`payment_confirmed`);

-- 1.5 chat_groups: add client owner + index
ALTER TABLE `chat_groups`
  ADD COLUMN `id_client_owner` int(11) DEFAULT NULL AFTER `id_user_owner`,
  ADD KEY `idx_id_client_owner` (`id_client_owner`);

-- 1.6 chat_last_read_message: add client support and adjust indexes
ALTER TABLE `chat_last_read_message`
  ADD COLUMN `id_client` int(11) DEFAULT NULL AFTER `id_user`;

DROP INDEX `id_group` ON `chat_last_read_message`;

ALTER TABLE `chat_last_read_message`
  ADD KEY `idx_group_user` (`id_group`,`id_user`),
  ADD KEY `idx_group_client` (`id_group`,`id_client`);

-- 1.7 chat_messages: add client sender + index
ALTER TABLE `chat_messages`
  ADD COLUMN `message_from_client_id` int(11) DEFAULT NULL AFTER `message_from_user_id`,
  ADD KEY `idx_message_from_client_id` (`message_from_client_id`);

-- 1.8 chat_read_by: add client column + index
ALTER TABLE `chat_read_by`
  ADD COLUMN `id_client` int(11) DEFAULT NULL AFTER `id_user`,
  ADD KEY `idx_id_client` (`id_client`);

-- 1.9 chat_users: add client column + index
ALTER TABLE `chat_users`
  ADD COLUMN `id_client` int(11) DEFAULT NULL AFTER `id_user`,
  ADD KEY `idx_id_client` (`id_client`);

-- 1.10 clients: add share_token and adjust constraints
ALTER TABLE `clients`
  ADD COLUMN `share_token` varchar(64) DEFAULT NULL AFTER `id`,
  MODIFY `nin` varchar(40) DEFAULT NULL;

ALTER TABLE `clients`
  DROP INDEX `client_name_unic`,
  DROP INDEX `nin`,
  ADD UNIQUE KEY `client_name_unic` (`name`,`id_no`),
  ADD UNIQUE KEY `idx_share_token` (`share_token`);

-- 1.11 containers: add unique on name
ALTER TABLE `containers`
  ADD UNIQUE KEY `nm` (`name`);

-- 1.12 priorities: add unique on priority
ALTER TABLE `priorities`
  ADD UNIQUE KEY `priority` (`priority`);

-- 1.13 sell_bill: extend to v25 structure
-- NOTE: this assumes existing sell_bill.notes values are valid JSON or NULL.
ALTER TABLE `sell_bill`
  MODIFY `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  MODIFY `notes` JSON DEFAULT NULL COMMENT 'JSON array of notes: [{"id_user": int, "note": "text", "timestamp": "datetime"}]';

ALTER TABLE `sell_bill`
  ADD COLUMN `payment_confirmed` tinyint(1) DEFAULT 0 COMMENT 'Payment confirmed status - only admins or users with can_confirm_payment permission can set to true' AFTER `is_batch_sell`,
  ADD COLUMN `payment_confirmed_by_user_id` int(11) DEFAULT NULL COMMENT 'ID of user who confirmed the payment' AFTER `payment_confirmed`,
  ADD COLUMN `edited_by` JSON DEFAULT NULL COMMENT 'JSON array tracking edits: [{"id_user": int, "timestamp": "datetime"}]' AFTER `payment_confirmed_by_user_id`,
  ADD KEY `idx_payment_confirmed` (`payment_confirmed`),
  ADD KEY `idx_payment_confirmed_by_user_id` (`payment_confirmed_by_user_id`);

ALTER TABLE `sell_bill`
  ADD CONSTRAINT `fk_sell_bill_payment_confirmed_by_user`
    FOREIGN KEY (`payment_confirmed_by_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL;

-- 1.14 users: make role_id NOT NULL, align id_bank_account
ALTER TABLE `users`
  MODIFY `role_id` int(11) NOT NULL,
  MODIFY `id_bank_account` int(10) UNSIGNED DEFAULT NULL;

-- 2) New tables from v25 that did not exist in v19

-- 2.1 Multi-DB management tables
CREATE TABLE IF NOT EXISTS `login` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user` varchar(255) DEFAULT NULL,
  `pass` text,
  `active` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `dbs` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `db_code` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `db_host_start` date DEFAULT NULL,
  `db_host_end` date DEFAULT NULL,
  `serv_host_start` date DEFAULT NULL,
  `serv_host_end` date DEFAULT NULL,
  `db_host_cost_per_month` double DEFAULT NULL,
  `serv_host_cost_per_month` double DEFAULT NULL,
  `files_dir` varchar(255) DEFAULT NULL,
  `js_dir` varchar(255) DEFAULT NULL,
  `db_name` varchar(255) DEFAULT NULL,
  `is_created` int DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`db_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `db_updates` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `sql` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `from_version` int NOT NULL,
  `current_version` int NOT NULL,
  `description` text,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 2.2 Car media table
CREATE TABLE IF NOT EXISTS `car_name_media` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_name_id` int(11) NOT NULL COMMENT 'FK to cars_names',
  `file_path` varchar(500) NOT NULL COMMENT 'Storage path to media file',
  `file_name` varchar(255) NOT NULL COMMENT 'Original filename',
  `file_size` bigint(20) DEFAULT NULL COMMENT 'File size in bytes',
  `file_type` varchar(100) DEFAULT NULL COMMENT 'MIME type (image/* or video/*)',
  `media_type` enum('photo', 'video') NOT NULL COMMENT 'Type: photo or video',
  `uploaded_by` int(11) NOT NULL COMMENT 'FK to users - who uploaded',
  `uploaded_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '0=Deleted, 1=Active',
  PRIMARY KEY (`id`),
  KEY `idx_car_name_id` (`car_name_id`),
  KEY `idx_uploaded_by` (`uploaded_by`),
  KEY `idx_uploaded_at` (`uploaded_at`),
  KEY `idx_media_type` (`media_type`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_car_name_media_car_name` FOREIGN KEY (`car_name_id`) REFERENCES `cars_names` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_name_media_uploaded_by` FOREIGN KEY (`uploaded_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- 2.3 Upgrades & car_apgrades
CREATE TABLE IF NOT EXISTS `upgrades` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `description` varchar(255) DEFAULT NULL,
  `notes` text,
  `id_user_owner` int DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id_user_owner` (`id_user_owner`),
  CONSTRAINT `fk_upgrades_user_owner` FOREIGN KEY (`id_user_owner`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE IF NOT EXISTS `car_apgrades` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `id_car` int DEFAULT NULL,
  `id_upgrade` int DEFAULT NULL,
  `value` float NOT NULL,
  `date_done` datetime DEFAULT NULL,
  `id_uder_done` int DEFAULT NULL,
  `id_user_create` int DEFAULT NULL,
  `time_creation` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_id_car` (`id_car`),
  KEY `idx_id_upgrade` (`id_upgrade`),
  KEY `idx_id_uder_done` (`id_uder_done`),
  KEY `idx_id_user_create` (`id_user_create`),
  CONSTRAINT `fk_car_apgrades_car` FOREIGN KEY (`id_car`) REFERENCES `cars_stock` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_apgrades_upgrade` FOREIGN KEY (`id_upgrade`) REFERENCES `upgrades` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_apgrades_user` FOREIGN KEY (`id_uder_done`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_car_apgrades_user_create` FOREIGN KEY (`id_user_create`) REFERENCES `users` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 2.4 Car file management tables
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
  KEY `idx_car_category` (`car_id`,`category_id`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_car_files_car` FOREIGN KEY (`car_id`) REFERENCES `cars_stock` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_car_files_category` FOREIGN KEY (`category_id`) REFERENCES `car_file_categories` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_car_files_uploaded_by` FOREIGN KEY (`uploaded_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `custom_clearance_agents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT 'Agent name or company name',
  `contact_person` varchar(255) DEFAULT NULL COMMENT 'Contact person name',
  `phone` varchar(50) DEFAULT NULL COMMENT 'Phone number',
  `email` varchar(255) DEFAULT NULL COMMENT 'Email address',
  `address` text DEFAULT NULL COMMENT 'Physical address',
  `license_number` varchar(100) DEFAULT NULL COMMENT 'License or registration number',
  `notes` text DEFAULT NULL COMMENT 'Additional notes',
  `is_active` tinyint(1) NOT NULL DEFAULT 1 COMMENT '1=Active, 0=Inactive',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `car_file_physical_tracking` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `current_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who has physical copy now, NULL = available',
  `previous_holder_id` int(11) DEFAULT NULL COMMENT 'FK to users - who had it before',
  `custom_clearance_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if checked out to agent',
  `checkout_type` enum('user', 'client', 'custom_clearance_agent') NOT NULL DEFAULT 'user' COMMENT 'Type of checkout: user, client, or custom clearance agent',
  `client_id` int(11) DEFAULT NULL COMMENT 'FK to clients - if checked out to client',
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
  KEY `idx_car_file_status` (`car_file_id`,`status`),
  KEY `idx_agent_id` (`custom_clearance_agent_id`),
  KEY `idx_client_id` (`client_id`),
  KEY `idx_checkout_type` (`checkout_type`),
  CONSTRAINT `fk_physical_tracking_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_physical_tracking_current_holder` FOREIGN KEY (`current_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_previous_holder` FOREIGN KEY (`previous_holder_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_physical_tracking_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_tracking_agent` FOREIGN KEY (`custom_clearance_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_tracking_client` FOREIGN KEY (`client_id`) REFERENCES `clients` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

CREATE TABLE IF NOT EXISTS `car_file_transfers` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `car_file_id` int(11) NOT NULL COMMENT 'FK to car_files',
  `from_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who transferred from, NULL = available',
  `to_user_id` int(11) DEFAULT NULL COMMENT 'FK to users - who received (NULL if transferred to agent or client)',
  `from_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if transferred from agent',
  `to_agent_id` int(11) DEFAULT NULL COMMENT 'FK to custom_clearance_agents - if transferred to agent',
  `from_client_name` varchar(255) DEFAULT NULL COMMENT 'Client name if transferred from client',
  `to_client_name` varchar(255) DEFAULT NULL COMMENT 'Client name if transferred to client',
  `transfer_type` enum('user_to_user', 'user_to_agent', 'agent_to_user', 'user_to_client', 'client_to_user') NOT NULL DEFAULT 'user_to_user',
  `transfer_status` enum('pending', 'approved', 'rejected') NOT NULL DEFAULT 'approved' COMMENT 'Transfer status: pending (waiting approval), approved (completed), rejected (cancelled)',
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
  KEY `idx_from_agent` (`from_agent_id`),
  KEY `idx_to_agent` (`to_agent_id`),
  KEY `idx_transfer_type` (`transfer_type`),
  KEY `idx_transfer_status` (`transfer_status`),
  CONSTRAINT `fk_transfers_file` FOREIGN KEY (`car_file_id`) REFERENCES `car_files` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_transfers_from_user` FOREIGN KEY (`from_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_to_user` FOREIGN KEY (`to_user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_transferred_by` FOREIGN KEY (`transferred_by`) REFERENCES `users` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_transfers_from_agent` FOREIGN KEY (`from_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_transfers_to_agent` FOREIGN KEY (`to_agent_id`) REFERENCES `custom_clearance_agents` (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- 2.5 Teams, team_members, jobs, car_selections & related tables
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

-- 3) Triggers (no data changes)

DELIMITER $$

DROP TRIGGER IF EXISTS `trg_increment_team_jobs_completed`$$

CREATE TRIGGER `trg_increment_team_jobs_completed`
AFTER UPDATE ON `car_selections`
FOR EACH ROW
BEGIN
  IF NEW.status = 'completed'
     AND OLD.status <> 'completed'
     AND NEW.job_done_on IS NOT NULL
     AND NEW.assigned_to_team IS NOT NULL THEN
    UPDATE `teams`
    SET `jobs_completed_count` = `jobs_completed_count` + 1
    WHERE `id` = NEW.assigned_to_team;
  END IF;
END$$

DROP TRIGGER IF EXISTS `generate_client_share_token`$$

CREATE TRIGGER `generate_client_share_token`
BEFORE INSERT ON `clients`
FOR EACH ROW
BEGIN
  IF NEW.share_token IS NULL OR NEW.share_token = '' THEN
    SET NEW.share_token = SUBSTRING(
      SHA2(CONCAT(NEW.id, NEW.name, COALESCE(NEW.id_no, ''), NOW(), RAND()), 256),
      1,
      64
    );
  END IF;
END$$

DELIMITER ;

COMMIT;
SET FOREIGN_KEY_CHECKS = 1;

