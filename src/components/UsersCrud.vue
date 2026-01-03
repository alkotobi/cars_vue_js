<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error, loading } = useApi()
const users = ref([])
const roles = ref([])
const editingUser = ref(null)
const editUserData = ref({
  username: '',
  email: '',
  password: '',
  role_id: '',
  max_unpayed_created_bills: 0,
})

const fetchUsers = async () => {
  const result = await callApi({
    query: `
      SELECT users.*, roles.role_name 
      FROM users 
      LEFT JOIN roles ON users.role_id = roles.id
    `,
  })
  if (result.success) {
    users.value = result.data
  }
}

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles',
  })
  if (result.success) {
    roles.value = result.data
  }
}

// Expose fetchUsers so parent can refresh the list
defineExpose({
  fetchUsers,
})

const startEditUser = (user) => {
  editingUser.value = user.id
  editUserData.value = {
    username: user.username,
    email: user.email,
    password: '', // Don't populate password for security
    role_id: user.role_id,
    max_unpayed_created_bills: user.max_unpayed_created_bills || 0,
  }
}

const updateUser = async (user) => {
  // Validate required fields
  if (!editUserData.value.username || !editUserData.value.email) {
    error.value = 'Username and email are required'
    return
  }

  try {
    // First update user info without password
    let result = await callApi({
      query:
        'UPDATE users SET username = ?, email = ?, role_id = ?, max_unpayed_created_bills = ? WHERE id = ?',
      params: [
        editUserData.value.username,
        editUserData.value.email,
        editUserData.value.role_id,
        editUserData.value.max_unpayed_created_bills,
        user.id,
      ],
    })

    // If there's a password to update, do it in a separate call
    if (editUserData.value.password && editUserData.value.password.trim()) {
      result = await callApi({
        query: 'UPDATE users SET password = ? WHERE id = ?',
        params: [editUserData.value.password, user.id],
        action: 'hash_password',
      })
    }

    if (result.success) {
      await fetchUsers()
      cancelEdit()
    } else {
      error.value = result.error || 'Failed to update user'
    }
  } catch (err) {
    error.value = err.message || 'An error occurred while updating user'
    console.error('Update error:', err)
  }
}

const cancelEdit = () => {
  editingUser.value = null
  editUserData.value = {
    username: '',
    email: '',
    password: '',
    role_id: '',
    max_unpayed_created_bills: 0,
  }
}

const deleteUser = async (user) => {
  if (!confirm(`Are you sure you want to delete user "${user.username}"?`)) {
    return
  }

  const result = await callApi({
    query: 'DELETE FROM users WHERE id = ?',
    params: [user.id],
  })
  if (result.success) {
    fetchUsers()
  }
}

onMounted(() => {
  fetchUsers()
  fetchRoles()
})
</script>

