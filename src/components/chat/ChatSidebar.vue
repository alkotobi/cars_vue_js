<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import ChatUsersPopup from './ChatUsersPopup.vue'
import AddGroup from './AddGroup.vue'

const { callApi, loading, error } = useApi()
const chatGroups = ref([])
const selectedGroup = ref(null)

// Popup state
const showUsersPopup = ref(false)
const selectedGroupForPopup = ref(null)
const showAddGroupModal = ref(false)

// Emit events to parent
const emit = defineEmits(['group-selected'])

onMounted(async () => {
  await fetchChatGroups()
})

const fetchChatGroups = async () => {
  try {
    // Get current user first
    const userStr = localStorage.getItem('user')
    if (!userStr) {
      console.log('No user found in localStorage')
      return
    }

    const currentUser = JSON.parse(userStr)
    console.log('Current user for group fetch:', currentUser)

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
      params: [currentUser.id],
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

const selectGroup = (group) => {
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
            <td>{{ group.name }}</td>
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
