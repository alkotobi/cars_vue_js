<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import RolePermissionsCrud from '../RolePermissionsCrud.vue'
import AddRoleForm from './AddRoleForm.vue'

const emit = defineEmits(['close'])

const { callApi, loading } = useApi()
const roles = ref([])
const selectedRole = ref(null)
const showAddRoleModal = ref(false)
const editingRole = ref(null)
const editRoleData = ref({ role_name: '', description: '' })

const fetchRoles = async () => {
  const result = await callApi({
    query: 'SELECT * FROM roles ORDER BY role_name',
  })
  if (result.success) {
    roles.value = result.data
  }
}

const selectRole = (role) => {
  selectedRole.value = role
}

const handleRoleAdded = () => {
  fetchRoles()
  showAddRoleModal.value = false
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
    // Update selected role if it was the one being edited
    if (selectedRole.value && selectedRole.value.id === role.id) {
      selectedRole.value = { ...selectedRole.value, ...editRoleData.value }
    }
    cancelEdit()
  }
}

const deleteRole = async (role) => {
  if (!confirm(`Are you sure you want to delete the role "${role.role_name}"?`)) {
    return
  }

  // Check if role is used by any users
  const checkResult = await callApi({
    query: 'SELECT COUNT(*) as count FROM users WHERE role_id = ?',
    params: [role.id],
  })

  if (checkResult.success && checkResult.data[0].count > 0) {
    alert(`Cannot delete role "${role.role_name}" because it is assigned to ${checkResult.data[0].count} user(s)`)
    return
  }

  const result = await callApi({
    query: 'DELETE FROM roles WHERE id = ?',
    params: [role.id],
  })
  if (result.success) {
    // Clear selection if deleted role was selected
    if (selectedRole.value && selectedRole.value.id === role.id) {
      selectedRole.value = null
    }
    fetchRoles()
  }
}

onMounted(() => {
  fetchRoles()
})
</script>

<template>
  <teleport to="body">
    <div class="modal-overlay" @click.self="emit('close')">
      <div class="modal-container" @click.stop>
        <div class="modal-header">
          <h3>
            <i class="fas fa-shield-alt"></i>
            Roles & Permissions Management
          </h3>
          <button @click.stop="emit('close')" class="close-btn" type="button">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <div class="modal-content">
          <div class="roles-section">
            <div class="section-header">
              <h4>
                <i class="fas fa-list"></i>
                Roles
              </h4>
              <button @click="showAddRoleModal = true" class="btn-add-role" :disabled="loading">
                <i class="fas fa-plus"></i>
                Add Role
              </button>
            </div>

            <!-- Roles Table -->
            <div class="table-wrapper">
              <table class="roles-table">
                <thead>
                  <tr>
                    <th>Role Name</th>
                    <th>Description</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="role in roles"
                    :key="role.id"
                    :class="{ 
                      'selected': selectedRole && selectedRole.id === role.id,
                      'editing': editingRole === role.id
                    }"
                    @click="selectRole(role)"
                  >
                    <template v-if="editingRole === role.id">
                      <td>
                        <input
                          v-model="editRoleData.role_name"
                          class="edit-input"
                          :disabled="loading"
                          @click.stop
                        />
                      </td>
                      <td>
                        <input
                          v-model="editRoleData.description"
                          class="edit-input"
                          :disabled="loading"
                          @click.stop
                        />
                      </td>
                      <td class="actions-cell" @click.stop>
                        <button
                          @click="updateRole(role)"
                          :disabled="loading || !editRoleData.role_name.trim()"
                          class="btn btn-sm btn-save"
                        >
                          <i class="fas fa-save"></i>
                        </button>
                        <button
                          @click="cancelEdit"
                          :disabled="loading"
                          class="btn btn-sm btn-cancel"
                        >
                          <i class="fas fa-times"></i>
                        </button>
                      </td>
                    </template>
                    <template v-else>
                      <td class="role-name-cell">
                        <i class="fas fa-shield-alt"></i>
                        {{ role.role_name }}
                      </td>
                      <td class="description-cell">
                        {{ role.description || '-' }}
                      </td>
                      <td class="actions-cell" @click.stop>
                        <button
                          @click.stop="startEditRole(role)"
                          :disabled="loading || editingRole !== null"
                          class="btn btn-sm btn-edit"
                          title="Edit role"
                        >
                          <i class="fas fa-edit"></i>
                        </button>
                        <button
                          @click.stop="deleteRole(role)"
                          :disabled="loading || editingRole !== null"
                          class="btn btn-sm btn-delete"
                          title="Delete role"
                        >
                          <i class="fas fa-trash"></i>
                        </button>
                      </td>
                    </template>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="permissions-section">
            <div v-if="selectedRole" class="permissions-content">
              <div class="section-header">
                <h4>
                  <i class="fas fa-key"></i>
                  Permissions for: <span class="selected-role-name">{{ selectedRole.role_name }}</span>
                </h4>
              </div>
              <RolePermissionsCrud :selected-role-id="selectedRole?.id" />
            </div>
            <div v-else class="no-selection">
              <i class="fas fa-hand-pointer"></i>
              <p>Select a role from the table to view and manage its permissions</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Add Role Modal -->
    <AddRoleForm v-if="showAddRoleModal" @role-added="handleRoleAdded" @close="showAddRoleModal = false" />
  </teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
  overflow-y: auto;
}

