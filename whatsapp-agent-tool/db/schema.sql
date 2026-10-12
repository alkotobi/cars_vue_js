-- =============================================================================
-- ai — shared business database (pricing catalog + procurement + FAQ).
--
-- SHARED ACROSS MACHINES / TEAMMATES. Everything the whole team must see lives
-- here. Per-machine state (WhatsApp session, question queue, run reports) stays
-- in local JSON files and is deliberately NOT in this schema.
--
-- Scalability decisions:
--   * InnoDB + utf8mb4/utf8mb4_unicode_ci (Arabic/French text, row-level locking,
--     concurrent writers from several machines).
--   * Surrogate BIGINT PKs + a unique business key per table (car_id, supplier
--     name, option name…), so joins never depend on mutable display names.
--   * Child rows normalised (car_colors, car_options, clearance_aliases, the
--     freight port/line lookups, supplier_quotation_item_prices) instead of JSON
--     blobs or free-text strings, so they are queryable and indexable.
--   * Money as DECIMAL (12,2 USD / 15,2 DZD / 12,4 rate) — never FLOAT.
--   * History is APPEND-ONLY (car_price_history, supplier_quotations) and config
--     is a time series (config): newest row by effective_at wins, giving a free
--     audit trail and conflict-free concurrent inserts.
--   * Every lookup column is indexed; FAQ gets a FULLTEXT index for search.
--
-- Procurement is STANDALONE: it does not reference the cars catalog and does not
-- feed car_price_history. A supplier issues timestamped quotations; each
-- quotation has car items; each item carries one price per colour.
--
-- Idempotent: safe to run repeatedly (CREATE TABLE IF NOT EXISTS). This file is
-- the FINAL shape for a fresh database — it is applied with the `mysql` client
-- (see PRICING.md). It does not ALTER an already-populated database; to upgrade
-- a populated database in place, run db/migrate-002-live.sql instead (it is the
-- data-preserving upgrade path, see PRICING.md).
-- =============================================================================

CREATE DATABASE IF NOT EXISTS ai CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ai;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version    VARCHAR(64) NOT NULL PRIMARY KEY,
  applied_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- cars catalog
