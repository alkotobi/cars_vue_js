-- ============================================
-- Migration: supplier_credibility_checks
-- ============================================
-- Every credibility run is kept, not just the latest: a score that changes
-- between two checks is itself information, and overwriting the row would make
-- it impossible to see that it moved or why.
--
-- red_flags is a JSON array of strings, validated by the database as well as by
-- lib/ai_client.php, so a bad write cannot leave unparseable JSON behind.
--
-- Checks cascade away with the supplier. That is deliberate: a check is a
-- reading of one supplier's fields, and a deleted supplier has no fields left to
-- be re-read.
--
-- Idempotent: re-running is a no-op.

CREATE TABLE IF NOT EXISTS `supplier_credibility_checks` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_supplier` int NOT NULL COMMENT 'FK to suppliers',
  `score` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '0-100, higher means lower risk',
  `risk_level` enum('low','medium','high') NOT NULL DEFAULT 'medium',
  `summary` text NOT NULL,
  `red_flags` longtext NOT NULL COMMENT 'JSON array of short strings',
  `confidence` enum('low','medium','high') NOT NULL DEFAULT 'low' COMMENT 'How much the input limited the conclusion',
  `model` varchar(100) NOT NULL DEFAULT '' COMMENT 'Model that produced the answer',
  `lang` varchar(8) NOT NULL DEFAULT 'en',
  `id_user` int DEFAULT NULL COMMENT 'Admin who ran the check',
  `date_create` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_supplier_date` (`id_supplier`, `date_create`),
  KEY `idx_id_user` (`id_user`),
  CONSTRAINT `chk_credibility_red_flags_json` CHECK (JSON_VALID(`red_flags`)),
  CONSTRAINT `chk_credibility_score_range` CHECK (`score` BETWEEN 0 AND 100),
  CONSTRAINT `fk_credibility_supplier` FOREIGN KEY (`id_supplier`) REFERENCES `suppliers` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
