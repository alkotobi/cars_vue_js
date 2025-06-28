<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../../composables/useApi'

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

const emit = defineEmits(['close', 'user-added'])

const { callApi, loading, error } = useApi()
const availableUsers = ref([])
const selectedUsers = ref(new Set()) // Use Set to track selected user IDs

// Watch for changes in groupId and isVisible to fetch users
watch(
  [() => props.groupId, () => props.isVisible],
  async ([groupId, isVisible]) => {
    console.log('AddUserToGroup watch triggered - groupId:', groupId, 'isVisible:', isVisible)
    if (groupId && isVisible) {
      console.log('Conditions met, fetching available users...')
      await fetchAvailableUsers()
    } else {
      console.log('Conditions not met - groupId:', groupId, 'isVisible:', isVisible)
    }
  },
  { immediate: true },
)

const fetchAvailableUsers = async () => {
  try {
    console.log('Fetching available users for group:', props.groupId)

    // Get all users who are not already in this chat group
    const result = await callApi({
      query: `
        SELECT 
          u.id,
          u.username,
          u.email,
          r.role_name
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        WHERE u.id NOT IN (
          SELECT cu.id_user 
          FROM chat_users cu 
          WHERE cu.id_chat_group = ? AND cu.is_active = 1
        )
        ORDER BY u.username ASC
      `,
      params: [props.groupId],
      requiresAuth: true,
    })

    console.log('Available users result:', result)

    if (result.success && result.data) {
      availableUsers.value = result.data
      console.log('Available users loaded:', availableUsers.value)
    } else {
      console.log('No data returned or API call failed:', result)
    }
  } catch (err) {
    console.error('Error fetching available users:', err)
  }
}

const toggleUserSelection = (userId) => {
  if (selectedUsers.value.has(userId)) {
    selectedUsers.value.delete(userId)
  } else {
    selectedUsers.value.add(userId)
  }
}

const isUserSelected = (userId) => {
  return selectedUsers.value.has(userId)
}

const getSelectedUsersCount = () => {
  return selectedUsers.value.size
}

const getSelectedUsersInfo = () => {
  return availableUsers.value.filter((user) => selectedUsers.value.has(user.id))
}

const addUsersToGroup = async () => {
  if (selectedUsers.value.size === 0) {
    alert('Please select at least one user to add')
    return
  }

  try {
    // Add all selected users to the group
    const selectedUserIds = Array.from(selectedUsers.value)

    // Create multiple INSERT statements
    const insertQueries = selectedUserIds
      .map(() => 'INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)')
      .join('; ')

    const params = selectedUserIds.flatMap((userId) => [userId, props.groupId])

    const result = await callApi({
      query: insertQueries,
      params: params,
      requiresAuth: true,
    })

    console.log('Add users result:', result)

    if (result.success) {
      emit('user-added')
      closeModal()
    } else {
      alert('Failed to add users to group')
    }
  } catch (err) {
    console.error('Error adding users to group:', err)
    alert('Error adding users to group')
  }
}

const closeModal = () => {
  selectedUsers.value.clear()
  emit('close')
}

const handleBackdropClick = (event) => {
  if (event.target === event.currentTarget) {
    closeModal()
  }
}
</script>

