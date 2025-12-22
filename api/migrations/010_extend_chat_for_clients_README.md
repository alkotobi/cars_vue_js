# Migration 010: Extend Chat System for Clients

## Overview
This migration extends the chat system to support clients (not just users) in chat groups and messages.

## Changes Made

### 1. `chat_users` table
- Added `id_client` column (nullable)
- Added index on `id_client`
- **Constraint**: Either `id_user` OR `id_client` should be set, but not both (enforced in application code)

### 2. `chat_messages` table
- Added `message_from_client_id` column (nullable)
- Added index on `message_from_client_id`
- Messages can now be sent by either users or clients

### 3. `chat_groups` table
- Added `id_client_owner` column (nullable)
- Added index on `id_client_owner`
- Groups can now be owned by either users or clients

### 4. `chat_last_read_message` table
- Added `id_client` column (nullable)
- Added indexes for both user and client lookups
- **Note**: The unique constraint `id_group` is dropped. Application code should enforce uniqueness.

### 5. `chat_read_by` table
- Added `id_client` column (nullable)
- Added index on `id_client`

## Important Notes

1. **Data Integrity**: The application code must enforce that:
   - In `chat_users`: Either `id_user` IS NOT NULL OR `id_client` IS NOT NULL (but not both)
   - In `chat_messages`: Either `message_from_user_id` IS NOT NULL OR `message_from_client_id` IS NOT NULL (but not both)

2. **Backward Compatibility**: Existing data with `id_user` will continue to work. The new `id_client` fields are nullable and optional.

3. **Index Dropping**: If the migration fails at Step 10 (dropping the `id_group` index), check the actual index name with:
   ```sql
   SHOW INDEX FROM chat_last_read_message;
   ```
   Then manually drop it and continue with the rest of the migration.

## Testing
After running this migration:
1. Verify all tables have the new columns
2. Test that existing user-based chats still work
3. Test adding clients to chat groups
4. Test sending messages as a client

