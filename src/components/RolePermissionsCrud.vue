<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { useApi } from '../composables/useApi'

const props = defineProps({
  selectedRoleId: {
    type: Number,
    default: null,
  },
})

const { callApi, error, loading } = useApi()
const rolePermissions = ref([])
const allPermissions = ref([])

// Computed property to get only permissions assigned to selected role
const assignedPermissions = computed(() => {
  if (!props.selectedRoleId || rolePermissions.value.length === 0) {
    return []
  }
  
  // Get permission IDs assigned to this role
  const assignedPermissionIds = rolePermissions.value
    .filter(rp => rp.role_id === props.selectedRoleId)
    .map(rp => rp.permission_id)
  
  // Return full permission objects for assigned permissions
  return allPermissions.value.filter(p => assignedPermissionIds.includes(p.id))
})

// Computed property to get permissions NOT assigned to selected role
const unassignedPermissions = computed(() => {
  if (!props.selectedRoleId || allPermissions.value.length === 0) {
    return allPermissions.value
  }
  
  // Get permission IDs assigned to this role
  const assignedPermissionIds = rolePermissions.value
    .filter(rp => rp.role_id === props.selectedRoleId)
    .map(rp => rp.permission_id)
  
  // Return permissions that are NOT assigned
  return allPermissions.value.filter(p => !assignedPermissionIds.includes(p.id))
})

const fetchRolePermissions = async () => {
  if (!props.selectedRoleId) {
    rolePermissions.value = []
    return
  }
  
  const result = await callApi({
    query: 'SELECT * FROM role_permissions WHERE role_id = ?',
    params: [props.selectedRoleId],
  })
  if (result.success) {
    rolePermissions.value = result.data
  }
}

const fetchPermissions = async () => {
  const result = await callApi({
    query: 'SELECT * FROM permissions ORDER BY permission_name',
  })
  if (result.success) {
    allPermissions.value = result.data
  }
}

const hasPermission = (permission_id) => {
  if (!props.selectedRoleId) return false
  return rolePermissions.value.some(
    (rp) => rp.role_id === props.selectedRoleId && rp.permission_id === permission_id,
  )
}

const togglePermission = async (permission_id, checked) => {
  if (!props.selectedRoleId) return

  if (checked) {
    // Add permission
    const result = await callApi({
      query: 'INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)',
      params: [props.selectedRoleId, permission_id],
    })
    if (result.success) {
      fetchRolePermissions()
    }
  } else {
    // Remove permission
    const result = await callApi({
      query: 'DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?',
      params: [props.selectedRoleId, permission_id],
    })
    if (result.success) {
      fetchRolePermissions()
    }
  }
}

const addAllPermissions = async () => {
  if (!props.selectedRoleId || allPermissions.value.length === 0) {
    console.log('Cannot add permissions: no role selected or no permissions available')
    return
  }

  // Get permission IDs that are not already assigned
  const assignedPermissionIds = rolePermissions.value
    .filter(rp => rp.role_id === props.selectedRoleId)
    .map(rp => rp.permission_id)
  
  const unassignedPermissions = allPermissions.value.filter(
    p => !assignedPermissionIds.includes(p.id)
  )

  if (unassignedPermissions.length === 0) {
    alert('All permissions are already assigned to this role')
    return
  }

  // Insert permissions one by one to avoid bulk insert issues
  let successCount = 0
  let failCount = 0

  for (const permission of unassignedPermissions) {
    const result = await callApi({
      query: 'INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)',
      params: [props.selectedRoleId, permission.id],
    })
    
    if (result.success) {
      successCount++
    } else {
      failCount++
      console.error(`Failed to add permission ${permission.permission_name}:`, result.error)
    }
  }

  // Refresh the permissions list
  await fetchRolePermissions()

  if (failCount === 0) {
    // Success message is optional, but we can show it
    console.log(`Successfully added ${successCount} permission(s)`)
  } else {
    alert(`Added ${successCount} permission(s), but ${failCount} failed. Please check the console for details.`)
  }
}

const removeAllPermissions = async () => {
  if (!props.selectedRoleId) {
    console.log('Cannot remove permissions: no role selected')
    return
  }

  if (assignedPermissions.value.length === 0) {
    alert('This role has no permissions to remove')
    return
  }

  if (!confirm(`Are you sure you want to remove all ${assignedPermissions.value.length} permission(s) from this role?`)) {
    return
  }

  const result = await callApi({
    query: 'DELETE FROM role_permissions WHERE role_id = ?',
    params: [props.selectedRoleId],
  })
  
  if (result.success) {
    await fetchRolePermissions()
    console.log('Successfully removed all permissions')
  } else {
    error.value = result.error || 'Failed to remove permissions'
    console.error('Failed to remove permissions:', result.error)
  }
}

// Watch for selectedRoleId changes
watch(
  () => props.selectedRoleId,
  (newRoleId) => {
    if (newRoleId) {
      fetchRolePermissions()
    } else {
      rolePermissions.value = []
    }
  },
  { immediate: true }
)

onMounted(() => {
  fetchPermissions()
  if (props.selectedRoleId) {
    fetchRolePermissions()
  }
})
</script>

