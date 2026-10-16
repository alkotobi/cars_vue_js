-- Migration: chat unread-badge correctness + indexes
--
-- Why this exists
-- ---------------
-- Two production symptoms traced back to this schema:
--
--   1. The unread badge sticks at 0 (or the wrong number). The client writes
--      chat_last_read_message with a "UPDATE, and INSERT when affectedRows = 0"
--      pattern. MySQL's rowCount() reports rows *changed*, not matched, so
--      re-saving the same last-read id returned 0 and the client inserted a
--      second row for the same (id_group, id_user). With duplicates present the
--      correlated subquery in the badge query (one row expected) raised
--      "Subquery returns more than 1 row", the whole count query failed, and
--      every badge was reset to 0.
--
--   2. Every chat query filtered on id_chat_group / id_user with no index, so
--      each poll full-scanned the messages and membership tables. Repeated
--      every few seconds per browser tab, it grows with the message history.
--
-- This migration:
--   * collapses duplicate last-read rows (keeping the furthest read position),
--   * makes (id_group, id_user) and (id_group, id_client) unique so a row can
--     never be duplicated again,
--   * adds the indexes the chat queries need.
--
-- Idempotency: setup.sql is kept at the final shape, so replaying this file on a
-- fresh tenant is a no-op. Adding an index that already exists raises 1061,
-- which api/lib/tenant-provision.php treats as "already applied".
--
-- NOTE: MySQL has no "DROP INDEX IF EXISTS", so the two unique keys are created
-- by dropping and re-adding the existing key names in a single ALTER. On a fresh
-- database the keys already exist under those names, so the drop succeeds.

-- ---------------------------------------------------------------------------
-- 1. Collapse duplicate last-read rows.
--
-- First lift every duplicate to the furthest read position so keeping one of
-- them cannot lose progress, then drop the extras and keep the highest id.
-- ---------------------------------------------------------------------------

UPDATE chat_last_read_message lr
JOIN (
    SELECT id_group, id_user, MAX(id_last_read_message) AS max_read
    FROM chat_last_read_message
    WHERE id_user IS NOT NULL AND id_client IS NULL
    GROUP BY id_group, id_user
) k ON k.id_group = lr.id_group AND k.id_user = lr.id_user
SET lr.id_last_read_message = k.max_read
WHERE lr.id_client IS NULL AND lr.id_user IS NOT NULL;

DELETE FROM chat_last_read_message
WHERE id_user IS NOT NULL
  AND id_client IS NULL
  AND id NOT IN (
      SELECT keep_id FROM (
          SELECT MAX(id) AS keep_id
          FROM chat_last_read_message
          WHERE id_user IS NOT NULL AND id_client IS NULL
          GROUP BY id_group, id_user
      ) AS t
  );

UPDATE chat_last_read_message lr
JOIN (
    SELECT id_group, id_client, MAX(id_last_read_message) AS max_read
    FROM chat_last_read_message
    WHERE id_client IS NOT NULL AND id_user IS NULL
    GROUP BY id_group, id_client
) k ON k.id_group = lr.id_group AND k.id_client = lr.id_client
SET lr.id_last_read_message = k.max_read
WHERE lr.id_user IS NULL AND lr.id_client IS NOT NULL;

DELETE FROM chat_last_read_message
WHERE id_client IS NOT NULL
  AND id_user IS NULL
  AND id NOT IN (
      SELECT keep_id FROM (
          SELECT MAX(id) AS keep_id
          FROM chat_last_read_message
          WHERE id_client IS NOT NULL AND id_user IS NULL
          GROUP BY id_group, id_client
      ) AS t
  );

-- ---------------------------------------------------------------------------
-- 2. Make one last-read row per (group, user) and per (group, client).
--
-- MySQL treats NULLs as distinct in a UNIQUE index, so each key only enforces
-- the rows that belong to it: a user row has id_client NULL and is governed by
-- uniq_group_user; a client row has id_user NULL and is governed by
-- uniq_group_client. The names match the keys setup.sql declares, so replaying
-- this on a fresh tenant drops and re-adds them rather than colliding.
-- ---------------------------------------------------------------------------

ALTER TABLE chat_last_read_message
    DROP INDEX idx_group_user,
    ADD UNIQUE INDEX idx_group_user (id_group, id_user),
    DROP INDEX idx_group_client,
    ADD UNIQUE INDEX idx_group_client (id_group, id_client);

-- ---------------------------------------------------------------------------
-- 3. Indexes the chat queries filter and order on.
--
-- Leading column is the equality predicate in every case; the trailing column
-- lets the ORDER BY / MAX(id) walk the index instead of sorting.
-- ---------------------------------------------------------------------------

CREATE INDEX idx_chat_group_id ON chat_messages (id_chat_group, id);

CREATE INDEX idx_chat_users_group_user ON chat_users (id_chat_group, id_user);

CREATE INDEX idx_chat_users_user_active ON chat_users (id_user, is_active);

CREATE INDEX idx_chat_users_group_client ON chat_users (id_chat_group, id_client);
