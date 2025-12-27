-- ============================================
-- Migration: Add photos and videos support to car names
-- ============================================
-- This migration creates a table to store photos and videos for car names

-- Create car_name_media table
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

