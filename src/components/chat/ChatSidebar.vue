<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import ChatUsersPopup from './ChatUsersPopup.vue'
import AddGroup from './AddGroup.vue'

const { callApi, loading, error } = useApi()
const chatGroups = ref([])
const selectedGroup = ref(null)
const unreadCounts = ref({}) // Track unread counts per group
const currentUser = ref(null)

// Popup state
const showUsersPopup = ref(false)
const selectedGroupForPopup = ref(null)
const showAddGroupModal = ref(false)

// Emit events to parent
const emit = defineEmits(['group-selected'])

onMounted(async () => {
  await getCurrentUser()
  await fetchChatGroups()
  await fetchUnreadCounts()

  // Start polling for new messages
  startMessagePolling()
})

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
      console.log('Current user loaded for sidebar:', currentUser.value)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
  }
}

const fetchChatGroups = async () => {
  try {
    if (!currentUser.value) {
      console.log('No current user available')
      return
    }

    console.log('Current user for group fetch:', currentUser.value)

    const result = await callApi({
      query: `
        SELECT DISTINCT
          cg.id, 
          cg.name, 
          cg.description, 
          cg.id_user_owner, 
          cg.is_active 
        FROM chat_groups cg
        INNER JOIN chat_users cu ON cg.id = cu.id_chat_group
        WHERE cg.is_active = 1 
          AND cu.is_active = 1
          AND cu.id_user = ?
        ORDER BY cg.name ASC
      `,
      params: [currentUser.value.id],
      requiresAuth: true,
    })

    console.log('Chat groups result:', result)

    if (result.success && result.data) {
      chatGroups.value = result.data
      console.log('Chat groups loaded:', chatGroups.value)
    } else {
      console.log('No groups found or API failed:', result)
      chatGroups.value = []
    }
  } catch (err) {
    console.error('Error fetching chat groups:', err)
    chatGroups.value = []
  }
}

const fetchUnreadCounts = async () => {
  try {
    if (!currentUser.value) return

    // Get the last viewed timestamp for each group from localStorage
    const lastViewed = JSON.parse(localStorage.getItem('chat_last_viewed') || '{}')

    // For each group, get unread count based on its last viewed timestamp
    const counts = {}

    for (const group of chatGroups.value) {
      const lastViewedTime = lastViewed[group.id] || '1970-01-01 00:00:00'

      const result = await callApi({
        query: `
          SELECT COUNT(*) as unread_count
          FROM chat_messages cm
          WHERE cm.id_chat_group = ?
            AND cm.message_from_user_id != ?
            AND cm.time > ?
        `,
        params: [group.id, currentUser.value.id, lastViewedTime],
        requiresAuth: true,
      })

      if (result.success && result.data && result.data.length > 0) {
        counts[group.id] = parseInt(result.data[0].unread_count)
      } else {
        counts[group.id] = 0
      }
    }

    unreadCounts.value = counts
    console.log('Unread counts loaded:', unreadCounts.value)
  } catch (err) {
    console.error('Error fetching unread counts:', err)
  }
}

const selectGroup = (group) => {
  selectedGroup.value = group

  // Update last viewed timestamp for this group
  const lastViewed = JSON.parse(localStorage.getItem('chat_last_viewed') || '{}')
  lastViewed[group.id] = new Date().toISOString()
  localStorage.setItem('chat_last_viewed', JSON.stringify(lastViewed))

  // Clear unread count for selected group
  if (unreadCounts.value[group.id]) {
    unreadCounts.value[group.id] = 0
  }

  emit('group-selected', group)
}

const showUsers = (group, event) => {
  event.stopPropagation() // Prevent group selection
  console.log('Show users clicked for group:', group)
  selectedGroupForPopup.value = group
  showUsersPopup.value = true
  console.log(
    'Popup state set - selectedGroupForPopup:',
    selectedGroupForPopup.value,
    'showUsersPopup:',
    showUsersPopup.value,
  )
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

// Method to update unread count when new message is sent
const updateUnreadCount = (groupId) => {
  // If this is the currently selected group, mark it as viewed
  if (selectedGroup.value?.id === groupId) {
    const lastViewed = JSON.parse(localStorage.getItem('chat_last_viewed') || '{}')
    lastViewed[groupId] = new Date().toISOString()
    localStorage.setItem('chat_last_viewed', JSON.stringify(lastViewed))

    // Clear unread count for this group
    if (unreadCounts.value[groupId]) {
      unreadCounts.value[groupId] = 0
    }
  } else {
    // If it's not the selected group, increment the unread count
    if (!unreadCounts.value[groupId]) {
      unreadCounts.value[groupId] = 0
    }
    unreadCounts.value[groupId]++
  }
}

// Polling for new messages
let messagePollingInterval = null

const startMessagePolling = () => {
  messagePollingInterval = setInterval(async () => {
    await fetchUnreadCounts()
  }, 10000) // Check every 10 seconds
}

const stopMessagePolling = () => {
  if (messagePollingInterval) {
    clearInterval(messagePollingInterval)
    messagePollingInterval = null
  }
}

// Clean up on unmount
const cleanup = () => {
  stopMessagePolling()
}

// Expose cleanup function and updateUnreadCount method
defineExpose({ cleanup, updateUnreadCount })
</script>

<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3><i class="fas fa-comments"></i> Chat Groups</h3>
      <button @click="showAddGroup" class="add-group-btn" title="Add New Group">
        <i class="fas fa-plus"></i>
      </button>
    </div>
    <div class="groups-table">
      <table>
        <thead>
          <tr>
            <th>Group Name</th>
            <th>Description</th>
            <th>Actions</th>
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
                <span
                  v-if="
                    unreadCounts[group.id] &&
                    unreadCounts[group.id] > 0 &&
                    selectedGroup?.id !== group.id
                  "
                  class="unread-badge"
                >
                  {{ unreadCounts[group.id] > 99 ? '99+' : unreadCounts[group.id] }}
                </span>
              </div>
            </td>
            <td>{{ group.description || 'No description' }}</td>
            <td>
              <button @click="showUsers(group, $event)" class="users-btn" title="View Users">
                <i class="fas fa-users"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
      <div v-if="loading" class="loading">
        <i class="fas fa-spinner fa-spin"></i> Loading groups...
      </div>
      <div v-if="error" class="error"><i class="fas fa-exclamation-triangle"></i> {{ error }}</div>
      <div v-if="!loading && chatGroups.length === 0" class="no-groups">
        <i class="fas fa-comments"></i>
        <p>No chat groups available</p>
        <button @click="showAddGroup" class="btn btn-primary">
          <i class="fas fa-plus"></i> Create First Group
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

    <!-- Add Group Modal -->
    <AddGroup
      :is-visible="showAddGroupModal"
      @close="closeAddGroupModal"
      @group-added="handleGroupAdded"
    />
  </div>
</template>

<style scoped>
.chat-sidebar {
  width: 300px;
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
  position: relative;
}

.group-name {
  flex: 1;
}

.unread-badge {
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

.loading,
.error,
.no-groups {
  padding: 20px;
  text-align: center;
  color: #64748b;
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
