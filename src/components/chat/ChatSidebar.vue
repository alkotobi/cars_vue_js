<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../../composables/useApi'
import { useEnhancedI18n } from '../../composables/useI18n'
import ChatUsersPopup from './ChatUsersPopup.vue'
import AddGroup from './AddGroup.vue'

const { t } = useEnhancedI18n()

const props = defineProps({
  newMessagesCounts: {
    type: Object,
    default: () => ({}),
  },
  // Client mode props (optional)
  clientId: {
    type: Number,
    default: null,
  },
  clientName: {
    type: String,
    default: null,
  },
})

const { callApi, loading, error } = useApi()
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

// Emit events to parent
const emit = defineEmits(['group-selected'])

onMounted(async () => {
  await getCurrentUser()
  await fetchChatGroups()
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
  emit('group-selected', group)
}

const showUsers = (group, event) => {
  event.stopPropagation() // Prevent group selection
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
  const count = props.newMessagesCounts[groupId] || 0
  return count
}

// Method to refresh groups (can be called from outside)
const refreshGroups = async () => {
  await fetchChatGroups()
}

// Filter groups based on search term
const applyFilter = () => {
  const searchTerm = searchFilter.value.trim().toLowerCase()
  
  if (!searchTerm) {
    chatGroups.value = [...allChatGroups.value]
    return
  }

  chatGroups.value = allChatGroups.value.filter((group) => {
    // Search in name
    const nameMatch = group.name?.toLowerCase().includes(searchTerm)
    
    // Search in description
    const descriptionMatch = group.description?.toLowerCase().includes(searchTerm)
    
    // Search in user names
    const userNamesMatch = group.user_names?.toLowerCase().includes(searchTerm)
    
    // Search in recent messages (limit to first 1000 chars to avoid performance issues)
    const messagesMatch = group.recent_messages?.toLowerCase().includes(searchTerm)
    
    return nameMatch || descriptionMatch || userNamesMatch || messagesMatch
  })
}

// Watch for search filter changes
const handleSearchInput = () => {
  applyFilter()
}

// Clear search filter
const clearFilter = () => {
  searchFilter.value = ''
  applyFilter()
}

// Leave a group
const leaveGroup = async (group, event) => {
  event.stopPropagation() // Prevent group selection
  
  if (!currentUser.value || !currentUser.value.id) {
    alert(t('chat.noCurrentUser') || 'No current user available')
    return
  }

  // Check if user is the owner
  if (group.id_user_owner === currentUser.value.id) {
    if (!confirm(t('chat.confirmLeaveAsOwner') || `You are the owner of this group. Are you sure you want to leave? The group will remain active but you will no longer be a member.`)) {
      return
    }
  } else {
    if (!confirm(t('chat.confirmLeaveGroup', { groupName: group.name }) || `Are you sure you want to leave "${group.name}"?`)) {
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
        selectedGroup.value = null
        emit('group-selected', null)
      }
      
      // Refresh groups to remove it from the list
      await fetchChatGroups()
      
      alert(t('chat.leftGroupSuccess', { groupName: group.name }) || `You have left "${group.name}"`)
    } else {
      alert(t('chat.failedToLeaveGroup') || 'Failed to leave group')
    }
  } catch (err) {
    alert(t('chat.errorLeavingGroup') || 'Error leaving group')
  }
}

// Check if user can leave a group (not for clients)
const canLeaveGroup = (group) => {
  if (isClientMode.value) {
    return false // Clients cannot leave groups
  }
  // Users can always leave groups they're members of
  return true
}

// Expose methods to parent
defineExpose({
  cleanup: () => {
    // Cleanup function if needed
  },
  selectGroupById,
  refreshGroups,
  fetchChatGroups,
  getChatGroups: () => chatGroups.value,
})
</script>

