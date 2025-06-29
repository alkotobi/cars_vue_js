<script setup>
import { ref, onMounted, watch } from 'vue'
import { useApi } from '../../composables/useApi'
import AddUserToGroup from './AddUserToGroup.vue'

const props = defineProps({
  groupId: {
    type: Number,
    required: true,
  },
  groupName: {
    type: String,
    required: true,
  },
  isVisible: {
    type: Boolean,
    default: false,
  },
})

const { callApi, loading, error } = useApi()
const chatUsers = ref([])

// User permissions and group info
const currentUser = ref(null)
const groupOwner = ref(null)
const canRemoveUsers = ref(false)

// Emit events to parent
const emit = defineEmits(['close'])

// Add user modal state
const showAddUserModal = ref(false)

// Define all functions first
const fetchCurrentUser = async () => {
  try {
    const userStr = localStorage.getItem('user')
    console.log('User string from localStorage:', userStr)

    if (!userStr) {
      console.log('No user found in localStorage')
      return
    }

    const userData = JSON.parse(userStr)
    console.log('Parsed user data:', userData)

    // Use the username from the stored user data
    const username = userData.username
    console.log('Using username:', username)

    const result = await callApi({
      query: `
        SELECT 
          u.id,
          u.username,
          u.email,
          r.role_name
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.username = ?
      `,
      params: [username],
      requiresAuth: true,
    })

    console.log('Current user API result:', result)

    if (result.success && result.data && result.data.length > 0) {
      currentUser.value = result.data[0]
      console.log('Current user loaded:', currentUser.value)
    } else {
      console.log('No current user found or API failed')
    }
  } catch (err) {
    console.error('Error fetching current user:', err)
  }
}

