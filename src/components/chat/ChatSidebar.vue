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
  console.log('showAddGroup called')
  console.log('showAddGroupModal before:', showAddGroupModal.value)
  showAddGroupModal.value = true
  console.log('showAddGroupModal after:', showAddGroupModal.value)
}

const closeAddGroupModal = () => {
  console.log('closeAddGroupModal called')
  showAddGroupModal.value = false
}

const handleGroupAdded = () => {
  // Refresh the groups list after adding a new group
  fetchChatGroups()
}

const selectGroupById = async (groupId) => {
  console.log('ChatSidebar: selectGroupById called with:', groupId)

  // Find the group in the current list
  const group = chatGroups.value.find((g) => g.id == groupId)

  if (group) {
    console.log('Found group:', group)
    selectGroup(group)
    return group
  } else {
    console.log('Group not found in current list, refreshing groups...')
    // If group not found, refresh the list and try again
    await fetchChatGroups()
    const refreshedGroup = chatGroups.value.find((g) => g.id == groupId)
    if (refreshedGroup) {
      console.log('Found group after refresh:', refreshedGroup)
      selectGroup(refreshedGroup)
      return refreshedGroup
    }
  }

  console.log('Group not found with ID:', groupId)
  return null
}

const getNewMessageCount = (groupId) => {
  const count = props.newMessagesCounts[groupId] || 0
  console.log(`Getting new message count for group ${groupId}:`, count)
  console.log('All newMessagesCounts:', props.newMessagesCounts)
  return count
}

// Expose methods to parent
defineExpose({
  cleanup: () => {
    // Cleanup function if needed
  },
  selectGroupById,
})
</script>

<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3><i class="fas fa-comments"></i> Chat Groups</h3>
      <div class="header-buttons">
        <button @click="showAddGroup" class="add-group-btn" title="Add New Group">
          <i class="fas fa-plus"></i>
        </button>
      </div>
    </div>
    <div class="groups-table">
      <!-- Desktop Table View -->
      <div class="desktop-table">
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
      </div>

      <!-- Mobile Cards View -->
      <div class="mobile-cards">
        <div
          v-for="group in chatGroups"
          :key="group.id"
          @click="selectGroup(group)"
          :class="{ selected: selectedGroup?.id === group.id }"
          class="group-card"
        >
          <div class="card-header">
            <div class="group-name-container">
              <span class="group-name">{{ group.name }}</span>
              <span v-if="getNewMessageCount(group.id) > 0" class="new-messages-badge">
                {{ getNewMessageCount(group.id) > 99 ? '99+' : getNewMessageCount(group.id) }}
              </span>
            </div>
            <button @click="showUsers(group, $event)" class="users-btn" title="View Users">
              <i class="fas fa-users"></i>
            </button>
          </div>
          <div class="card-description">
            {{ group.description || 'No description' }}
          </div>
        </div>
      </div>

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

.desktop-table {
  display: block;
}

.mobile-cards {
  display: none;
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

  .desktop-table {
    display: none;
  }

  .mobile-cards {
    display: block;
    padding: 16px;
  }

  /* Mobile Cards Styles */
  .group-card {
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 16px;
    margin-bottom: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .group-card:hover {
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.15);
    transform: translateY(-1px);
  }

  .group-card.selected {
    background-color: #e0f2fe;
    border-color: #06b6d4;
    box-shadow: 0 2px 8px rgba(6, 182, 212, 0.2);
  }

  .group-card .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 8px;
  }

  .group-card .group-name-container {
    flex: 1;
    min-width: 0;
  }

  .group-card .group-name {
    font-weight: 600;
    color: #1f2937;
    font-size: 1rem;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .group-card .new-messages-badge {
    margin-left: 8px;
    flex-shrink: 0;
  }

  .group-card .users-btn {
    flex-shrink: 0;
    margin-left: 8px;
  }

  .group-card .card-description {
    color: #6b7280;
    font-size: 0.9rem;
    line-height: 1.4;
    word-break: break-word;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  /* Mobile sidebar header adjustments */
  .sidebar-header {
    padding: 16px;
  }

  .sidebar-header h3 {
    font-size: 1.1rem;
  }

  .add-group-btn {
    padding: 6px 10px;
    font-size: 0.9rem;
  }

  /* Mobile loading and error states */
  .loading,
  .error,
  .no-groups {
    padding: 16px;
    margin: 16px;
    border-radius: 8px;
  }

  .no-groups {
    background: white;
    border: 1px solid #e2e8f0;
  }

  .no-groups i {
    font-size: 2.5rem;
  }

  .no-groups p {
    font-size: 1rem;
  }

  .btn {
    padding: 10px 16px;
    font-size: 0.9rem;
  }
}

/* Touch-friendly improvements */
@media (max-width: 768px) {
  .group-card {
    min-height: 60px;
  }

  .users-btn {
    min-width: 44px;
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

/* Landscape orientation adjustments */
@media (max-width: 768px) and (orientation: landscape) {
  .chat-sidebar {
    height: 150px;
  }

  .mobile-cards {
    padding: 12px;
  }

  .group-card {
    padding: 12px;
    margin-bottom: 8px;
  }

  .group-card .card-description {
    -webkit-line-clamp: 1;
  }
}
</style>
