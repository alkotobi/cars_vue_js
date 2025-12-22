import { ref } from 'vue'
import { useApi } from './useApi'

/**
 * Composable to create/ensure a chat group for a client's car
 * The group includes all users who have this car in their selections
 */
export function useCarClientChat() {
  const { callApi } = useApi()
  const loading = ref(false)
  const error = ref(null)

  /**
   * Create or get a chat group for a client's car
   * @param {number} clientId - The client ID
   * @param {string} clientName - The client name
   * @param {number} carId - The car ID
   * @param {string} carName - The car name (e.g., "Toyota Camry" or "Brand Model")
   * @returns {Promise<number|null>} The chat group ID or null if failed
   */
  const createOrGetCarClientChatGroup = async (clientId, clientName, carId, carName) => {
    loading.value = true
    error.value = null

    try {
      // Generate group name: "{client name} - {car name} - {car id}"
      const groupName = `${clientName} - ${carName} - ${carId}`
      const groupDescription = `Chat group for client ${clientName} regarding car ${carName} (ID: ${carId})`

      // 1. Check if group already exists
      const existingGroupResult = await callApi({
        query: `SELECT id FROM chat_groups WHERE name = ? AND is_active = 1`,
        params: [groupName],
        requiresAuth: false,
      })

      let groupId

      if (existingGroupResult.success && existingGroupResult.data && existingGroupResult.data.length > 0) {
        // Group exists, get its ID
        groupId = existingGroupResult.data[0].id
        console.log('Existing chat group found:', groupId)
      } else {
        // 2. Create new group (owned by client)
        const createGroupResult = await callApi({
          query: `INSERT INTO chat_groups (name, description, id_client_owner, is_active) VALUES (?, ?, ?, 1)`,
          params: [groupName, groupDescription, clientId],
          requiresAuth: false,
        })

        if (!createGroupResult.success || !createGroupResult.lastInsertId) {
          throw new Error('Failed to create chat group.')
        }
        groupId = createGroupResult.lastInsertId
        console.log('New chat group created:', groupId)

        // Add client to the group
        await addClientToGroup(groupId, clientId)
      }

      // 3. Find all users who have this car in their selections
      const usersWithCarInSelections = await findUsersWithCarInSelections(carId)

      let usersAdded = 0
      if (usersWithCarInSelections.length > 0) {
        // 4. Ensure all these users are in the group (returns count of newly added users)
        usersAdded = await ensureUsersInGroup(groupId, usersWithCarInSelections)
      }

      // 5. If no users found, add all admins and send automatic message
      if (usersAdded === 0 && usersWithCarInSelections.length === 0) {
        const adminIds = await findAllAdmins()
        if (adminIds.length > 0) {
          await ensureUsersInGroup(groupId, adminIds)
          // Send automatic message from client
          await sendAutomaticMessage(groupId, clientId, carName)
        }
      }

      // 6. Ensure client is in the group (in case group existed but client wasn't added)
      await ensureClientInGroup(groupId, clientId)

      return { groupId, usersFound: usersAdded }
    } catch (err) {
      error.value = err.message
      console.error('Error creating/getting car client chat group:', err)
      return { groupId: null, usersFound: 0 }
    } finally {
      loading.value = false
    }
  }

  /**
   * Find all user IDs who have this car in their selections
   * @param {number} carId - The car ID to search for
   * @returns {Promise<number[]>} Array of user IDs
   */
  const findUsersWithCarInSelections = async (carId) => {
    try {
      // Get all selections that contain this car ID in their selection_data JSON
      // First, get all selections with the car
      const result = await callApi({
        query: `
          SELECT 
            cs.id,
            cs.owned_by,
            cs.selection_data
          FROM car_selections cs
          WHERE JSON_CONTAINS(cs.selection_data, CAST(? AS JSON))
            AND cs.owned_by IS NOT NULL
        `,
        params: [JSON.stringify(carId)],
        requiresAuth: false,
      })

      if (!result.success || !result.data) {
        return []
      }

      // Extract user IDs from owned_by JSON arrays
      const userIds = []
      result.data.forEach((row) => {
        try {
          // Parse owned_by JSON
          let ownedBy = row.owned_by
          if (typeof ownedBy === 'string') {
            ownedBy = JSON.parse(ownedBy)
          }

          if (Array.isArray(ownedBy)) {
            ownedBy.forEach((userId) => {
              const id = parseInt(userId)
              if (!isNaN(id) && id > 0 && !userIds.includes(id)) {
                userIds.push(id)
              }
            })
          }
        } catch (e) {
          console.warn('Error parsing owned_by for selection', row.id, ':', e)
        }
      })

      console.log(`Found ${userIds.length} users with car ${carId} in selections:`, userIds)
      return userIds
    } catch (err) {
      console.error('Error finding users with car in selections:', err)
      return []
    }
  }

  /**
   * Ensure users are in the chat group
   * @param {number} groupId - The chat group ID
   * @param {number[]} userIds - Array of user IDs to add
   * @returns {Promise<number>} Number of newly added users
   */
  const ensureUsersInGroup = async (groupId, userIds) => {
    if (userIds.length === 0) return 0

    try {
      // Get current members
      const currentMembersResult = await callApi({
        query: `SELECT id_user FROM chat_users WHERE id_chat_group = ? AND is_active = 1 AND id_user IS NOT NULL`,
        params: [groupId],
        requiresAuth: false,
      })

      const currentMemberIds = new Set()
      if (currentMembersResult.success && currentMembersResult.data) {
        currentMembersResult.data.forEach((m) => {
          if (m.id_user) currentMemberIds.add(m.id_user)
        })
      }

      // Find users to add
      const usersToAdd = userIds.filter((id) => !currentMemberIds.has(id))

      if (usersToAdd.length > 0) {
        await addUsersToGroup(groupId, usersToAdd)
        console.log(`Added ${usersToAdd.length} users to group ${groupId}`)
        return usersToAdd.length // Return count of newly added users
      }
      
      return 0 // No new users added
    } catch (err) {
      console.error('Error ensuring users in group:', err)
      return 0
    }
  }

  /**
   * Ensure client is in the chat group
   * @param {number} groupId - The chat group ID
   * @param {number} clientId - The client ID
   */
  const ensureClientInGroup = async (groupId, clientId) => {
    try {
      // Check if client is already in group
      const checkResult = await callApi({
        query: `SELECT id FROM chat_users WHERE id_chat_group = ? AND id_client = ? AND is_active = 1`,
        params: [groupId, clientId],
        requiresAuth: false,
      })

      if (!checkResult.success || !checkResult.data || checkResult.data.length === 0) {
        // Add client to group
        await addClientToGroup(groupId, clientId)
        console.log(`Added client ${clientId} to group ${groupId}`)
      }
    } catch (err) {
      console.error('Error ensuring client in group:', err)
    }
  }

  /**
   * Add users to a chat group
   * @param {number} groupId - The chat group ID
   * @param {number[]} userIds - Array of user IDs to add
   */
  const addUsersToGroup = async (groupId, userIds) => {
    if (userIds.length === 0) return

    try {
      const insertQueries = userIds
        .map(() => `INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)`)
        .join('; ')

      const params = userIds.flatMap((userId) => [userId, groupId])

      await callApi({
        query: insertQueries,
        params: params,
        requiresAuth: false,
      })
    } catch (err) {
      console.error('Error adding users to group:', err)
      throw err
    }
  }

  /**
   * Add client to a chat group
   * @param {number} groupId - The chat group ID
   * @param {number} clientId - The client ID
   */
  const addClientToGroup = async (groupId, clientId) => {
    try {
      await callApi({
        query: `INSERT INTO chat_users (id_client, id_chat_group, is_active) VALUES (?, ?, 1)`,
        params: [clientId, groupId],
        requiresAuth: false,
      })
    } catch (err) {
      console.error('Error adding client to group:', err)
      throw err
    }
  }

  /**
   * Find all admin users (role_id = 1)
   * @returns {Promise<number[]>} Array of admin user IDs
   */
  const findAllAdmins = async () => {
    try {
      const result = await callApi({
        query: `SELECT id FROM users WHERE role_id = 1 AND id IS NOT NULL`,
        params: [],
        requiresAuth: false,
      })

      if (!result.success || !result.data) {
        return []
      }

      const adminIds = result.data.map((row) => parseInt(row.id)).filter((id) => !isNaN(id) && id > 0)
      console.log(`Found ${adminIds.length} admin users:`, adminIds)
      return adminIds
    } catch (err) {
      console.error('Error finding admin users:', err)
      return []
    }
  }

  /**
   * Send an automatic message from the client to the chat group
   * @param {number} groupId - The chat group ID
   * @param {number} clientId - The client ID
   * @param {string} carName - The car name for the message
   */
  const sendAutomaticMessage = async (groupId, clientId, carName) => {
    try {
      // Check if message already exists (to avoid duplicates)
      const checkResult = await callApi({
        query: `
          SELECT id FROM chat_messages 
          WHERE id_chat_group = ? 
            AND message_from_client_id = ? 
            AND message LIKE ?
          ORDER BY time DESC
          LIMIT 1
        `,
        params: [groupId, clientId, '%please assign this car to a staff%'],
        requiresAuth: false,
      })

      // If message already exists, don't send again
      if (checkResult.success && checkResult.data && checkResult.data.length > 0) {
        console.log('Automatic message already sent, skipping')
        return
      }

      // Send the automatic message
      // Note: This message is sent from the client, so we use a simple English message
      // In a real scenario, you might want to use i18n, but since this is server-side,
      // we'll use a simple English message that can be understood by admins
      const messageText = `Please assign this car to a staff`
      const result = await callApi({
        query: `
          INSERT INTO chat_messages (id_chat_group, message_from_client_id, message, time) 
          VALUES (?, ?, ?, UTC_TIMESTAMP())
        `,
        params: [groupId, clientId, messageText],
        requiresAuth: false,
      })

      if (result.success) {
        console.log(`Automatic message sent to group ${groupId}: "${messageText}"`)
      } else {
        console.error('Failed to send automatic message')
      }
    } catch (err) {
      console.error('Error sending automatic message:', err)
    }
  }

  return {
    createOrGetCarClientChatGroup,
    loading,
    error,
  }
}