-- One row per car_id = the CURRENT catalog entry (latest buy price).
CREATE TABLE IF NOT EXISTS cars (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_id       VARCHAR(80)  NOT NULL,
  name         VARCHAR(160) NOT NULL,
  size         ENUM('big','small')  NOT NULL DEFAULT 'small',
  `condition`  ENUM('new','used')   NOT NULL DEFAULT 'new',
  cost_usd     DECIMAL(12,2) NOT NULL,
  supplier_id  BIGINT UNSIGNED NULL,   -- who supplies the current cost_usd (see suppliers)
  loading_port VARCHAR(60)  NULL,
  notes        TEXT NULL,
  price_at     DATETIME(3) NULL,   -- source timestamp of the buy price
  created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_cars_car_id (car_id),
  KEY idx_cars_name (name),
  KEY idx_cars_supplier (supplier_id),
  KEY idx_cars_updated (updated_at),
  CONSTRAINT fk_cars_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS car_colors (
  id      BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_ref BIGINT UNSIGNED NOT NULL,
  color   VARCHAR(40) NOT NULL,
  UNIQUE KEY uq_car_color (car_ref, color),
  CONSTRAINT fk_car_colors_car FOREIGN KEY (car_ref) REFERENCES cars(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Per-car option reference; price_usd NULL => fall back to options.price_usd.
CREATE TABLE IF NOT EXISTS car_options (
  id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_ref     BIGINT UNSIGNED NOT NULL,
  option_name VARCHAR(120) NOT NULL,
  price_usd   DECIMAL(12,2) NULL,
  UNIQUE KEY uq_car_option (car_ref, option_name),
  CONSTRAINT fk_car_options_car FOREIGN KEY (car_ref) REFERENCES cars(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Append-only buy-price audit trail (never edited/deleted). car_id is kept as a
-- text snapshot so history survives a catalog delete.
CREATE TABLE IF NOT EXISTS car_price_history (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_ref       BIGINT UNSIGNED NULL,
  car_id        VARCHAR(80)  NOT NULL,
  name          VARCHAR(160) NOT NULL,
  cost_usd      DECIMAL(12,2) NOT NULL,
  price_at      DATETIME(3) NOT NULL,
  note          VARCHAR(255) NULL,
  supplier_id   BIGINT UNSIGNED NULL,   -- who gave this price
  quotation_id  BIGINT UNSIGNED NULL,   -- the quotation it came from
  created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_cph_car (car_ref, price_at),
  KEY idx_cph_carid (car_id, price_at),
  KEY idx_cph_supplier (supplier_id),
  KEY idx_cph_quotation (quotation_id),
  CONSTRAINT fk_cph_car FOREIGN KEY (car_ref) REFERENCES cars(id) ON DELETE SET NULL,
  CONSTRAINT fk_cph_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL,
  CONSTRAINT fk_cph_quotation FOREIGN KEY (quotation_id) REFERENCES supplier_quotations(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- options
CREATE TABLE IF NOT EXISTS options (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name       VARCHAR(120) NOT NULL,
  price_usd  DECIMAL(12,2) NOT NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_options_name (name)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- config (time series)
-- One generic time series for every tunable number. config_key is one of
-- 'benefits' | 'rate' | 'customs'; value_json holds the per-key shape (see
-- PRICING.md). The newest row by effective_at wins.
CREATE TABLE IF NOT EXISTS config (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  config_key   VARCHAR(40) NOT NULL,
  value_json   JSON NOT NULL,
  effective_at DATETIME(3) NOT NULL,
  created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_config_key_eff (config_key, effective_at)
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- freight
CREATE TABLE IF NOT EXISTS loading_ports (
  id   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(80) NOT NULL,
  UNIQUE KEY uq_loading_ports_name (name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS discharge_ports (
  id   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(80) NOT NULL,
  UNIQUE KEY uq_discharge_ports_name (name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS shipping_lines (
  id   BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name VARCHAR(40) NOT NULL,
  UNIQUE KEY uq_shipping_lines_name (name)
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS freight (
  id                 BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  id_loading_port     BIGINT UNSIGNED NOT NULL,
  id_discharge_port   BIGINT UNSIGNED NOT NULL,
  id_shipping_line    BIGINT UNSIGNED NOT NULL,
  container_40ft_usd DECIMAL(12,2) NOT NULL,
  valid_from         DATE NULL,
  valid_to           DATE NULL,
  etd                DATE NULL,
  free_time_days     INT NULL,
  notes              VARCHAR(500) NULL,
  quoted_at          DATETIME(3) NOT NULL,
  created_at         DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_freight_line (id_shipping_line),
  KEY idx_freight_valid (valid_to),
  KEY idx_freight_quoted (quoted_at),
  CONSTRAINT fk_freight_loading FOREIGN KEY (id_loading_port) REFERENCES loading_ports(id) ON DELETE RESTRICT,
  CONSTRAINT fk_freight_discharge FOREIGN KEY (id_discharge_port) REFERENCES discharge_ports(id) ON DELETE RESTRICT,
  CONSTRAINT fk_freight_line FOREIGN KEY (id_shipping_line) REFERENCES shipping_lines(id) ON DELETE RESTRICT
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- customs clearance
-- car_id is the business key (NULL = unmatched, ask the boss); car_ref links to
-- the catalog when the car exists.
CREATE TABLE IF NOT EXISTS clearance (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  car_ref       BIGINT UNSIGNED NULL,
  car_id        VARCHAR(80)  NULL,
  name          VARCHAR(160) NOT NULL,
  clearance_dza DECIMAL(15,2) NOT NULL,
  note          TEXT NULL,
  effective_at  DATETIME(3) NOT NULL,
  created_at    DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_clearance_car (car_ref),
  KEY idx_clearance_carid (car_id),
  CONSTRAINT fk_clearance_car FOREIGN KEY (car_ref) REFERENCES cars(id) ON DELETE SET NULL
) ENGINE=InnoDB;

CREATE TABLE IF NOT EXISTS clearance_aliases (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  clearance_id BIGINT UNSIGNED NOT NULL,
  alias        VARCHAR(160) NOT NULL,
  UNIQUE KEY uq_clearance_alias (clearance_id, alias),
  KEY idx_alias_lookup (alias),
  CONSTRAINT fk_alias_clearance FOREIGN KEY (clearance_id) REFERENCES clearance(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- procurement (suppliers + quotations)
-- Standalone: no relation to the cars catalog, no car_price_history coupling.
CREATE TABLE IF NOT EXISTS suppliers (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  name         VARCHAR(160) NOT NULL,
  contact_info TEXT NULL,
  notes        TEXT NULL,
  created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_suppliers_name (name)
) ENGINE=InnoDB;

-- Header: one timestamped quotation from a supplier.
CREATE TABLE IF NOT EXISTS supplier_quotations (
  id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  supplier_id BIGINT UNSIGNED NOT NULL,
  quoted_at   DATETIME(3) NOT NULL,   -- when the quotation was received
  reference   VARCHAR(120) NULL,      -- supplier quote no.
  currency    VARCHAR(3) NOT NULL DEFAULT 'USD',
  status      ENUM('requested','received','accepted','rejected','expired') NOT NULL DEFAULT 'received',
  valid_until DATE NULL,
  notes       TEXT NULL,
  created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_supplier_quotations_ref (reference),
  KEY idx_sq_supplier (supplier_id, quoted_at),
  KEY idx_sq_quoted (quoted_at),
  CONSTRAINT fk_sq_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE RESTRICT
) ENGINE=InnoDB;

-- One car line ("a car we need"). car_name is free text (may not be catalogued).
CREATE TABLE IF NOT EXISTS supplier_quotation_items (
  id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  quotation_id BIGINT UNSIGNED NOT NULL,
  car_ref      BIGINT UNSIGNED NULL,    -- linked catalog car when the item matches one
  car_name     VARCHAR(160) NOT NULL,   -- free text (may not be catalogued)
  size         ENUM('big','small') NOT NULL DEFAULT 'small',
  `condition`  ENUM('new','used')  NOT NULL DEFAULT 'new',
  quantity     INT NOT NULL DEFAULT 1,
  `year`       INT NULL,
  `month`      INT NULL,
  notes        TEXT NULL,
  created_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  KEY idx_sqi_quotation (quotation_id),
  KEY idx_sqi_name (car_name),
  KEY idx_sqi_car (car_ref),
  CONSTRAINT fk_sqi_quotation FOREIGN KEY (quotation_id) REFERENCES supplier_quotations(id) ON DELETE CASCADE,
  CONSTRAINT fk_sqi_car FOREIGN KEY (car_ref) REFERENCES cars(id) ON DELETE SET NULL
) ENGINE=InnoDB;

-- One price row per colour of an item — this is where "price differs by colour" lives.
CREATE TABLE IF NOT EXISTS supplier_quotation_item_prices (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  item_id    BIGINT UNSIGNED NOT NULL,
  color      VARCHAR(40) NOT NULL,
  unit_price DECIMAL(12,2) NOT NULL,
  notes      VARCHAR(255) NULL,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_sqip_item_color (item_id, color),
  CONSTRAINT fk_sqip_item FOREIGN KEY (item_id) REFERENCES supplier_quotation_items(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- documents
-- Real supplier documents (proforma invoice, sales contract, …): the paper trail
-- that backs quotations. source_text keeps the full transcription so an AI can
-- re-read the original wording; file_path optionally points at a stored copy.
CREATE TABLE IF NOT EXISTS documents (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  doc_no          VARCHAR(64)  NOT NULL,
  doc_type        ENUM('proforma_invoice','sales_contract','packing_list','bill_of_lading','other') NOT NULL DEFAULT 'other',
  doc_date        DATE NULL,
  supplier_id     BIGINT UNSIGNED NULL,
  counterparty    VARCHAR(255) NULL,        -- buyer / other party named on the document
  currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
  total           DECIMAL(14,2) NULL,
  amount_in_words VARCHAR(255) NULL,
  source_text     LONGTEXT NULL,            -- full transcription (AI-readable)
  file_path       VARCHAR(255) NULL,        -- repo/absolute path to a stored copy
  status          ENUM('received','accepted','paid','cancelled') NOT NULL DEFAULT 'received',
  notes           TEXT NULL,
  created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  KEY idx_documents_no (doc_no),
  KEY idx_documents_supplier (supplier_id),
  KEY idx_documents_date (doc_date),
  KEY idx_documents_type (doc_type),
  CONSTRAINT fk_documents_supplier FOREIGN KEY (supplier_id) REFERENCES suppliers(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- M:N document <-> quotation (a PI can cover several quotations; a quotation can
-- appear in several documents, e.g. PI + sales contract + credit note).
CREATE TABLE IF NOT EXISTS document_quotations (
  document_id  BIGINT UNSIGNED NOT NULL,
  quotation_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (document_id, quotation_id),
  KEY idx_dq_quotation (quotation_id),
  CONSTRAINT fk_dq_document FOREIGN KEY (document_id) REFERENCES documents(id) ON DELETE CASCADE,
  CONSTRAINT fk_dq_quotation FOREIGN KEY (quotation_id) REFERENCES supplier_quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- Supplier payment ledger + SWIFT proof (migration 005).
-- Chain closed: supplier_payments -> suppliers; M:N -> supplier_quotations
-- (payment_quotations, the trades it settles); payment_swifts (1:1) holds the
-- SWIFT MT103: file_url = the PDF link, source_text = the full manuscript (AI-readable).
CREATE TABLE IF NOT EXISTS supplier_payments (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  payment_no      VARCHAR(64)  NOT NULL,        -- our payment reference (unique)
  id_supplier     BIGINT UNSIGNED NOT NULL,
  amount          DECIMAL(14,2) NOT NULL,
  currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
  payment_date    DATE NULL,
  value_date      DATE NULL,                    -- SWIFT value/execution date
  method          ENUM('swift','fps','chats','cash','other') NOT NULL DEFAULT 'swift',
  status          ENUM('planned','sent','cleared','returned','cancelled') NOT NULL DEFAULT 'sent',
  notes           TEXT NULL,
  created_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at      DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_payments_no (payment_no),
  KEY idx_payments_supplier (id_supplier),
  KEY idx_payments_status (status),
  KEY idx_payments_date (payment_date),
  CONSTRAINT fk_payments_supplier FOREIGN KEY (id_supplier) REFERENCES suppliers(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS payment_swifts (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  id_payment          BIGINT UNSIGNED NOT NULL,   -- 1:1 with a supplier payment
  swift_ref           VARCHAR(64)  NULL,          -- e.g. MT103 field 20 : Sender's Reference
  uetr                CHAR(36)     NULL,          -- Unique End-to-end Transaction Reference (field 121)
  sender_bic          VARCHAR(11)  NULL,          -- ordering bank BIC8/BIC11
  receiver_bic        VARCHAR(11)  NULL,          -- beneficiary bank BIC
  beneficiary_name    VARCHAR(255) NULL,
  beneficiary_account VARCHAR(64)  NULL,          -- IBAN / account number
  file_url            VARCHAR(500) NULL,          -- the PDF link you supply
  file_path           VARCHAR(255) NULL,          -- local copy path, if downloaded
  source_text         LONGTEXT NULL,              -- full SWIFT manuscript (AI-readable)
  created_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at          DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_swifts_payment (id_payment),
  UNIQUE KEY uq_swifts_uetr (uetr),
  KEY idx_swifts_ref (swift_ref),
  CONSTRAINT fk_swifts_payment FOREIGN KEY (id_payment) REFERENCES supplier_payments(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- M:N payment <-> quotation (a single payment can settle several invoices).
CREATE TABLE IF NOT EXISTS payment_quotations (
  payment_id   BIGINT UNSIGNED NOT NULL,
  quotation_id BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (payment_id, quotation_id),
  KEY idx_pq_quotation (quotation_id),
  CONSTRAINT fk_pq_payment FOREIGN KEY (payment_id) REFERENCES supplier_payments(id) ON DELETE CASCADE,
  CONSTRAINT fk_pq_quotation FOREIGN KEY (quotation_id) REFERENCES supplier_quotations(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- ---------------------------------------------------------------- faq
CREATE TABLE IF NOT EXISTS faq (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  question   TEXT NOT NULL,
  answer     TEXT NOT NULL,
  context    TEXT NULL,
  active     TINYINT(1) NOT NULL DEFAULT 1,
  added_at   DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  FULLTEXT KEY ft_faq (question, answer)
) ENGINE=InnoDB;

INSERT INTO schema_migrations (version) VALUES ('001_initial')
  ON DUPLICATE KEY UPDATE version = version;
INSERT INTO schema_migrations (version) VALUES ('002_procurement_and_restructure')
  ON DUPLICATE KEY UPDATE version = version;
INSERT INTO schema_migrations (version) VALUES ('003_prices_linked_to_suppliers')
  ON DUPLICATE KEY UPDATE version = version;
INSERT INTO schema_migrations (version) VALUES ('004_documents_linked_to_quotations')
  ON DUPLICATE KEY UPDATE version = version;
INSERT INTO schema_migrations (version) VALUES ('005_supplier_payments_with_swifts')
  ON DUPLICATE KEY UPDATE version = version;
