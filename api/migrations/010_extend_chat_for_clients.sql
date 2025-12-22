-- Migration: Extend chat system to support clients
-- Description: Adds id_client field to chat_users and chat_messages tables to allow clients to participate in chat

-- Step 1: Add id_client to chat_users table
ALTER TABLE `chat_users` 
ADD COLUMN `id_client` int DEFAULT NULL AFTER `id_user`;

-- Step 2: Add constraint to ensure either id_user OR id_client is set (but not both or neither)
-- Note: MySQL doesn't support CHECK constraints in older versions, so we'll add a comment
-- Application code should enforce: (id_user IS NOT NULL AND id_client IS NULL) OR (id_user IS NULL AND id_client IS NOT NULL)

-- Step 3: Add index on id_client for performance
ALTER TABLE `chat_users`
ADD INDEX `idx_id_client` (`id_client`);

-- Step 4: Add id_client to chat_messages table for sender identification
ALTER TABLE `chat_messages`
ADD COLUMN `message_from_client_id` int DEFAULT NULL AFTER `message_from_user_id`;

-- Step 5: Add index on message_from_client_id for performance
ALTER TABLE `chat_messages`
ADD INDEX `idx_message_from_client_id` (`message_from_client_id`);

-- Step 6: Update chat_groups to support client owners
ALTER TABLE `chat_groups`
ADD COLUMN `id_client_owner` int DEFAULT NULL AFTER `id_user_owner`;

-- Step 7: Add index on id_client_owner
ALTER TABLE `chat_groups`
ADD INDEX `idx_id_client_owner` (`id_client_owner`);

-- Step 8: Update chat_last_read_message to support clients
ALTER TABLE `chat_last_read_message`
ADD COLUMN `id_client` int DEFAULT NULL AFTER `id_user`;

-- Step 9: Add index on id_client in chat_last_read_message
ALTER TABLE `chat_last_read_message`
ADD INDEX `idx_id_client` (`id_client`);

-- Step 10: Update unique constraint on chat_last_read_message to include id_client
-- Note: We need to handle this carefully. The unique constraint ensures one read status per group per user/client
-- The application should enforce: one row per (id_group, id_user) OR one row per (id_group, id_client)

-- Drop existing unique constraint
-- Note: If this fails, check the actual index name with: SHOW INDEX FROM chat_last_read_message;
-- and drop it manually, then continue with the rest of the migration
ALTER TABLE `chat_last_read_message`
DROP INDEX `id_group`;

-- Add index on (id_group, id_user) for users (allows NULL for id_client)
ALTER TABLE `chat_last_read_message`
ADD INDEX `idx_group_user` (`id_group`, `id_user`);

-- Add index on (id_group, id_client) for clients (allows NULL for id_user)
ALTER TABLE `chat_last_read_message`
ADD INDEX `idx_group_client` (`id_group`, `id_client`);

-- Step 11: Update chat_read_by to support clients
ALTER TABLE `chat_read_by`
ADD COLUMN `id_client` int DEFAULT NULL AFTER `id_user`;

-- Step 12: Add index on id_client in chat_read_by
ALTER TABLE `chat_read_by`
ADD INDEX `idx_id_client` (`id_client`);

