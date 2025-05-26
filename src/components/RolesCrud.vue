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
    query: 'SELECT * FROM roles',
  })
  if (result.success) {
    roles.value = result.data
  }
}

const addRole = async () => {
  if (!newRole.value.role_name) return
  const result = await callApi({
    query: 'INSERT INTO roles (role_name, description) VALUES (?, ?)',
    params: [newRole.value.role_name, newRole.value.description],
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
    params: [editRoleData.value.role_name, editRoleData.value.description, role.id],
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
    params: [role.id],
  })

  if (checkResult.success && checkResult.data[0].count > 0) {
    error.value = `Cannot delete role "${role.role_name}" because it is assigned to ${checkResult.data[0].count} user(s)`
    return
  }

  const result = await callApi({
    query: 'DELETE FROM roles WHERE id = ?',
    params: [role.id],
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
    <!-- Loading Overlay -->
    <div v-if="loading" class="loading-overlay">
      <i class="fas fa-spinner fa-spin fa-2x"></i>
      <span>Processing...</span>
    </div>

    <h2 class="title">
      <i class="fas fa-user-tag"></i>
      Roles
    </h2>

    <div class="content-container">
      <ul class="items-list">
        <li
          v-for="role in roles"
          :key="role.id"
          class="item"
          :class="{ editing: editingRole === role.id }"
        >
          <template v-if="editingRole === role.id">
            <div class="edit-form">
              <div class="input-wrapper">
                <i class="fas fa-tag input-icon"></i>
                <input
                  v-model="editRoleData.role_name"
                  placeholder="Role name"
                  class="input-field with-icon"
                  :disabled="loading"
                />
              </div>
              <div class="input-wrapper">
                <i class="fas fa-info-circle input-icon"></i>
                <input
                  v-model="editRoleData.description"
                  placeholder="Description"
                  class="input-field with-icon"
                  :disabled="loading"
                />
              </div>
              <div class="button-group">
                <button
                  @click="updateRole(role)"
                  :disabled="loading || !editRoleData.role_name.trim()"
                  class="btn save-btn"
                >
                  <i class="fas" :class="loading ? 'fa-spinner fa-spin' : 'fa-save'"></i>
                  {{ loading ? 'Saving...' : 'Save' }}
                </button>
                <button @click="cancelEdit" class="btn cancel-btn" :disabled="loading">
                  <i class="fas fa-times"></i>
                  Cancel
                </button>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="item-content">
              <div class="item-text">
                <strong>
                  <i class="fas fa-tag text-blue"></i>
                  {{ role.role_name }}
                </strong>
                <span class="description">
                  <i class="fas fa-info-circle text-gray"></i>
                  {{ role.description || 'No description' }}
                </span>
              </div>
              <div class="button-group">
                <button
                  @click="startEditRole(role)"
                  class="btn edit-btn"
                  :disabled="loading || editingRole !== null"
                  title="Edit role"
                >
                  <i class="fas fa-edit"></i>
                  Edit
                </button>
                <button
                  @click="deleteRole(role)"
                  :disabled="loading || editingRole !== null"
                  class="btn delete-btn"
                  title="Delete role"
                >
                  <i class="fas fa-trash-alt"></i>
                  Delete
                </button>
              </div>
            </div>
          </template>
        </li>
      </ul>

      <div class="add-form">
        <h3>
          <i class="fas fa-plus-circle"></i>
          Add New Role
        </h3>
        <div class="input-wrapper">
          <i class="fas fa-tag input-icon"></i>
          <input
            v-model="newRole.role_name"
            placeholder="Role name"
            class="input-field with-icon"
            :disabled="loading"
          />
        </div>
        <div class="input-wrapper">
          <i class="fas fa-info-circle input-icon"></i>
          <input
            v-model="newRole.description"
            placeholder="Description"
            class="input-field with-icon"
            :disabled="loading"
          />
        </div>
        <button
          @click="addRole"
          :disabled="loading || !newRole.role_name.trim()"
          class="btn add-btn"
        >
          <i class="fas" :class="loading ? 'fa-spinner fa-spin' : 'fa-plus'"></i>
          {{ loading ? 'Adding...' : 'Add Role' }}
        </button>
      </div>

      <div v-if="error" class="error-message">
        <i class="fas fa-exclamation-circle"></i>
        {{ error }}
      </div>
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
  border-radius: 8px;
  backdrop-filter: blur(2px);
}

.loading-overlay i {
  color: #3b82f6;
}

.loading-overlay span {
  color: #4b5563;
  font-size: 0.875rem;
}

.title {
  color: #2c3e50;
  font-size: 24px;
  margin-bottom: 20px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 12px;
}

.title i {
  color: #3b82f6;
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
  transition: all 0.2s ease;
  border: 1px solid transparent;
}

.item:hover {
  background: #f1f5f9;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.item.editing {
  background: #f0f9ff;
  border-color: #3b82f6;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.1);
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
  gap: 8px;
}

.item-text strong {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #1e293b;
}

.description {
  color: #64748b;
  font-size: 0.9em;
  display: flex;
  align-items: center;
  gap: 8px;
}

.text-blue {
  color: #3b82f6;
}

.text-gray {
  color: #6b7280;
}

.button-group {
  display: flex;
  gap: 8px;
}

.input-wrapper {
  position: relative;
  margin-bottom: 12px;
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
  border: 1px solid #e2e8f0;
  border-radius: 6px;
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
  border-radius: 6px;
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
  margin-top: 24px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 8px;
  border: 1px solid #e2e8f0;
}

.add-form h3 {
  margin: 0 0 16px 0;
  color: #2c3e50;
  font-size: 18px;
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
  padding: 10px;
}

.add-btn:hover:not(:disabled) {
  background: #2563eb;
  transform: translateY(-1px);
}

.error-message {
  background-color: #fee2e2;
  color: #dc2626;
  padding: 12px;
  border-radius: 6px;
  margin-top: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  animation: slideIn 0.3s ease;
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

@media (max-width: 640px) {
  .item-content {
    flex-direction: column;
    align-items: stretch;
  }

  .button-group {
    margin-top: 12px;
  }

  .btn {
    flex: 1;
    justify-content: center;
  }
}
</style>
