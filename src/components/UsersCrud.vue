<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error, loading } = useApi()
const users = ref([])
const roles = ref([])
const newUser = ref({ username: '', email: '', password: '', role_id: '' })
const editingUser = ref(null)
const editUserData = ref({ username: '', email: '', password: '', role_id: '' })

const fetchUsers = async () => {
  const result = await callApi({
    query: `
      SELECT users.*, roles.role_name 
      FROM users 
      LEFT JOIN roles ON users.role_id = roles.id
    `
  })
  if (result.success) {
    users.value = result.data
  }
}

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles'
  })
  if (result.success) {
    roles.value = result.data
  }
}

const addUser = async () => {
  if (!newUser.value.username || !newUser.value.password || !newUser.value.email) return
  
  const result = await callApi({
    query: 'INSERT INTO users (username, email, password, role_id) VALUES (?, ?, ?, ?)',
    params: [newUser.value.username, newUser.value.email, newUser.value.password, newUser.value.role_id],
    action: 'insert_user'  // Tell API to hash the password before inserting
  })
  if (result.success) {
    fetchUsers()
    newUser.value = { username: '', email: '', password: '', role_id: '' }
  } else {
    error.value = result.error
    console.error(result.error)
  }
}

const startEditUser = (user) => {
  editingUser.value = user.id
  editUserData.value = { 
    username: user.username,
    email: user.email, 
    password: '', // Don't populate password for security
    role_id: user.role_id 
  }
}

const updateUser = async (user) => {
  let query = 'UPDATE users SET username = ?, email = ?, role_id = ?'
  let params = [editUserData.value.username, editUserData.value.email, editUserData.value.role_id]
  
  if (editUserData.value.password) {
    query += ', password = ?'
    params.push(editUserData.value.password)
  }
  
  query += ' WHERE id = ?'
  params.push(user.id)

  const result = await callApi({ 
    query, 
    params,
    action: editUserData.value.password ? 'hash_password' : undefined 
  })
  
  if (result.success) {
    fetchUsers()
    cancelEdit()
  }
}

const cancelEdit = () => {
  editingUser.value = null
  editUserData.value = { username: '', password: '', role_id: '' }
}

const deleteUser = async (user) => {
  if (!confirm(`Are you sure you want to delete user "${user.username}"?`)) {
    return
  }

  const result = await callApi({
    query: 'DELETE FROM users WHERE id = ?',
    params: [user.id]
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
    <h2 class="title">Users</h2>
    <div class="content-container">
      <table class="users-table">
        <thead>
          <tr>
            <th>Username</th>
            <th>Role</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in users" :key="user.id">
            <template v-if="editingUser === user.id">
              <td>
                <input v-model="editUserData.username" placeholder="Username" class="input-field" />
              </td>
              <td>
                <input v-model="editUserData.email" placeholder="Email" class="input-field" />
              </td>
              <td>
                <select v-model="editUserData.role_id" class="input-field">
                  <option disabled value="">Select Role</option>
                  <option v-for="role in roles" :key="role.id" :value="role.id">
                    {{ role.role_name }}
                  </option>
                </select>
              </td>
              <td class="actions-cell">
                <input 
                  v-model="editUserData.password" 
                  type="password" 
                  placeholder="New Password (optional)" 
                  class="input-field"
                />
                <div class="button-group">
                  <button @click="updateUser(user)" :disabled="loading" class="btn save-btn">Save</button>
                  <button @click="cancelEdit" class="btn cancel-btn">Cancel</button>
                </div>
              </td>
            </template>
            <template v-else>
              <td>{{ user.username }}</td>
              <td>{{ user.email }}</td>
              <td>{{ user.role_name || 'No Role' }}</td>
              <td class="actions-cell">
                <div class="button-group">
                  <button @click="startEditUser(user)" class="btn edit-btn">Edit</button>
                  <button @click="deleteUser(user)" :disabled="loading" class="btn delete-btn">Delete</button>
                </div>
              </td>
            </template>
          </tr>
        </tbody>
      </table>

      <div class="add-form">
        <h3>Add New User</h3>
        <input v-model="newUser.username" placeholder="Username" class="input-field" />
        <input v-model="newUser.email" type="email" placeholder="Email" class="input-field" />
        <input v-model="newUser.password" type="password" placeholder="Password" class="input-field" />
        <select v-model="newUser.role_id" class="input-field">
          <option disabled value="">Select Role</option>
          <option v-for="role in roles" :key="role.id" :value="role.id">
            {{ role.role_name }}
          </option>
        </select>
        <button @click="addUser" :disabled="loading" class="btn add-btn">Add User</button>
      </div>
      <div v-if="error" class="error-message">{{ error }}</div>
    </div>
  </div>
</template>

<style scoped>
.users-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 24px;
}

.title {
  color: #2c3e50;
  font-size: 24px;
  margin-bottom: 20px;
  font-weight: 600;
}

.content-container {
  border-radius: 6px;
  border: 1px solid #eaeaea;
  padding: 16px;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
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
}

.actions-cell {
  min-width: 200px;
}

.button-group {
  display: flex;
  gap: 8px;
}

.input-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 8px;
  font-size: 14px;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  transition: background-color 0.2s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.edit-btn {
  background: #3b82f6;
  color: white;
}

.edit-btn:hover {
  background: #2563eb;
}

.delete-btn {
  background: #ef4444;
  color: white;
}

.delete-btn:hover {
  background: #dc2626;
}

.save-btn {
  background: #10b981;
  color: white;
}

.save-btn:hover {
  background: #059669;
}

.cancel-btn {
  background: #6b7280;
  color: white;
}

.cancel-btn:hover {
  background: #4b5563;
}

.add-form {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eaeaea;
}

.add-form h3 {
  margin-bottom: 16px;
  color: #2c3e50;
}

.add-btn {
  background: #3b82f6;
  color: white;
  width: 100%;
  padding: 8px;
}

.add-btn:hover {
  background: #2563eb;
}

.error-message {
  color: #dc2626;
  background: #fee2e2;
  padding: 12px;
  border-radius: 6px;
  margin-top: 16px;
  font-size: 14px;
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
  }
}
</style>