.modal-container {
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  width: 100%;
  max-width: 1400px;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 2px solid #e5e7eb;
  background: #f8f9fa;
  flex-shrink: 0;
}

.modal-header h3 {
  margin: 0;
  font-size: 22px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 12px;
}

.modal-header h3 i {
  color: #6366f1;
}

.close-btn {
  background: #ef4444;
  color: white;
  border: none;
  border-radius: 6px;
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s ease;
  flex-shrink: 0;
  position: relative;
  z-index: 10;
  pointer-events: auto;
}

.close-btn:hover {
  background: #dc2626;
  transform: scale(1.05);
}

.modal-content {
  padding: 24px;
  overflow-y: auto;
  display: grid;
  grid-template-columns: 1fr 1.5fr;
  gap: 24px;
  min-height: 500px;
}

.roles-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 12px;
  border-bottom: 2px solid #e5e7eb;
}

.section-header h4 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header h4 i {
  color: #6366f1;
}

.selected-role-name {
  color: #6366f1;
  font-weight: 700;
}


.btn-add-role {
  padding: 6px 12px;
  background: #10b981;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.btn-add-role:hover:not(:disabled) {
  background: #059669;
}

.btn-add-role:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.table-wrapper {
  border: 1px solid #e5e7eb;
  border-radius: 6px;
  overflow: hidden;
}

.roles-table {
  width: 100%;
  border-collapse: collapse;
}

.roles-table thead {
  background: #f8f9fa;
}

.roles-table th {
  padding: 12px;
  text-align: left;
  font-weight: 600;
  color: #374151;
  font-size: 14px;
  border-bottom: 2px solid #e5e7eb;
}

.roles-table td {
  padding: 12px;
  border-bottom: 1px solid #e5e7eb;
}

.roles-table tbody tr {
  cursor: pointer;
  transition: all 0.2s ease;
}

.roles-table tbody tr:hover {
  background: #f8fafc;
}

.roles-table tbody tr.selected {
  background: #eff6ff;
  border-left: 3px solid #3b82f6;
}

.roles-table tbody tr.editing {
  background: #f0f9ff;
}

.role-name-cell {
  font-weight: 600;
  color: #1e293b;
  display: flex;
  align-items: center;
  gap: 8px;
}

.role-name-cell i {
  color: #6366f1;
}

.description-cell {
  color: #64748b;
  font-size: 14px;
}

.actions-cell {
  width: 100px;
}

.actions-cell .btn {
  margin-right: 4px;
}

.btn {
  padding: 6px 10px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-primary {
  background: #3b82f6;
  color: white;
}

.btn-primary:hover:not(:disabled) {
  background: #2563eb;
}

.btn-edit {
  background: #3b82f6;
  color: white;
}

.btn-edit:hover:not(:disabled) {
  background: #2563eb;
}

.btn-delete {
  background: #ef4444;
  color: white;
}

.btn-delete:hover:not(:disabled) {
  background: #dc2626;
}

.btn-save {
  background: #10b981;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background: #059669;
}

.btn-cancel {
  background: #6b7280;
  color: white;
}

.btn-cancel:hover:not(:disabled) {
  background: #4b5563;
}

.edit-input {
  width: 100%;
  padding: 6px 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 14px;
}

.edit-input:focus {
  outline: none;
  border-color: #3b82f6;
}

.permissions-section {
  display: flex;
  flex-direction: column;
}

.permissions-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.no-selection {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
  color: #9ca3af;
  background: #f9fafb;
  border-radius: 8px;
  border: 2px dashed #e5e7eb;
}

.no-selection i {
  font-size: 48px;
  margin-bottom: 16px;
  color: #d1d5db;
}

.no-selection p {
  margin: 0;
  font-size: 16px;
  color: #6b7280;
}

@media (max-width: 1024px) {
  .modal-content {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .modal-overlay {
    padding: 10px;
  }

  .modal-container {
    max-height: 95vh;
  }

  .modal-header {
    padding: 16px 20px;
  }

  .modal-header h3 {
    font-size: 18px;
  }

  .modal-content {
    padding: 16px;
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }

  .btn-add-role {
    width: 100%;
    justify-content: center;
  }
}
</style>

