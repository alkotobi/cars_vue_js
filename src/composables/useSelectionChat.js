import { ref } from 'vue'
import { useApi } from './useApi'

export function useSelectionChat() {
  const { callApi } = useApi()
  const loading = ref(false)
  const error = ref(null)

  /**
   * Get all user IDs who have access to this selection
   */
  const getSelectionUsers = (selection) => {
    const userIds = []
    
    // Get users from owned_by field
    if (selection.owned_by) {
      const ownedBy = typeof selection.owned_by === 'string' 
        ? JSON.parse(selection.owned_by) 
        : selection.owned_by
      
      if (Array.isArray(ownedBy)) {
        userIds.push(...ownedBy.map(id => parseInt(id)))
      }
    }
    
    // Get current user
    const userStr = localStorage.getItem('user')
    if (userStr) {
      try {
        const user = JSON.parse(userStr)
        if (user && user.id) {
          const userId = parseInt(user.id)
          if (!userIds.includes(userId)) {
            userIds.push(userId)
          }
        }
      } catch (e) {
        console.error('Error parsing user:', e)
      }
    }
    
    // Remove duplicates
    return [...new Set(userIds)]
  }

  /**
   * Get group name for a selection
   */
  const getGroupName = (selection) => {
    return `Selection: ${selection.name} (ID: ${selection.id})`
  }

  /**
   * Check if a group chat exists for this selection
   */
  const checkGroupExists = async (selection) => {
    try {
      const groupName = getGroupName(selection)
      
      const result = await callApi({
        query: `
          SELECT id, name
          FROM chat_groups
          WHERE name = ? AND is_active = 1
        `,
        params: [groupName],
        requiresAuth: true,
      })

      if (result.success && result.data && result.data.length > 0) {
        return result.data[0]
      }
      
      return null
    } catch (err) {
      console.error('Error checking group existence:', err)
      return null
    }
  }

  /**
   * Check if current user is a member of the group
   */
  const isUserInGroup = async (groupId, userId) => {
    try {
      const result = await callApi({
        query: `
          SELECT id
          FROM chat_users
          WHERE id_chat_group = ? AND id_user = ? AND is_active = 1
        `,
        params: [groupId, userId],
        requiresAuth: true,
      })

      return result.success && result.data && result.data.length > 0
    } catch (err) {
      console.error('Error checking user membership:', err)
      return false
    }
  }

  /**
   * Create a group chat for a selection
   */
  const createGroupForSelection = async (selection) => {
    try {
      loading.value = true
      error.value = null

      const userStr = localStorage.getItem('user')
      if (!userStr) {
        throw new Error('User session not found')
      }

      const currentUser = JSON.parse(userStr)
      const userIds = getSelectionUsers(selection)
      const groupName = getGroupName(selection)
      const groupDescription = `Group chat for selection: ${selection.name}`

      // Create the chat group
      const groupResult = await callApi({
        query: `
          INSERT INTO chat_groups (name, description, id_user_owner, is_active) 
          VALUES (?, ?, ?, 1)
        `,
        params: [groupName, groupDescription, currentUser.id],
        requiresAuth: true,
      })

      if (!groupResult.success || !groupResult.lastInsertId) {
        throw new Error('Failed to create chat group')
      }

      const groupId = groupResult.lastInsertId

      // Add all users to the group
      if (userIds.length > 0) {
        const insertQueries = userIds
          .map(() => 'INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)')
          .join('; ')

        const params = userIds.flatMap((userId) => [userId, groupId])

        const usersResult = await callApi({
          query: insertQueries,
          params: params,
          requiresAuth: true,
        })

        if (!usersResult.success) {
          console.error('Failed to add users to group, but group was created')
        }
      }

      return { success: true, groupId }
    } catch (err) {
      error.value = err.message
      console.error('Error creating group:', err)
      return { success: false, error: err.message }
    } finally {
      loading.value = false
    }
  }

  /**
   * Add current user to a group if not already a member
   */
  const joinGroup = async (groupId) => {
    try {
      loading.value = true
      error.value = null

      const userStr = localStorage.getItem('user')
      if (!userStr) {
        throw new Error('User session not found')
      }

      const currentUser = JSON.parse(userStr)
      const userId = parseInt(currentUser.id)

      // Check if already a member
      const isMember = await isUserInGroup(groupId, userId)
      if (isMember) {
        return { success: true, alreadyMember: true }
      }

      // Add user to group
      const result = await callApi({
        query: `
          INSERT INTO chat_users (id_user, id_chat_group, is_active) 
          VALUES (?, ?, 1)
        `,
        params: [userId, groupId],
        requiresAuth: true,
      })

      if (result.success) {
        return { success: true, alreadyMember: false }
      } else {
        throw new Error('Failed to join group')
      }
    } catch (err) {
      error.value = err.message
      console.error('Error joining group:', err)
      return { success: false, error: err.message }
    } finally {
      loading.value = false
    }
  }

  /**
   * Ensure all selection users are in the group (add missing ones)
   */
  const ensureAllUsersInGroup = async (groupId, selection) => {
    try {
      loading.value = true
      error.value = null

      const userIds = getSelectionUsers(selection)
      const missingUserIds = []

      // Check which users are missing
      for (const userId of userIds) {
        const isMember = await isUserInGroup(groupId, userId)
        if (!isMember) {
          missingUserIds.push(userId)
        }
      }

      // Add missing users
      if (missingUserIds.length > 0) {
        const insertQueries = missingUserIds
          .map(() => 'INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)')
          .join('; ')

        const params = missingUserIds.flatMap((userId) => [userId, groupId])

        const result = await callApi({
          query: insertQueries,
          params: params,
          requiresAuth: true,
        })

        if (!result.success) {
          console.error('Failed to add some users to group')
        }
      }

      return { success: true, addedUsers: missingUserIds.length }
    } catch (err) {
      error.value = err.message
      console.error('Error ensuring users in group:', err)
      return { success: false, error: err.message }
    } finally {
      loading.value = false
    }
  }

  /**
   * Main function to get or create group chat for a selection
   */
  const getOrCreateSelectionGroup = async (selection) => {
    try {
      loading.value = true
      error.value = null

      // Check if group exists
      let group = await checkGroupExists(selection)
      let groupId

      if (group) {
        // Group exists, ensure current user is in it
        groupId = group.id
        const userStr = localStorage.getItem('user')
        if (userStr) {
          const currentUser = JSON.parse(userStr)
          const userId = parseInt(currentUser.id)
          const isMember = await isUserInGroup(groupId, userId)
          
          if (!isMember) {
            await joinGroup(groupId)
          }
        }

        // Also ensure all selection users are in the group
        await ensureAllUsersInGroup(groupId, selection)
      } else {
        // Create new group
        const result = await createGroupForSelection(selection)
        if (result.success) {
          groupId = result.groupId
          group = { id: groupId, name: getGroupName(selection) }
        } else {
          throw new Error(result.error || 'Failed to create group')
        }
      }

      return { success: true, group }
    } catch (err) {
      error.value = err.message
      console.error('Error getting or creating group:', err)
      return { success: false, error: err.message, group: null }
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    getOrCreateSelectionGroup,
    getSelectionUsers,
    getGroupName,
  }
}