<template>
  <div class="role-permissions-container">
    <div class="header-section">
      <h3 class="main-title">
        <i class="fas fa-shield-alt"></i>
        Role Permissions
      </h3>
      <div class="action-buttons">
        <button
          @click="addAllPermissions"
          :disabled="!selectedRoleId || loading || unassignedPermissions.length === 0"
          class="btn btn-add-all"
          title="Add all available permissions to this role"
        >
          <i class="fas fa-plus-circle"></i>
          Add All Permissions
        </button>
        <button
          @click="removeAllPermissions"
          :disabled="!selectedRoleId || loading || assignedPermissions.length === 0"
          class="btn btn-remove-all"
          title="Remove all permissions from this role"
        >
          <i class="fas fa-minus-circle"></i>
          Remove All
        </button>
      </div>
    </div>
    <div v-if="!selectedRoleId" class="no-role-selected">
      <i class="fas fa-info-circle"></i>
      <p>Please select a role to manage its permissions</p>
    </div>
    <div v-else>
      <div v-if="assignedPermissions.length === 0" class="no-permissions">
        <i class="fas fa-info-circle"></i>
        <p>This role has no permissions assigned yet.</p>
        <p class="hint">Use the checkboxes below to assign permissions.</p>
      </div>
      <div v-else>
        <h5 class="section-subtitle">
          <i class="fas fa-check-circle"></i>
          Assigned Permissions ({{ assignedPermissions.length }})
        </h5>
        <div class="permissions-list">
          <div
            v-for="permission in assignedPermissions"
            :key="permission.id"
            class="permission-item has-permission"
          >
            <label class="checkbox-label">
              <input
                type="checkbox"
                :checked="true"
                @change="togglePermission(permission.id, $event.target.checked)"
                :disabled="loading"
              />
              <span class="checkbox-text">{{ permission.permission_name }}</span>
            </label>
          </div>
        </div>
      </div>
      
      <!-- Unassigned permissions for assignment -->
      <div v-if="unassignedPermissions.length > 0" class="all-permissions-section">
        <h5 class="section-subtitle">
          <i class="fas fa-list-check"></i>
          Add More Permissions
        </h5>
        <div class="permissions-list">
          <div
            v-for="permission in unassignedPermissions"
            :key="permission.id"
            class="permission-item"
          >
            <label class="checkbox-label">
              <input
                type="checkbox"
                :checked="false"
                @change="togglePermission(permission.id, $event.target.checked)"
                :disabled="loading"
              />
              <span class="checkbox-text">{{ permission.permission_name }}</span>
            </label>
          </div>
        </div>
      </div>
      <div v-else-if="allPermissions.length > 0 && assignedPermissions.length > 0" class="all-permissions-section">
        <div class="no-permissions">
          <i class="fas fa-check-circle"></i>
          <p>All permissions are assigned to this role</p>
        </div>
      </div>
    </div>
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>
  </div>
</template>

<style scoped>
.role-permissions-container {
  background: white;
  border-radius: 8px;
  padding: 0;
}

.header-section {
  border-bottom: 2px solid #e5e7eb;
  padding: 16px;
  margin-bottom: 20px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 16px;
  background: #f8f9fa;
  border-radius: 8px 8px 0 0;
  min-height: 60px;
  position: relative;
  z-index: 1;
}

.main-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 10px;
}

.main-title i {
  color: #6366f1;
}

.action-buttons {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
  visibility: visible !important;
  opacity: 1 !important;
}

.btn {
  padding: 10px 18px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  gap: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-add-all {
  background: #10b981;
  color: white;
}

.btn-add-all:hover:not(:disabled) {
  background: #059669;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(5, 150, 105, 0.2);
}

.btn-remove-all {
  background: #ef4444;
  color: white;
}

.btn-remove-all:hover:not(:disabled) {
  background: #dc2626;
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(239, 68, 68, 0.2);
}

.no-role-selected,
.no-permissions {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  text-align: center;
  color: #9ca3af;
}

.no-role-selected i,
.no-permissions i {
  font-size: 36px;
  margin-bottom: 12px;
  color: #d1d5db;
}

.no-role-selected p,
.no-permissions p {
  margin: 0;
  font-size: 14px;
  color: #6b7280;
}

.no-permissions .hint {
  margin-top: 8px;
  font-size: 12px;
  color: #9ca3af;
  font-style: italic;
}

.all-permissions-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 2px solid #e5e7eb;
}

.section-subtitle {
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-subtitle i {
  color: #6366f1;
}

.permissions-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
  padding: 16px;
}

.permission-item {
  background: #f8f9fa;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  padding: 12px 16px;
  transition: all 0.2s ease;
  cursor: pointer;
}

.permission-item:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.permission-item.has-permission {
  background: #eff6ff;
  border-color: #3b82f6;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.1);
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  user-select: none;
  width: 100%;
  margin: 0;
}

.checkbox-text {
  color: #374151;
  font-size: 14px;
  font-weight: 500;
  flex: 1;
}

.permission-item.has-permission .checkbox-text {
  color: #1e40af;
  font-weight: 600;
}

input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #3b82f6;
  flex-shrink: 0;
}

input[type="checkbox"]:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.error-message {
  color: #dc2626;
  background: #fee2e2;
  padding: 12px 16px;
  border-radius: 6px;
  margin: 16px;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 8px;
  border: 1px solid #fecaca;
}

.error-message i {
  font-size: 16px;
}

@media (max-width: 768px) {
  .permissions-list {
    grid-template-columns: 1fr;
    padding: 12px;
  }
}
</style>