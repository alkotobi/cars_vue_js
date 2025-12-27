-- ============================================
-- Migration: Add logo support to brands table
-- ============================================
-- This migration adds a logo_path column to store brand logo file paths

-- Add logo_path column to brands table
ALTER TABLE `brands`
ADD COLUMN `logo_path` varchar(500) DEFAULT NULL COMMENT 'Path to brand logo image file' AFTER `brand`;