<template>
  <div v-if="isVisible" class="modal-overlay" @click="handleBackdropClick">
    <div class="modal-content">
      <div class="modal-header">
        <h3><i class="fas fa-user-plus"></i> Add Users to "{{ groupName }}"</h3>
        <button @click="closeModal" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading">
          <i class="fas fa-spinner fa-spin"></i> Loading available users...
        </div>

        <div v-else-if="error" class="error">
          <i class="fas fa-exclamation-triangle"></i> {{ error }}
        </div>

        <div v-else-if="availableUsers.length === 0" class="no-users">
          <i class="fas fa-users-slash"></i>
          <p>No available users to add</p>
          <p class="sub-text">All users are already in this group</p>
        </div>

        <div v-else class="users-selection">
          <div class="selection-info">
            <p>
              <strong>Select users to add to "{{ groupName }}":</strong>
            </p>
            <p class="selected-count">
              <i class="fas fa-check-circle"></i>
              {{ getSelectedUsersCount() }} user{{
                getSelectedUsersCount() !== 1 ? 's' : ''
              }}
              selected
            </p>
          </div>

          <div class="users-table-container">
            <table class="users-table">
              <thead>
                <tr>
                  <th class="checkbox-header">
                    <input
                      type="checkbox"
                      :checked="
                        selectedUsers.size === availableUsers.length && availableUsers.length > 0
                      "
                      :indeterminate="
                        selectedUsers.size > 0 && selectedUsers.size < availableUsers.length
                      "
                      @change="
                        (e) => {
                          if (e.target.checked) {
                            availableUsers.forEach((user) => selectedUsers.add(user.id))
                          } else {
                            selectedUsers.clear()
                          }
                        }
                      "
                      class="select-all-checkbox"
                    />
                  </th>
                  <th>Username</th>
                  <th>Email</th>
                  <th>Role</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="user in availableUsers"
                  :key="user.id"
                  class="user-row"
                  :class="{ selected: isUserSelected(user.id) }"
                  @click="toggleUserSelection(user.id)"
                >
                  <td class="checkbox-cell">
                    <input
                      type="checkbox"
                      :checked="isUserSelected(user.id)"
                      @click.stop
                      @change="toggleUserSelection(user.id)"
                      class="user-checkbox"
                    />
                  </td>
                  <td>
                    <i class="fas fa-user-circle"></i>
                    {{ user.username }}
                  </td>
                  <td>{{ user.email }}</td>
                  <td>
                    <span class="role-badge">{{ user.role_name }}</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="getSelectedUsersCount() > 0" class="selected-users-summary">
            <h4>Selected Users:</h4>
            <div class="selected-users-list">
              <span v-for="user in getSelectedUsersInfo()" :key="user.id" class="selected-user-tag">
                <i class="fas fa-user"></i>
                {{ user.username }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="closeModal" class="btn btn-secondary">Cancel</button>
        <button
          @click="addUsersToGroup"
          class="btn btn-primary"
          :disabled="getSelectedUsersCount() === 0 || loading"
        >
          <i class="fas fa-plus"></i>
          Add {{ getSelectedUsersCount() }} User{{ getSelectedUsersCount() !== 1 ? 's' : '' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1001;
}

.modal-content {
  background-color: white;
  border-radius: 8px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  width: 90%;
  max-width: 700px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  padding: 20px;
  border-bottom: 1px solid #e2e8f0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background-color: #06b6d4;
  color: white;
  border-radius: 8px 8px 0 0;
}

.modal-header h3 {
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
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

.modal-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.selection-info {
  margin-bottom: 20px;
  padding: 15px;
  background-color: #f0f9ff;
  border: 1px solid #e0f2fe;
  border-radius: 6px;
}

.selection-info p {
  margin: 5px 0;
  color: #0c4a6e;
}

.selected-count {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #059669;
  font-weight: 500;
}

.users-table-container {
  overflow-x: auto;
  margin-bottom: 20px;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
  background-color: white;
  border-radius: 6px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
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

.user-row {
  cursor: pointer;
  transition: background-color 0.2s;
}

.user-row:hover {
  background-color: #f8fafc;
}

.user-row.selected {
  background-color: #e0f2fe;
}

.checkbox-header,
.checkbox-cell {
  width: 50px;
  text-align: center;
}

.select-all-checkbox,
.user-checkbox {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.user-row td:first-child {
  text-align: center;
}

.user-row td:nth-child(2) {
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

.selected-users-summary {
  margin-top: 20px;
  padding: 15px;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
}

.selected-users-summary h4 {
  margin: 0 0 10px 0;
  color: #166534;
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-users-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.selected-user-tag {
  background-color: #dcfce7;
  color: #166534;
  padding: 4px 8px;
  border-radius: 12px;
  font-size: 0.8rem;
  display: flex;
  align-items: center;
  gap: 4px;
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

.sub-text {
  font-size: 0.9rem;
  color: #9ca3af;
  margin-top: 8px;
}

.modal-footer {
  padding: 20px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: flex-end;
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

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: #6b7280;
  color: white;
}

.btn-secondary:hover:not(:disabled) {
  background-color: #4b5563;
}

.btn-primary {
  background-color: #06b6d4;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background-color: #0891b2;
}

/* Responsive design */
@media (max-width: 768px) {
  .modal-content {
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

  .modal-footer {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }

  .selected-users-list {
    flex-direction: column;
  }
}
</style>
