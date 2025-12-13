-- Database Setup SQL File
-- This file contains all CREATE TABLE statements for setting up the databases
-- Execute this file to create all necessary tables

-- ============================================
-- Database: merhab_databases
-- ============================================

-- Login table
CREATE TABLE IF NOT EXISTS `login` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `user` varchar(255) DEFAULT NULL,
  `pass` text,
  `active` int DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Databases table
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
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`db_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ============================================
-- Add more CREATE TABLE statements below
-- ============================================


