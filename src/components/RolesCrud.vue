<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error, loading } = useApi()
const roles = ref([])
const newRole = ref({ role_name: '', description: '' })
const editingRole = ref(null)
const editRoleData = ref({ role_name: '', description: '' })

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles'
  })
  if (result.success) {
    roles.value = result.data
  }
}

const addRole = async () => {
  if (!newRole.value.role_name) return
  const result = await callApi({
    query: 'INSERT INTO roles (role_name, description) VALUES (?, ?)',
    params: [newRole.value.role_name, newRole.value.description]
  })
  if (result.success) {
    fetchRoles()
    newRole.value = { role_name: '', description: '' }
  }
}

const startEditRole = (role) => {
  editingRole.value = role.id
  editRoleData.value = { role_name: role.role_name, description: role.description }
}

const cancelEdit = () => {
  editingRole.value = null
  editRoleData.value = { role_name: '', description: '' }
}

const updateRole = async (role) => {
  const result = await callApi({
    query: 'UPDATE roles SET role_name = ?, description = ? WHERE id = ?',
    params: [editRoleData.value.role_name, editRoleData.value.description, role.id]
  })
  if (result.success) {
    fetchRoles()
    cancelEdit()
  }
}

const deleteRole = async (role) => {
  // First confirm deletion
  if (!confirm(`Are you sure you want to delete the role "${role.role_name}"?`)) {
    return
  }

  // Check if role is used by any users
  const checkResult = await callApi({
    query: 'SELECT COUNT(*) as count FROM users WHERE role_id = ?',
    params: [role.id]
  })

  if (checkResult.success && checkResult.data[0].count > 0) {
    error.value = `Cannot delete role "${role.role_name}" because it is assigned to ${checkResult.data[0].count} user(s)`
    return
  }

  const result = await callApi({
    query: 'DELETE FROM roles WHERE id = ?',
    params: [role.id]
  })
  if (result.success) {
    fetchRoles()
  }
}

onMounted(() => {
  fetchRoles()
})
</script>

<template>
  <div class="roles-container">
    <h2 class="title">Roles</h2>
    <div class="content-container">
      <ul class="items-list">
        <li v-for="role in roles" :key="role.id" class="item">
          <template v-if="editingRole === role.id">
            <div class="edit-form">
              <input v-model="editRoleData.role_name" placeholder="Role name" class="input-field" />
              <input v-model="editRoleData.description" placeholder="Description" class="input-field" />
              <div class="button-group">
                <button @click="updateRole(role)" :disabled="loading" class="btn save-btn">Save</button>
                <button @click="cancelEdit" class="btn cancel-btn">Cancel</button>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="item-content">
              <div class="item-text">
                <strong>{{ role.role_name }}</strong>
                <span class="description">{{ role.description }}</span>
              </div>
              <div class="button-group">
                <button @click="startEditRole(role)" class="btn edit-btn">Edit</button>
                <button @click="deleteRole(role)" :disabled="loading" class="btn delete-btn">Delete</button>
              </div>
            </div>
          </template>
        </li>
      </ul>
      <div class="add-form">
        <input v-model="newRole.role_name" placeholder="Role name" class="input-field" />
        <input v-model="newRole.description" placeholder="Description" class="input-field" />
        <button @click="addRole" :disabled="loading" class="btn add-btn">Add Role</button>
      </div>
      <div v-if="error" class="error-message">{{ error }}</div>
    </div>
  </div>
</template>

<style scoped>
.roles-container {
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  padding: 24px;
  min-width: 300px;
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

.items-list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.item {
  margin-bottom: 12px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
  transition: background-color 0.2s;
}

.item:hover {
  background: #edf2f7;
}

.item-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
}

.item-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.description {
  color: #666;
  font-size: 0.9em;
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

.add-btn {
  background: #3b82f6;
  color: white;
  width: 100%;
  padding: 8px;
}

.add-btn:hover {
  background: #2563eb;
}

.add-form {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #eaeaea;
}

.error-message {
  color: #dc2626;
  background: #fee2e2;
  padding: 12px;
  border-radius: 6px;
  margin-top: 16px;
  font-size: 14px;
}

@media (max-width: 640px) {
  .roles-container {
    padding: 16px;
  }

  .item-content {
    flex-direction: column;
    align-items: flex-start;
  }

  .button-group {
    margin-top: 8px;
    width: 100%;
    justify-content: flex-end;
  }
}
</style>