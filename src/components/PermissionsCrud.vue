<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error, loading } = useApi()
const permissions = ref([])
const newPermission = ref({ permission_name: '', description: '' })
const editingPermission = ref(null)
const editPermissionData = ref({ permission_name: '', description: '' })

const fetchPermissions = async () => {
  const result = await callApi({
    query: 'SELECT * FROM permissions'
  })
  if (result.success) {
    permissions.value = result.data
  }
}

const addPermission = async () => {
  if (!newPermission.value.permission_name) return
  const result = await callApi({
    query: 'INSERT INTO permissions (permission_name, description) VALUES (?, ?)',
    params: [newPermission.value.permission_name, newPermission.value.description]
  })
  if (result.success) {
    fetchPermissions()
    newPermission.value = { permission_name: '', description: '' }
  }
}

const startEditPermission = (perm) => {
  editingPermission.value = perm.id
  editPermissionData.value = { permission_name: perm.permission_name, description: perm.description }
}

const cancelEdit = () => {
  editingPermission.value = null
  editPermissionData.value = { permission_name: '', description: '' }
}

const updatePermission = async (perm) => {
  const result = await callApi({
    query: 'UPDATE permissions SET permission_name = ?, description = ? WHERE id = ?',
    params: [editPermissionData.value.permission_name, editPermissionData.value.description, perm.id]
  })
  if (result.success) {
    fetchPermissions()
    cancelEdit()
  }
}

const deletePermission = async (perm) => {
  // First confirm deletion
  if (!confirm(`Are you sure you want to delete the permission "${perm.permission_name}"?`)) {
    return
  }

  // Check if permission is used by any roles
  const checkResult = await callApi({
    query: 'SELECT COUNT(*) as count FROM role_permissions WHERE permission_id = ?',
    params: [perm.id]
  })

  if (checkResult.success && checkResult.data[0].count > 0) {
    error.value = `Cannot delete permission "${perm.permission_name}" because it is assigned to ${checkResult.data[0].count} role(s)`
    return
  }

  const result = await callApi({
    query: 'DELETE FROM permissions WHERE id = ?',
    params: [perm.id]
  })
  if (result.success) {
    fetchPermissions()
  }
}

onMounted(() => {
  fetchPermissions()
})
</script>

<template>
  <div class="permissions-container">
    <h2 class="title">Permissions</h2>
    <div class="content-container">
      <ul class="items-list">
        <li v-for="perm in permissions" :key="perm.id" class="item">
          <template v-if="editingPermission === perm.id">
            <div class="edit-form">
              <input v-model="editPermissionData.permission_name" placeholder="Permission name" class="input-field" />
              <input v-model="editPermissionData.description" placeholder="Description" class="input-field" />
              <div class="button-group">
                <button @click="updatePermission(perm)" :disabled="loading" class="btn save-btn">Save</button>
                <button @click="cancelEdit" class="btn cancel-btn">Cancel</button>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="item-content">
              <div class="item-text">
                <strong>{{ perm.permission_name }}</strong>
                <span class="description">{{ perm.description }}</span>
              </div>
              <div class="button-group">
                <button @click="startEditPermission(perm)" class="btn edit-btn">Edit</button>
                <button @click="deletePermission(perm)" :disabled="loading" class="btn delete-btn">Delete</button>
              </div>
            </div>
          </template>
        </li>
      </ul>
      <div class="add-form">
        <input v-model="newPermission.permission_name" placeholder="Permission name" class="input-field" />
        <input v-model="newPermission.description" placeholder="Description" class="input-field" />
        <button @click="addPermission" :disabled="loading" class="btn add-btn">Add Permission</button>
      </div>
      <div v-if="error" class="error-message">{{ error }}</div>
    </div>
  </div>
</template>

<style scoped>
.permissions-container {
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
  .permissions-container {
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