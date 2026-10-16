import { ref, computed } from 'vue'
import { useApi } from './useApi'

// Shared group-list logic for the chat surfaces. Extracted from ChatSidebar so
// the desktop sidebar and the mobile list render the exact same data/behaviour
// from one source of truth (fetch, filter, select, leave, readiness).
export function useChatGroups(props, { onGroupSelected, t } = {}) {
  const { callApi, loading, error } = useApi()

  // Localized lookup with a safe identity fallback so the composable works even
  // if a host forgets to pass `t` (the component still supplies English defaults).
  const translate = t || ((key) => key)


  const chatGroups = ref([])
  const allChatGroups = ref([]) // Store all groups for filtering
  const selectedGroup = ref(null)
  const currentUser = ref(null)
  const currentClient = ref(null)
  const searchFilter = ref('')
  const isClientMode = computed(() => props.clientId !== null)

  // Popup state
  const showUsersPopup = ref(false)
  const selectedGroupForPopup = ref(null)
  const showAddGroupModal = ref(false)

  // Readiness signal so parents can wait for the initial groups fetch to finish
  // instead of relying on arbitrary timeouts.
  const groupsReady = ref(false)
  let resolveGroupsReady
  const groupsReadyPromise = new Promise((resolve) => {
    resolveGroupsReady = resolve
  })

  const getCurrentUser = () => {
    if (isClientMode.value) {
      // In client mode, set client info
      currentClient.value = { id: props.clientId, name: props.clientName }
      return
    }

    try {
      const userStr = localStorage.getItem('user')
      if (userStr) {
        currentUser.value = JSON.parse(userStr)
      }
    } catch (err) {
      // Error getting current user - silently fail
    }
  }

  const fetchChatGroups = async () => {
    try {
      if (isClientMode.value) {
        if (!currentClient.value) {
          return
        }
      } else {
        if (!currentUser.value) {
          return
        }
      }

      let query, params, requiresAuth

      if (isClientMode.value) {
        // Client mode: fetch groups where client is a member
        query = `
          SELECT DISTINCT
            cg.id, 
            cg.name, 
            cg.description, 
            cg.id_user_owner,
            cg.id_client_owner,
            cg.is_active,
            COALESCE(MAX(cm.time), '1970-01-01 00:00:00') as latest_message_time,
            CONCAT_WS(', ',
              NULLIF(GROUP_CONCAT(DISTINCT u.username ORDER BY u.username SEPARATOR ', '), ''),
              NULLIF(GROUP_CONCAT(DISTINCT c.name ORDER BY c.name SEPARATOR ', '), '')
            ) as user_names,
            SUBSTRING_INDEX(
              GROUP_CONCAT(DISTINCT cm.message ORDER BY cm.time DESC SEPARATOR ' '), 
              ' ', 
              100
            ) as recent_messages
          FROM chat_groups cg
          INNER JOIN chat_users cu_member ON cg.id = cu_member.id_chat_group
          LEFT JOIN chat_users cu ON cg.id = cu.id_chat_group AND cu.is_active = 1
          LEFT JOIN users u ON cu.id_user = u.id
          LEFT JOIN clients c ON cu.id_client = c.id
          LEFT JOIN chat_messages cm ON cg.id = cm.id_chat_group
          WHERE cg.is_active = 1 
            AND cu_member.is_active = 1
            AND cu_member.id_client = ?
          GROUP BY cg.id, cg.name, cg.description, cg.id_user_owner, cg.id_client_owner, cg.is_active
          ORDER BY latest_message_time DESC, cg.name ASC
        `
        params = [currentClient.value.id]
        requiresAuth = false
      } else {
        // User mode: fetch groups where user is a member
        query = `
          SELECT DISTINCT
            cg.id, 
            cg.name, 
            cg.description, 
            cg.id_user_owner,
            cg.id_client_owner,
            cg.is_active,
            COALESCE(MAX(cm.time), '1970-01-01 00:00:00') as latest_message_time,
            CONCAT_WS(', ',
              NULLIF(GROUP_CONCAT(DISTINCT u.username ORDER BY u.username SEPARATOR ', '), ''),
              NULLIF(GROUP_CONCAT(DISTINCT c.name ORDER BY c.name SEPARATOR ', '), '')
            ) as user_names,
            SUBSTRING_INDEX(
              GROUP_CONCAT(DISTINCT cm.message ORDER BY cm.time DESC SEPARATOR ' '), 
              ' ', 
              100
            ) as recent_messages
          FROM chat_groups cg
          INNER JOIN chat_users cu_member ON cg.id = cu_member.id_chat_group
          LEFT JOIN chat_users cu ON cg.id = cu.id_chat_group AND cu.is_active = 1
          LEFT JOIN users u ON cu.id_user = u.id
          LEFT JOIN clients c ON cu.id_client = c.id
          LEFT JOIN chat_messages cm ON cg.id = cm.id_chat_group
          WHERE cg.is_active = 1 
            AND cu_member.is_active = 1
            AND cu_member.id_user = ?
          GROUP BY cg.id, cg.name, cg.description, cg.id_user_owner, cg.id_client_owner, cg.is_active
          ORDER BY latest_message_time DESC, cg.name ASC
        `
        params = [currentUser.value.id]
        requiresAuth = true
      }

      const result = await callApi({
        query,
        params,
        requiresAuth,
      })

      if (result.success && result.data) {
        allChatGroups.value = result.data
        applyFilter()
      } else {
        allChatGroups.value = []
        chatGroups.value = []
      }
    } catch (err) {
      chatGroups.value = []
    }
  }

  const selectGroup = async (group) => {
    selectedGroup.value = group
    if (onGroupSelected) {
      onGroupSelected(group)
    }
  }

  const showUsers = (group, event) => {
    if (event) {
      event.stopPropagation() // Prevent group selection
    }
    selectedGroupForPopup.value = group
    showUsersPopup.value = true
  }

  const closeUsersPopup = () => {
    showUsersPopup.value = false
    selectedGroupForPopup.value = null
  }

  const showAddGroup = () => {
    showAddGroupModal.value = true
  }

  const closeAddGroupModal = () => {
    showAddGroupModal.value = false
  }

  const handleGroupAdded = () => {
    // Refresh the groups list after adding a new group
    fetchChatGroups()
  }

  const selectGroupById = async (groupId) => {
    // Find the group in the current list
    const group = chatGroups.value.find((g) => g.id == groupId)

    if (group) {
      selectGroup(group)
      return group
    } else {
      // If group not found, refresh the list and try again
      await fetchChatGroups()
      const refreshedGroup = chatGroups.value.find((g) => g.id == groupId)
      if (refreshedGroup) {
        selectGroup(refreshedGroup)
        return refreshedGroup
      }
    }
    return null
  }

  const getNewMessageCount = (groupId) => {
    return props.newMessagesCounts?.[groupId] || 0
  }

  const applyFilter = () => {
    const filter = searchFilter.value.trim().toLowerCase()
    if (!filter) {
      chatGroups.value = allChatGroups.value
      return
    }

    chatGroups.value = allChatGroups.value.filter((group) => {
      const name = (group.name || '').toLowerCase()
      const description = (group.description || '').toLowerCase()
      const users = (group.user_names || '').toLowerCase()
      const recent = (group.recent_messages || '').toLowerCase()
      return (
        name.includes(filter) ||
        description.includes(filter) ||
        users.includes(filter) ||
        recent.includes(filter)
      )
    })
  }

  const handleSearchInput = () => {
    applyFilter()
  }

  const clearFilter = () => {
    searchFilter.value = ''
    applyFilter()
  }

  const leaveGroup = async (group, event) => {
    if (event) {
      event.stopPropagation()
    }

    if (!currentUser.value) {
      return
    }

    // Check if user is the owner
    if (group.id_user_owner === currentUser.value.id) {
      if (
        !confirm(
          translate('chat.confirmLeaveAsOwner') ||
            `You are the owner of this group. Are you sure you want to leave? The group will remain active but you will no longer be a member.`,
        )
      ) {
        return
      }
    } else {
      if (
        !confirm(
          translate('chat.confirmLeaveGroup', { groupName: group.name }) ||
            `Are you sure you want to leave "${group.name}"?`,
        )
      ) {
        return
      }
    }

    try {
      const result = await callApi({
        query: `
          UPDATE chat_users 
          SET is_active = 0 
          WHERE id_chat_group = ? AND id_user = ? AND is_active = 1
        `,
        params: [group.id, currentUser.value.id],
        requiresAuth: true,
      })

      if (result.success) {
        // If the user left the currently selected group, deselect it
        if (selectedGroup.value && selectedGroup.value.id === group.id) {
          selectGroup(null)
        }

        // Refresh groups to remove it from the list
        await fetchChatGroups()


      } else {

      }
    } catch (err) {
    }
  }

  const canLeaveGroup = (group) => {
    if (isClientMode.value) {
      return false
    }
    return true
  }

  const refreshGroups = () => fetchChatGroups()

  const getChatGroups = () => chatGroups.value

  // Called by the host component on mount.
  const init = async () => {
    await getCurrentUser()
    await fetchChatGroups()
    groupsReady.value = true
    resolveGroupsReady()
  }

  const whenReady = () => groupsReadyPromise
  const isReady = () => groupsReady.value

  return {
    // state
    chatGroups,
    allChatGroups,
    selectedGroup,
    currentUser,
    currentClient,
    searchFilter,
    isClientMode,
    showUsersPopup,
    selectedGroupForPopup,
    showAddGroupModal,
    groupsReady,
    loading,
    error,
    // methods
    init,
    getCurrentUser,
    fetchChatGroups,
    selectGroup,
    selectGroupById,
    showUsers,
    closeUsersPopup,
    showAddGroup,
    closeAddGroupModal,
    handleGroupAdded,
    applyFilter,
    handleSearchInput,
    clearFilter,
    leaveGroup,
    canLeaveGroup,
    getNewMessageCount,
    refreshGroups,
    getChatGroups,
    whenReady,
    isReady,
  }
}
