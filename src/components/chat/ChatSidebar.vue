<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import ChatUsersPopup from './ChatUsersPopup.vue'
import AddGroup from './AddGroup.vue'

const props = defineProps({
  newMessagesCounts: {
    type: Object,
    default: () => ({}),
  },
})

const { callApi, loading, error } = useApi()
const chatGroups = ref([])
const selectedGroup = ref(null)
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

const selectGroup = async (group) => {
  selectedGroup.value = group
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

const getNewMessageCount = (groupId) => {
  const count = props.newMessagesCounts[groupId] || 0
  console.log(`Getting new message count for group ${groupId}:`, count)
  console.log('All newMessagesCounts:', props.newMessagesCounts)
  return count
}

// Debug section state
const showDebugSection = ref(false)
const debugInfo = ref('')

const toggleDebugSection = () => {
  showDebugSection.value = !showDebugSection.value
  if (showDebugSection.value) {
    updateDebugInfo()
  }
}

const updateDebugInfo = () => {
  let info = `=== DEBUG: NEW MESSAGES COUNTS ===\n`
  info += `Current User: ${currentUser.value?.username} (ID: ${currentUser.value?.id})\n`
  info += `Selected Group: ${selectedGroup.value?.name || 'None'} (ID: ${selectedGroup.value?.id || 'None'})\n`
  info += `Total Groups: ${chatGroups.value.length}\n\n`

  info += `=== GROUPS AND NEW MESSAGE COUNTS ===\n`
  if (chatGroups.value.length === 0) {
    info += `No groups found\n`
  } else {
    chatGroups.value.forEach((group) => {
      const newCount = getNewMessageCount(group.id)
      const isSelected = selectedGroup.value?.id === group.id
      const status = isSelected ? 'SELECTED' : 'NOT SELECTED'
      info += `${group.name} (ID: ${group.id}) - ${status}\n`
      info += `  New Messages: ${newCount}\n`
      info += `  Badge Display: ${newCount > 0 ? 'SHOWING' : 'HIDDEN'}\n\n`
    })
  }

  info += `=== RAW COUNTS DATA ===\n`
  if (Object.keys(props.newMessagesCounts).length === 0) {
    info += `No new message counts data\n`
  } else {
    Object.keys(props.newMessagesCounts).forEach((groupId) => {
      const count = props.newMessagesCounts[groupId]
      const group = chatGroups.value.find((g) => g.id == groupId)
      const groupName = group ? group.name : `Group ${groupId}`
      info += `${groupName} (ID: ${groupId}): ${count} new message${count !== 1 ? 's' : ''}\n`
    })
  }

  info += `\n=== SYSTEM INFO ===\n`
  info += `Last Updated: ${new Date().toLocaleString()}\n`
  info += `User ID: ${currentUser.value?.id}\n`
  info += `Username: ${currentUser.value?.username}\n`
  info += `Props newMessagesCounts type: ${typeof props.newMessagesCounts}\n`
  info += `Props newMessagesCounts keys: ${Object.keys(props.newMessagesCounts)}\n`

  debugInfo.value = info
}

const testNewMessageCount = () => {
  console.log('=== TESTING NEW MESSAGE COUNT ===')
  console.log('Props newMessagesCounts:', props.newMessagesCounts)
  console.log('Props newMessagesCounts type:', typeof props.newMessagesCounts)
  console.log('Props newMessagesCounts keys:', Object.keys(props.newMessagesCounts))

  if (chatGroups.value.length > 0) {
    const firstGroup = chatGroups.value[0]
    const count = getNewMessageCount(firstGroup.id)
    console.log(`Test count for group ${firstGroup.name} (ID: ${firstGroup.id}):`, count)
  }

  updateDebugInfo()
}
</script>

<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3><i class="fas fa-comments"></i> Chat Groups</h3>
      <div class="header-buttons">
        <button @click="toggleDebugSection" class="debug-btn" title="Debug New Messages">
          <i class="fas fa-bug"></i>
        </button>
        <button @click="showAddGroup" class="add-group-btn" title="Add New Group">
          <i class="fas fa-plus"></i>
        </button>
      </div>
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
                <span v-if="getNewMessageCount(group.id) > 0" class="new-messages-badge">
                  {{ getNewMessageCount(group.id) > 99 ? '99+' : getNewMessageCount(group.id) }}
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
      v-if="showAddGroupModal"
      @close="closeAddGroupModal"
      @group-added="handleGroupAdded"
    />

    <!-- Debug Info Modal -->
    <div v-if="showDebugSection" class="debug-info-modal">
      <div class="debug-info-content">
        <h3>Debug: New Messages Counts</h3>
        <pre>{{ debugInfo }}</pre>
        <div class="debug-buttons">
          <button @click="testNewMessageCount" class="test-btn">Test Count</button>
          <button @click="toggleDebugSection">Close</button>
        </div>
      </div>
    </div>
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

.debug-info-modal {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
}

.debug-info-content {
  background-color: white;
  padding: 24px;
  border-radius: 12px;
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
}

.debug-info-content h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #06b6d4;
  border-bottom: 2px solid #e2e8f0;
  padding-bottom: 8px;
}

.debug-info-content pre {
  white-space: pre-wrap;
  word-break: break-word;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  background-color: #f8fafc;
  padding: 16px;
  border-radius: 6px;
  border: 1px solid #e2e8f0;
  max-height: 400px;
  overflow-y: auto;
}

.debug-info-content button {
  margin-top: 16px;
  padding: 10px 20px;
  border: none;
  border-radius: 6px;
  background-color: #06b6d4;
  color: white;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.debug-info-content button:hover {
  background-color: #0891b2;
}

.debug-buttons {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.test-btn {
  background-color: #06b6d4;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 8px 16px;
  cursor: pointer;
  font-weight: 500;
  transition: background-color 0.2s;
}

.test-btn:hover {
  background-color: #0891b2;
}
</style>
