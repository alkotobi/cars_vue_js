<script setup>
import { ref, watch } from 'vue'
import { useApi } from '../../composables/useApi'

const props = defineProps({
  isVisible: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['close', 'group-added'])

const { callApi, loading, error } = useApi()
const availableUsers = ref([])
const selectedUsers = ref(new Set()) // Use Set to track selected user IDs
const groupName = ref('')
const groupDescription = ref('')
const currentUser = ref(null) // Store current user info

// Watch for changes in isVisible to fetch users
watch(
  [() => props.isVisible],
  async ([isVisible]) => {
    console.log('AddGroup watch triggered - isVisible:', isVisible)
    if (isVisible) {
      console.log('Modal is visible, fetching available users...')
      await getCurrentUser()
      await fetchAvailableUsers()
    }
  },
  { immediate: true },
)

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      currentUser.value = JSON.parse(userStr)
      console.log('Current user loaded:', currentUser.value)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
  }
}

const fetchAvailableUsers = async () => {
  try {
    console.log('Fetching all available users for new group')

    const result = await callApi({
      query: `
        SELECT 
          u.id,
          u.username,
          u.email,
          r.role_name
        FROM users u
        LEFT JOIN roles r ON u.role_id = r.id
        ORDER BY u.username ASC
      `,
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
  // Prevent selecting the current user (owner)
  if (userId === currentUser.value?.id) {
    return
  }

  if (selectedUsers.value.has(userId)) {
    selectedUsers.value.delete(userId)
  } else {
    selectedUsers.value.add(userId)
  }
}

const isUserSelected = (userId) => {
  return selectedUsers.value.has(userId)
}

const isCurrentUser = (userId) => {
  return userId === currentUser.value?.id
}

const getSelectedUsersCount = () => {
  return selectedUsers.value.size
}

const getSelectedUsersInfo = () => {
  return availableUsers.value.filter((user) => selectedUsers.value.has(user.id))
}

const getTotalGroupMembers = () => {
  // Owner + selected users
  return 1 + selectedUsers.value.size
}

const createGroup = async () => {
  if (!groupName.value.trim()) {
    alert('Please enter a group name')
    return
  }

  if (selectedUsers.value.size === 0) {
    alert('Please select at least one user for the group')
    return
  }

  try {
    // Get current user to set as group owner
    const userStr = localStorage.getItem('user')
    if (!userStr) {
      alert('User session not found. Please log in again.')
      return
    }

    const currentUser = JSON.parse(userStr)
    const selectedUserIds = Array.from(selectedUsers.value)

    // Remove the current user from selected users to prevent duplication
    // The current user will be added automatically as the owner
    const filteredUserIds = selectedUserIds.filter((userId) => userId !== currentUser.id)

    // Create the final list: current user (owner) + selected users
    const finalUserIds = [currentUser.id, ...filteredUserIds]

    console.log('Current user (owner):', currentUser.id)
    console.log('Selected users (excluding owner):', filteredUserIds)
    console.log('Final user list:', finalUserIds)

    // Create the chat group first
    const groupResult = await callApi({
      query: `
        INSERT INTO chat_groups (name, description, id_user_owner, is_active) 
        VALUES (?, ?, ?, 1)
      `,
      params: [groupName.value.trim(), groupDescription.value.trim() || null, currentUser.id],
      requiresAuth: true,
    })

    console.log('Create group result:', groupResult)

    if (groupResult.success && groupResult.lastInsertId) {
      const groupId = groupResult.lastInsertId

      // Add all users to the group (owner + selected users)
      const insertQueries = finalUserIds
        .map(() => 'INSERT INTO chat_users (id_user, id_chat_group, is_active) VALUES (?, ?, 1)')
        .join('; ')

      const params = finalUserIds.flatMap((userId) => [userId, groupId])

      const usersResult = await callApi({
        query: insertQueries,
        params: params,
        requiresAuth: true,
      })

      console.log('Add users to group result:', usersResult)

      if (usersResult.success) {
        emit('group-added')
        closeModal()
        const memberCount = finalUserIds.length
        const additionalMembers = filteredUserIds.length
        alert(
          `Group "${groupName.value}" created successfully with ${memberCount} members (you + ${additionalMembers} additional users)`,
        )
      } else {
        alert('Group created but failed to add users')
      }
    } else {
      alert('Failed to create group')
    }
  } catch (err) {
    console.error('Error creating group:', err)
    alert('Error creating group')
  }
}

const closeModal = () => {
  groupName.value = ''
  groupDescription.value = ''
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
        <h3><i class="fas fa-plus-circle"></i> Create New Chat Group</h3>
        <button @click="closeModal" class="close-btn">
          <i class="fas fa-times"></i>
        </button>
      </div>

      <div class="modal-body">
        <div v-if="loading" class="loading">
          <i class="fas fa-spinner fa-spin"></i> Loading users...
        </div>

        <div v-else-if="error" class="error">
          <i class="fas fa-exclamation-triangle"></i> {{ error }}
        </div>

        <div v-else class="group-creation">
          <!-- Group Details Section -->
          <div class="group-details">
            <h4><i class="fas fa-info-circle"></i> Group Information</h4>

            <div class="form-group">
              <label for="group-name" class="form-label">Group Name *</label>
              <input
                id="group-name"
                v-model="groupName"
                type="text"
                placeholder="Enter group name"
                class="form-input"
                maxlength="255"
              />
            </div>

            <div class="form-group">
              <label for="group-description" class="form-label">Description (Optional)</label>
              <textarea
                id="group-description"
                v-model="groupDescription"
                placeholder="Enter group description"
                class="form-textarea"
                rows="3"
                maxlength="500"
              ></textarea>
            </div>
          </div>

          <!-- User Selection Section -->
          <div class="user-selection">
            <div class="selection-header">
              <h4><i class="fas fa-users"></i> Select Group Members *</h4>
              <p class="selection-info">
                <i class="fas fa-info-circle"></i>
                You will be automatically added as the group owner. Select at least one additional
                member.
              </p>
              <p class="selected-count">
                <i class="fas fa-check-circle"></i>
                {{ getSelectedUsersCount() }} additional user{{
                  getSelectedUsersCount() !== 1 ? 's' : ''
                }}
                selected ({{ getTotalGroupMembers() }} total members including you)
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
                          selectedUsers.size ===
                            availableUsers.filter((u) => !isCurrentUser(u.id)).length &&
                          availableUsers.filter((u) => !isCurrentUser(u.id)).length > 0
                        "
                        :indeterminate="
                          selectedUsers.size > 0 &&
                          selectedUsers.size <
                            availableUsers.filter((u) => !isCurrentUser(u.id)).length
                        "
                        @change="
                          (e) => {
                            const nonOwnerUsers = availableUsers.filter((u) => !isCurrentUser(u.id))
                            if (e.target.checked) {
                              nonOwnerUsers.forEach((user) => selectedUsers.add(user.id))
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
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="user in availableUsers"
                    :key="user.id"
                    class="user-row"
                    :class="{
                      selected: isUserSelected(user.id),
                      owner: isCurrentUser(user.id),
                      disabled: isCurrentUser(user.id),
                    }"
                    @click="toggleUserSelection(user.id)"
                  >
                    <td class="checkbox-cell">
                      <input
                        type="checkbox"
                        :checked="isUserSelected(user.id)"
                        :disabled="isCurrentUser(user.id)"
                        @click.stop
                        @change="toggleUserSelection(user.id)"
                        class="user-checkbox"
                      />
                    </td>
                    <td>
                      <i class="fas fa-user-circle"></i>
                      {{ user.username }}
                      <span v-if="isCurrentUser(user.id)" class="owner-badge">
                        <i class="fas fa-crown"></i> You (Owner)
                      </span>
                    </td>
                    <td>{{ user.email }}</td>
                    <td>
                      <span class="role-badge">{{ user.role_name }}</span>
                    </td>
                    <td>
                      <span v-if="isCurrentUser(user.id)" class="status-badge owner-status">
                        <i class="fas fa-crown"></i> Owner
                      </span>
                      <span
                        v-else-if="isUserSelected(user.id)"
                        class="status-badge selected-status"
                      >
                        <i class="fas fa-check"></i> Selected
                      </span>
                      <span v-else class="status-badge available-status">
                        <i class="fas fa-user"></i> Available
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div v-if="getSelectedUsersCount() > 0" class="selected-users-summary">
              <h4>Group Members:</h4>
              <div class="selected-users-list">
                <!-- Show owner first -->
                <span v-if="currentUser" class="selected-user-tag owner-tag">
                  <i class="fas fa-crown"></i>
                  {{ currentUser.username }} (Owner)
                </span>
                <!-- Show selected users -->
                <span
                  v-for="user in getSelectedUsersInfo()"
                  :key="user.id"
                  class="selected-user-tag"
                >
                  <i class="fas fa-user"></i>
                  {{ user.username }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="modal-footer">
        <button @click="closeModal" class="btn btn-secondary">Cancel</button>
        <button
          @click="createGroup"
          class="btn btn-primary"
          :disabled="!groupName.trim() || getSelectedUsersCount() === 0 || loading"
        >
          <i class="fas fa-plus"></i>
          Create Group
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
  max-width: 800px;
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

.group-details {
  margin-bottom: 30px;
  padding: 20px;
  background-color: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.group-details h4 {
  margin: 0 0 15px 0;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group {
  margin-bottom: 15px;
}

.form-label {
  display: block;
  margin-bottom: 5px;
  font-weight: 600;
  color: #374151;
}

.form-input,
.form-textarea {
  width: 100%;
  padding: 10px;
  border: 1px solid #d1d5db;
  border-radius: 6px;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #06b6d4;
  box-shadow: 0 0 0 3px rgba(6, 182, 212, 0.1);
}

.form-textarea {
  resize: vertical;
  min-height: 80px;
}

.user-selection {
  margin-top: 20px;
}

.selection-header {
  margin-bottom: 20px;
}

.selection-header h4 {
  margin: 0 0 10px 0;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.selection-info {
  margin: 5px 0;
  color: #6b7280;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 6px;
}

.selected-count {
  margin: 5px 0;
  color: #059669;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 6px;
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

.user-row.owner {
  background-color: #fef3c7;
  border-left: 3px solid #f59e0b;
}

.user-row.disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.user-row.disabled:hover {
  background-color: #fef3c7;
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
.error {
  text-align: center;
  padding: 40px 20px;
  color: #64748b;
}

.error {
  color: #ef4444;
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

.owner-badge {
  background-color: #f59e0b;
  color: white;
  padding: 2px 6px;
  border-radius: 8px;
  font-size: 0.7rem;
  margin-left: 8px;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.8rem;
  font-weight: 500;
  padding: 4px 8px;
  border-radius: 12px;
}

.owner-status {
  background-color: #fef3c7;
  color: #92400e;
}

.selected-status {
  background-color: #dcfce7;
  color: #166534;
}

.available-status {
  background-color: #f1f5f9;
  color: #64748b;
}

.owner-tag {
  background-color: #fef3c7;
  color: #92400e;
  border: 1px solid #f59e0b;
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
