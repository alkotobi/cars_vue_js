-- 009 — redundant chat-poll daemon: shared cursors + a single-writer lease (ai DB).
-- Lets several machines run `whatsapp-tool chat-poll` against the SAME tenant with
-- only one active at a time (the lease holder); a standby takes over once the
-- holder's lease expires (its heartbeat stopped). The cursor moves out of the
-- per-machine data/chatbot-state.json into chatbot_group_cursors so a machine
-- that takes over resumes at the right message instead of re-answering history.
-- Idempotent: guarded DDL + ODKU migration row.

-- Per-group last-answered client message id (one row per tenant+group). Writes
-- are monotonic (GREATEST) so a late/stale writer can never move it backwards.
CREATE TABLE IF NOT EXISTS chatbot_group_cursors (
  id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  tenant          VARCHAR(64)     NOT NULL,   -- cars-app tenant (e.g. mig_26)
  id_chat_group   BIGINT UNSIGNED NOT NULL,   -- chat_groups.id in that tenant
  last_message_id BIGINT UNSIGNED NOT NULL DEFAULT 0,   -- highest answered client message id
  updated_at      DATETIME(3)     NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  UNIQUE KEY uq_chatbot_cursor (tenant, id_chat_group)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Single-writer lease: one row per tenant. The active daemon holds it (owner =
-- host:pid:profile) and renews expires_at every heartbeat; a standby may claim it
-- only once expires_at has passed (the holder died / stopped renewing). All time
-- comparisons use the DB clock (NOW(3)) so machine clock skew cannot split the lock.
CREATE TABLE IF NOT EXISTS chatbot_leases (
  tenant       VARCHAR(64)  NOT NULL PRIMARY KEY,
  owner        VARCHAR(190) NOT NULL,             -- host:pid:profile of the current holder
  acquired_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  heartbeat_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  expires_at   DATETIME(3)  NOT NULL,             -- heartbeat_at + TTL; claimable once in the past
  KEY idx_chatbot_lease_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- migration row
INSERT INTO schema_migrations (version) VALUES ('009_chatbot_lease') ON DUPLICATE KEY UPDATE version = version;