<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3><i class="fas fa-comments"></i> {{ t('navigation.teams') || 'Groups' }}</h3>
      <div class="header-buttons">
        <button v-if="!isClientMode" @click="showAddGroup" class="add-group-btn" :title="t('chat.addNewGroup')">
          <i class="fas fa-plus"></i>
        </button>
      </div>
    </div>
    
    <!-- Search Filter -->
    <div class="search-filter-container">
      <div class="search-input-wrapper">
        <i class="fas fa-search search-icon"></i>
        <input
          type="text"
          v-model="searchFilter"
          @input="handleSearchInput"
          :placeholder="t('chat.searchGroups')"
          class="search-input"
        />
        <button
          v-if="searchFilter"
          @click="clearFilter"
          class="clear-search-btn"
          :title="t('chat.clearSearch')"
        >
          <i class="fas fa-times"></i>
        </button>
      </div>
      <div v-if="searchFilter" class="filter-info">
        {{ chatGroups.length }} / {{ allChatGroups.length }} groups
      </div>
    </div>

    <div class="groups-table">
      <table>
        <thead>
          <tr>
            <th>{{ t('cars.teamName') || 'Group Name' }}</th>
            <th>{{ t('cars.description') || 'Description' }}</th>
            <th>{{ t('cars.actions') || 'Actions' }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="group in chatGroups"
            :key="group.id"
            @click="selectGroup(group)"
            :class="{ selected: selectedGroup?.id === group.id }"
            class="group-row"
          >
            <td>
              <div class="group-name-container">
                <span class="group-name">{{ group.name }}</span>
                <span v-if="getNewMessageCount(group.id) > 0" class="new-messages-badge">
                  {{ getNewMessageCount(group.id) > 99 ? '99+' : getNewMessageCount(group.id) }}
                </span>
              </div>
            </td>
            <td>{{ group.description || (t('cars.description') || 'No description') }}</td>
            <td>
              <div class="action-buttons">
                <button @click="showUsers(group, $event)" class="users-btn" :title="t('chat.members')">
                  <i class="fas fa-users"></i>
                </button>
                <button 
                  v-if="!isClientMode && canLeaveGroup(group)" 
                  @click="leaveGroup(group, $event)" 
                  class="leave-btn" 
                  :title="t('chat.leaveGroup')"
                >
                  <i class="fas fa-sign-out-alt"></i>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i> {{ t('chat.loadingGroups') }}
      </div>
      <div v-if="error" class="error"><i class="fas fa-exclamation-triangle"></i> {{ error }}</div>
      <div v-if="!loading && chatGroups.length === 0 && !searchFilter" class="no-groups">
        <i class="fas fa-comments"></i>
        <p>{{ t('chat.noGroupsFound') }}</p>
        <button v-if="!isClientMode" @click="showAddGroup" class="btn btn-primary">
          <i class="fas fa-plus"></i> {{ t('chat.addNewGroup') }}
        </button>
      </div>
      <div v-if="!loading && chatGroups.length === 0 && searchFilter" class="no-results">
        <i class="fas fa-search"></i>
        <p>{{ t('chat.noGroupsFound') }}</p>
        <button @click="clearFilter" class="btn btn-secondary">
          {{ t('chat.clearSearch') }}
        </button>
      </div>
    </div>

    <!-- Users Popup -->
    <ChatUsersPopup
      v-if="selectedGroupForPopup"
      :group-id="selectedGroupForPopup.id"
      :group-name="selectedGroupForPopup.name"
      :is-visible="showUsersPopup"
      @close="closeUsersPopup"
    />

    <!-- Add Group Modal (only for users) -->
    <AddGroup
      v-if="!isClientMode"
      :is-visible="showAddGroupModal"
      @close="closeAddGroupModal"
      @group-added="handleGroupAdded"
    />
  </div>
</template>

<style scoped>
.chat-sidebar {
  width: 400px;
  background-color: white;
  border-right: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #06b6d4;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.search-filter-container {
  padding: 12px 16px;
  border-bottom: 1px solid #e2e8f0;
  background-color: #f8fafc;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-icon {
  position: absolute;
  left: 12px;
  color: #64748b;
  font-size: 14px;
  pointer-events: none;
}

.search-input {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 14px;
  color: #374151;
  background-color: white;
  transition: all 0.2s ease;
}

.search-input:focus {
  outline: none;
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.search-input::placeholder {
  color: #9ca3af;
}

.clear-search-btn {
  position: absolute;
  right: 8px;
  background: none;
  border: none;
  color: #64748b;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
  font-size: 12px;
}

.clear-search-btn:hover {
  background-color: #e5e7eb;
  color: #374151;
}

.filter-info {
  margin-top: 8px;
  font-size: 12px;
  color: #64748b;
  text-align: right;
}

.header-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.debug-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.debug-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.add-group-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 8px 12px;
  border-radius: 4px;
  transition: background-color 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.add-group-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.groups-table {
  flex: 1;
  overflow-y: auto;
}

.groups-table table {
  width: 100%;
  border-collapse: collapse;
}

.groups-table th {
  background-color: #f1f5f9;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
}

.groups-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #374151;
}

.group-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.group-row:hover {
  background-color: #f8fafc;
}

.group-row.selected {
  background-color: #e0f2fe;
  border-left: 3px solid #06b6d4;
}

.group-name-container {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.group-name {
  flex: 1;
}

.new-messages-badge {
  background-color: #ef4444;
  color: white;
  border-radius: 12px;
  padding: 2px 8px;
  font-size: 0.7rem;
  font-weight: 600;
  min-width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    transform: scale(1);
  }
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.users-btn {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background-color 0.2s;
}

.users-btn:hover {
  background-color: #0891b2;
}

.leave-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background-color 0.2s;
}

.leave-btn:hover {
  background-color: #dc2626;
}

.loading,
.error,
.no-groups,
.no-results {
  padding: 20px;
  text-align: center;
  color: #64748b;
}

.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.no-results i {
  font-size: 2rem;
  color: #9ca3af;
}

.no-results p {
  margin: 0;
  font-size: 0.9rem;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background-color: #4b5563;
}

.error {
  color: #ef4444;
}

.no-groups {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 15px;
}

.no-groups i {
  font-size: 3rem;
  color: #06b6d4;
}

.no-groups p {
  margin: 0;
  font-size: 1.1rem;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-primary {
  background-color: #06b6d4;
  color: white;
}

.btn-primary:hover {
  background-color: #0891b2;
}

/* Responsive design */
@media (max-width: 768px) {
  .chat-sidebar {
    width: 100%;
    height: 200px;
  }
}
</style>
