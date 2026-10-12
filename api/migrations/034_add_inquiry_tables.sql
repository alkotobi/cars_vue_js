-- ============================================
-- Migration: Inquiry tables (supplier quote requests)
-- ============================================
-- Adds the purchase-side "inquiry" system, mirroring buy_bill / buy_details:
-- a header row (one inquiry to a supplier) and its car lines. The app has no
-- inquiry UI yet; the Go tool (whatsapp-tool inquiries ...) drives these tables.
--
-- 1. inquiry        - header, like buy_bill
-- 2. inquiry_details - car lines, like buy_details
--
-- Idempotent: re-running is a no-op (CREATE TABLE IF NOT EXISTS).
--
-- FK choices, matching the buy tables' guards (migrations 028/029):
--   * inquiry_details.id_inquiry -> inquiry.id  ON DELETE CASCADE. A line has no
--     meaning without its header, so deleting an inquiry removes its lines in the
--     same statement the query passthrough runs.
--   * inquiry_details.id_car_name -> cars_names.id ON DELETE RESTRICT. A direct
--     DELETE of a car name that still has inquiry history is refused, exactly as
--     buy_details is guarded. cars_names is created earlier (migration 025/029),
--     so the FK can be declared inline here.

-- ============================================
-- Inquiry header
-- ============================================
CREATE TABLE IF NOT EXISTS `inquiry` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_supplier` int DEFAULT NULL,
  `date_inquiry` datetime DEFAULT NULL,
  `amount` decimal(10,2) DEFAULT NULL,
  `inquiry_ref` varchar(255) DEFAULT NULL,
  `status` varchar(50) DEFAULT 'open',
  `notes` text,
  PRIMARY KEY (`id`),
  UNIQUE KEY `inquiry_ref` (`inquiry_ref`)
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- ============================================
-- Inquiry car lines
-- ============================================
CREATE TABLE IF NOT EXISTS `inquiry_details` (
  `id` int NOT NULL AUTO_INCREMENT,
  `id_inquiry` int NOT NULL,
  `id_car_name` int DEFAULT NULL,
  `id_color` int DEFAULT NULL,
  `QTY` int DEFAULT NULL,
  `year` int DEFAULT NULL,
  `month` int DEFAULT NULL,
  `is_used_car` tinyint(1) DEFAULT '0',
  `is_big_car` tinyint(1) DEFAULT '0',
  `target_price` decimal(10,2) DEFAULT NULL,
  `notes` text,
  PRIMARY KEY (`id`),
  KEY `idx_inquiry_details_id_inquiry` (`id_inquiry`),
  KEY `idx_inquiry_details_id_car_name` (`id_car_name`),
  CONSTRAINT `fk_inquiry_details_inquiry`
    FOREIGN KEY (`id_inquiry`) REFERENCES `inquiry` (`id`)
    ON DELETE CASCADE,
  CONSTRAINT `fk_inquiry_details_car_name`
    FOREIGN KEY (`id_car_name`) REFERENCES `cars_names` (`id`)
    ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=latin1;
