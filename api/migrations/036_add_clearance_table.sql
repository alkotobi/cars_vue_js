-- ============================================
-- Migration: per-car customs clearance (the cars app)
-- ============================================
-- A customs clearance amount (in DZD, **TVN excluded**) for a car model, so the
-- clearance quoted for a model lives beside the model instead of in a spreadsheet.
-- These are the small, not-used cars; the row points at cars_names (the car is
-- chosen from the catalog, never free text), and is_big_car / is_used_car keep the
-- big / used variants apart once they are added.
--
-- These values are fixed (customs clearance does not change), so they are reference
-- data seeded here and the table is registered in TENANT_SEED_TABLES, not
-- TENANT_BUSINESS_TABLES.
--
-- Idempotent: CREATE TABLE IF NOT EXISTS + INSERT ... NOT EXISTS guards. The brand
-- is looked up by name, so a tenant without a brand simply skips that car.

CREATE TABLE IF NOT EXISTS `clearance` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_car_name` int NOT NULL,
  `clearance_dza` decimal(15,2) NOT NULL COMMENT 'Customs clearance in DZD, TVN excluded',
  `is_big_car` tinyint(1) NOT NULL DEFAULT 0,
  `is_used_car` tinyint(1) NOT NULL DEFAULT 0,
  `notes` text,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_clearance_car` (`id_car_name`, `is_big_car`, `is_used_car`),
  KEY `idx_clearance_car` (`id_car_name`),
  CONSTRAINT `fk_clearance_car`
    FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Car models (small, not used). Brand is looked up by name so the seed is portable;
-- a tenant without the brand simply skips that car. The unique key on car_name and
-- the NOT EXISTS guard make re-running a no-op.
INSERT INTO cars_names (car_name, id_brand, is_big_car)
SELECT v.car_name, b.id, 0
FROM (
  SELECT 'COOLRAY BATTLE'    AS car_name, 'GEELY'   AS brand UNION ALL
  SELECT 'COOLRAY L',          'GEELY'   UNION ALL
  SELECT 'COOLRAY MANUAL',     'GEELY'   UNION ALL
  SELECT 'MG5 MANUAL',         'MG'      UNION ALL
  SELECT 'MG5 AUTO',           'MG'      UNION ALL
  SELECT 'LIVAN MANUAL',       'LIVAN'   UNION ALL
  SELECT 'LIVAN AUTO',         'LIVAN'   UNION ALL
  SELECT 'CHANGAN X5 MAN',     'CHANGAN' UNION ALL
  SELECT 'CHANGAN X5 BEYOND',  'CHANGAN' UNION ALL
  SELECT 'JETTA VS5 FLAGSHIP', 'JETTA'   UNION ALL
  SELECT 'GS3 MID',            'GAC'     UNION ALL
  SELECT 'GS3 FULL',           'GAC'     UNION ALL
  SELECT 'GS3 EMZOOM',         'GAC'
) v
JOIN brands b ON b.brand = v.brand
LEFT JOIN cars_names c ON c.car_name = v.car_name
WHERE c.id IS NULL;

-- Fixed customs clearance (DZD, TVN excluded) for the cars above.
INSERT INTO clearance (id_car_name, clearance_dza, is_big_car, is_used_car)
SELECT c.id, v.price, 0, 0
FROM (
  SELECT 'COOLRAY BATTLE'    AS car_name, 950000 AS price UNION ALL
  SELECT 'COOLRAY L',         950000 UNION ALL
  SELECT 'COOLRAY MANUAL',    850000 UNION ALL
  SELECT 'MG5 MANUAL',        800000 UNION ALL
  SELECT 'MG5 AUTO',          800000 UNION ALL
  SELECT 'LIVAN MANUAL',      800000 UNION ALL
  SELECT 'LIVAN AUTO',        800000 UNION ALL
  SELECT 'CHANGAN X5 MAN',    950000 UNION ALL
  SELECT 'CHANGAN X5 BEYOND', 950000 UNION ALL
  SELECT 'JETTA VS5 FLAGSHIP',800000 UNION ALL
  SELECT 'GS3 MID',           950000 UNION ALL
  SELECT 'GS3 FULL',          950000 UNION ALL
  SELECT 'GS3 EMZOOM',        950000
) v
JOIN cars_names c ON c.car_name = v.car_name
LEFT JOIN clearance cl
  ON cl.id_car_name = c.id AND cl.is_big_car = 0 AND cl.is_used_car = 0
WHERE cl.id IS NULL;