<template>
  <div class="users-container">
    <div class="content-container">
      <!-- Loading Overlay -->
      <div v-if="loading" class="loading-overlay">
        <i class="fas fa-spinner fa-spin fa-2x"></i>
        <span>Processing...</span>
      </div>

      <table class="users-table">
        <thead>
          <tr>
            <th><i class="fas fa-user"></i> Username</th>
            <th><i class="fas fa-envelope"></i> Email</th>
            <th><i class="fas fa-shield-alt"></i> Role</th>
            <th><i class="fas fa-receipt"></i> Max Unpaid Bills</th>
            <th><i class="fas fa-cogs"></i> Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id" :class="{ editing: editingUser === user.id }">
            <template v-if="editingUser === user.id">
              <td>
                <div class="input-wrapper">
                  <i class="fas fa-user input-icon"></i>
                  <input
                    v-model="editUserData.username"
                    placeholder="Username"
                    class="input-field with-icon"
                    :disabled="loading"
                  />
                </div>
              </td>
              <td>
                <div class="input-wrapper">
                  <i class="fas fa-envelope input-icon"></i>
                  <input
                    v-model="editUserData.email"
                    placeholder="Email"
                    class="input-field with-icon"
                    :disabled="loading"
                  />
                </div>
              </td>
              <td>
                <div class="input-wrapper">
                  <i class="fas fa-shield-alt input-icon"></i>
                  <select
                    v-model="editUserData.role_id"
                    class="input-field with-icon"
                    :disabled="loading"
                  >
                    <option disabled value="">Select Role</option>
                    <option v-for="role in roles" :key="role.id" :value="role.id">
                      {{ role.role_name }}
                    </option>
                  </select>
                </div>
              </td>
              <td>
                <div class="input-wrapper">
                  <i class="fas fa-receipt input-icon"></i>
                  <input
                    v-model.number="editUserData.max_unpayed_created_bills"
                    type="number"
                    min="0"
                    placeholder="Max Unpaid Bills"
                    class="input-field with-icon"
                    :disabled="loading"
                  />
                </div>
              </td>
              <td class="actions-cell">
                <div class="input-wrapper">
                  <i class="fas fa-key input-icon"></i>
                  <input
                    v-model="editUserData.password"
                    type="password"
                    placeholder="New Password (optional)"
                    class="input-field with-icon"
                    :disabled="loading"
                  />
                </div>
                <div class="button-group">
                  <button @click="updateUser(user)" :disabled="loading" class="btn save-btn">
                    <i class="fas" :class="loading ? 'fa-spinner fa-spin' : 'fa-save'"></i>
                    {{ loading ? 'Saving...' : 'Save' }}
                  </button>
                  <button @click="cancelEdit" class="btn cancel-btn" :disabled="loading">
                    <i class="fas fa-times"></i>
                    Cancel
                  </button>
                </div>
              </td>
            </template>
            <template v-else>
              <td>
                <i class="fas fa-user text-gray"></i>
                {{ user.username }}
              </td>
              <td>
                <i class="fas fa-envelope text-gray"></i>
                {{ user.email }}
              </td>
              <td>
                <i class="fas fa-shield-alt text-gray"></i>
                {{ user.role_name || 'No Role' }}
              </td>
              <td>
                <i class="fas fa-receipt text-gray"></i>
                {{ user.max_unpayed_created_bills || 0 }}
              </td>
              <td class="actions-cell">
                <div class="button-group">
                  <button
                    @click="startEditUser(user)"
                    class="btn edit-btn"
                    :disabled="loading || editingUser !== null"
                  >
                    <i class="fas fa-edit"></i>
                    Edit
                  </button>
                  <button
                    @click="deleteUser(user)"
                    :disabled="loading || editingUser !== null"
                    class="btn delete-btn"
                  >
                    <i class="fas fa-trash-alt"></i>
                    Delete
                  </button>
                </div>
              </td>
            </template>
          </tr>
        </tbody>
      </table>

      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>
    </div>
  </div>
</template>

<style scoped>
.users-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 24px;
  position: relative;
}

.content-container {
  border-radius: 6px;
  border: 1px solid #eaeaea;
  padding: 16px;
  position: relative;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  gap: 12px;
  z-index: 10;
  backdrop-filter: blur(2px);
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
}

.users-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  margin-bottom: 20px;
}

.users-table th,
.users-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #eaeaea;
}

.users-table th {
  background-color: #f8f9fa;
  font-weight: 600;
  color: #2c3e50;
  position: sticky;
  top: 0;
  z-index: 1;
}

.users-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.users-table tr {
  transition: all 0.2s ease;
}

.users-table tr:hover:not(.editing) {
  background-color: #f8fafc;
}

.users-table tr.editing {
  background-color: #f0f9ff;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.actions-cell {
  min-width: 200px;
}

.button-group {
  display: flex;
  gap: 8px;
}

.input-wrapper {
  position: relative;
  margin-bottom: 8px;
}

.input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
  pointer-events: none;
}

.input-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  transition: all 0.2s ease;
}

.input-field.with-icon {
  padding-left: 36px;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.input-field:disabled {
  background-color: #f9fafb;
  cursor: not-allowed;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn i {
  font-size: 14px;
}

.edit-btn {
  background: #3b82f6;
  color: white;
}

.edit-btn:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.delete-btn {
  background: #ef4444;
  color: white;
}

.delete-btn:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-1px);
}

.save-btn {
  background: #10b981;
  color: white;
}

.save-btn:hover:not(:disabled) {
  background: #059669;
  transform: translateY(-1px);
}

.cancel-btn {
  background: #6b7280;
  color: white;
}

.cancel-btn:hover:not(:disabled) {
  background: #4b5563;
  transform: translateY(-1px);
}

.add-form {
  margin-top: 20px;
  padding: 20px;
  border-top: 1px solid #eaeaea;
  background: #f8fafc;
  border-radius: 6px;
}

.add-form h3 {
  margin-bottom: 16px;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
}

.add-form h3 i {
  color: #3b82f6;
}

.add-btn {
  background: #3b82f6;
  color: white;
  width: 100%;
  justify-content: center;
  padding: 12px;
  margin-top: 8px;
}

.add-btn:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 1rem;
  border-radius: 8px;
  margin: 1rem 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  animation: slideIn 0.3s ease;
}

.text-gray {
  color: #6b7280;
  margin-right: 8px;
}

@keyframes slideIn {
  from {
    transform: translateY(-10px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

@media (max-width: 768px) {
  .users-container {
    padding: 16px;
  }

  .actions-cell {
    min-width: auto;
  }

  .button-group {
    flex-direction: column;
  }

  .btn {
    width: 100%;
    justify-content: center;
  }
}
</style>