const fetchGroupInfo = async () => {
  try {
    console.log('Fetching group info for groupId:', props.groupId)

    const result = await callApi({
      query: `
        SELECT 
          cg.id,
          cg.name,
          cg.id_user_owner,
          u.username as owner_username
        FROM chat_groups cg
        LEFT JOIN users u ON cg.id_user_owner = u.id
        WHERE cg.id = ?
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('Group info API result:', result)

    if (result.success && result.data && result.data.length > 0) {
      groupOwner.value = result.data[0]
      console.log('Group owner loaded:', groupOwner.value)

      // Check if current user can remove users
      checkRemovePermissions()
    } else {
      console.log('No group info found or API failed')
    }
  } catch (err) {
    console.error('Error fetching group info:', err)
  }
}

const checkRemovePermissions = () => {
  console.log('Checking remove permissions...')
  console.log('Current user:', currentUser.value)
  console.log('Group owner:', groupOwner.value)

  if (!currentUser.value || !groupOwner.value) {
    canRemoveUsers.value = false
    console.log('Missing user or group data, canRemoveUsers set to false')
    return
  }

  // Allow if user is admin or is the group owner
  const isAdmin = currentUser.value.role_name === 'admin'
  const isOwner = currentUser.value.id === groupOwner.value.id_user_owner

  canRemoveUsers.value = isAdmin || isOwner
  console.log(
    'Remove permissions - isAdmin:',
    isAdmin,
    'isOwner:',
    isOwner,
    'canRemove:',
    canRemoveUsers.value,
  )
  console.log(
    'Current user ID:',
    currentUser.value.id,
    'Group owner ID:',
    groupOwner.value.id_user_owner,
  )
  console.log('Current user role:', currentUser.value.role_name)
}

const fetchChatUsers = async () => {
  try {
    console.log('Fetching chat users for group:', props.groupId)

    const result = await callApi({
      query: `
        SELECT 
          cu.id,
          cu.id_user,
          cu.id_chat_group,
          cu.is_active,
          u.username,
          u.email,
          r.role_name
        FROM chat_users cu
        LEFT JOIN users u ON cu.id_user = u.id
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE cu.id_chat_group = ? AND cu.is_active = 1
        ORDER BY u.username ASC
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('API result:', result)

    if (result.success && result.data) {
      chatUsers.value = result.data
      console.log('Chat users loaded:', chatUsers.value)
    } else {
      console.log('No data returned or API call failed:', result)
    }
  } catch (err) {
    console.error('Error fetching chat users:', err)
  }
}

const removeUserFromGroup = async (userId, username) => {
  if (!canRemoveUsers.value) {
    alert('You do not have permission to remove users from this group')
    return
  }

  if (!confirm(`Are you sure you want to remove "${username}" from this group?`)) {
    return
  }

  try {
    const result = await callApi({
      query: `
        UPDATE chat_users 
        SET is_active = 0 
        WHERE id_user = ? AND id_chat_group = ?
      `,
      params: [userId, props.groupId],
      requiresAuth: true,
    })

    console.log('Remove user result:', result)

    if (result.success) {
      // Refresh the users list
      await fetchChatUsers()
      alert(`"${username}" has been removed from the group`)
    } else {
      alert('Failed to remove user from group')
    }
  } catch (err) {
    console.error('Error removing user from group:', err)
    alert('Error removing user from group')
  }
}

const closePopup = () => {
  emit('close')
}

// Close popup when clicking outside
const handleBackdropClick = (event) => {
  if (event.target === event.currentTarget) {
    closePopup()
  }
}

const showAddUser = () => {
  showAddUserModal.value = true
}

const closeAddUserModal = () => {
  showAddUserModal.value = false
}

const handleUserAdded = () => {
  // Refresh the users list after adding a new user
  fetchChatUsers()
}

// Now define the watcher after all functions are defined
watch(
  [() => props.groupId, () => props.isVisible],
  async ([groupId, isVisible]) => {
    console.log('Watch triggered - groupId:', groupId, 'isVisible:', isVisible)
    if (groupId && isVisible) {
      console.log('Conditions met, fetching users...')
      await fetchCurrentUser()
      await fetchGroupInfo()
      await fetchChatUsers()
    } else {
      console.log('Conditions not met - groupId:', groupId, 'isVisible:', isVisible)
    }
  },
  { immediate: true },
)
</script>

<template>
  <div v-if="isVisible" class="popup-overlay" @click="handleBackdropClick">
    <div class="popup-content">
      <div class="popup-header">
        <h3><i class="fas fa-users"></i> Users in "{{ groupName }}"</h3>
        <div class="header-actions">
          <button @click="showAddUser" class="add-user-btn" title="Add User">
            <i class="fas fa-user-plus"></i>
          </button>
          <button @click="closePopup" class="close-btn">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>

      <div class="popup-body">
        <div v-if="loading" class="loading">
          <i class="fas fa-spinner fa-spin"></i> Loading users...
        </div>

        <div v-else-if="error" class="error">
          <i class="fas fa-exclamation-triangle"></i> {{ error }}
        </div>

        <div v-else-if="chatUsers.length === 0" class="no-users">
          <i class="fas fa-user-slash"></i>
          <p>No users in this group</p>
          <button @click="showAddUser" class="btn btn-primary">
            <i class="fas fa-user-plus"></i> Add First User
          </button>
        </div>

        <div v-else class="users-table">
          <table>
            <thead>
              <tr>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th v-if="canRemoveUsers" class="actions-header">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in chatUsers" :key="user.id" class="user-row">
                <td>
                  <i class="fas fa-user-circle"></i>
                  {{ user.username }}
                </td>
                <td>{{ user.email }}</td>
                <td>
                  <span class="role-badge">{{ user.role_name }}</span>
                </td>
                <td>
                  <span class="status-badge active"> <i class="fas fa-circle"></i> Active </span>
                </td>
                <td v-if="canRemoveUsers" class="actions-cell">
                  <button
                    @click="removeUserFromGroup(user.id_user, user.username)"
                    class="remove-btn"
                    :title="`Remove ${user.username} from group`"
                    :disabled="user.id_user === currentUser?.id"
                  >
                    <i class="fas fa-user-minus"></i>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="popup-footer">
        <button @click="showAddUser" class="btn btn-primary">
          <i class="fas fa-user-plus"></i> Add User
        </button>
        <button @click="closePopup" class="btn btn-secondary">Close</button>
      </div>
    </div>

    <!-- Add User Modal -->
    <AddUserToGroup
      :group-id="groupId"
      :group-name="groupName"
      :is-visible="showAddUserModal"
      @close="closeAddUserModal"
      @user-added="handleUserAdded"
    />
  </div>
</template>

<style scoped>
.popup-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.popup-content {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.popup-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #06b6d4;
  color: white;
  border-radius: 8px 8px 0 0;
}

.popup-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.add-user-btn {
  background-color: rgba(255, 255, 255, 0.2);
  border: none;
  color: white;
  font-size: 1rem;
  cursor: pointer;
  padding: 6px 10px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.add-user-btn:hover {
  background-color: rgba(255, 255, 255, 0.3);
}

.close-btn {
  background: none;
  border: none;
  color: white;
  font-size: 1.2rem;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-btn:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.popup-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.users-table {
  overflow-x: auto;
}

.users-table table {
  width: 100%;
  border-collapse: collapse;
}

.users-table th {
  background-color: #f1f5f9;
  padding: 12px 16px;
  text-align: left;
  font-weight: 600;
  color: #475569;
  border-bottom: 1px solid #e2e8f0;
}

.users-table td {
  padding: 12px 16px;
  border-bottom: 1px solid #f1f5f9;
  color: #374151;
}

.user-row:hover {
  background-color: #f8fafc;
}

.user-row td:first-child {
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-badge {
  background-color: #e0f2fe;
  color: #0369a1;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.active {
  color: #059669;
}

.status-badge.active i {
  font-size: 0.6rem;
}

.actions-header {
  width: 100px;
  text-align: center;
}

.actions-cell {
  text-align: center;
}

.remove-btn {
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: all 0.2s;
}

.remove-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.remove-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
  opacity: 0.5;
}

.loading,
.error,
.no-users {
  text-align: center;
  padding: 40px 20px;
  color: #64748b;
}

.error {
  color: #ef4444;
}

.no-users i {
  font-size: 3rem;
  color: #06b6d4;
  margin-bottom: 16px;
  display: block;
}

.popup-footer {
  padding: 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  gap: 10px;
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

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover {
  background-color: #4b5563;
}

/* Responsive design */
@media (max-width: 768px) {
  .popup-content {
    width: 95%;
    margin: 10px;
  }

  .users-table {
    font-size: 0.9rem;
  }

  .users-table th,
  .users-table td {
    padding: 8px 12px;
  }

  .popup-footer {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
