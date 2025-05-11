<script setup>
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

const { callApi, error, loading } = useApi()
const rolePermissions = ref([])
const roles = ref([])
const permissions = ref([])

const fetchRolePermissions = async () => {
  const result = await callApi({
    query: 'SELECT * FROM role_permissions'
  })
  if (result.success) {
    rolePermissions.value = result.data
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

const fetchPermissions = async () => {
  const result = await callApi({
    query: 'SELECT * FROM permissions'
  })
  if (result.success) {
    permissions.value = result.data
  }
}

const hasPermission = (role_id, permission_id) => {
  return rolePermissions.value.some(rp => 
    rp.role_id === role_id && rp.permission_id === permission_id
  )
}

const togglePermission = async (role_id, permission_id, checked) => {
  if (checked) {
    // Add permission
    const result = await callApi({
      query: 'INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)',
      params: [role_id, permission_id]
    })
    if (result.success) {
      fetchRolePermissions()
    }
  } else {
    // Remove permission
    const result = await callApi({
      query: 'DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?',
      params: [role_id, permission_id]
    })
    if (result.success) {
      fetchRolePermissions()
    }
  }
}

onMounted(() => {
  fetchRolePermissions()
  fetchRoles()
  fetchPermissions()
})
</script>

<template>
  <div class="role-permissions-container">
    <h2 class="title">Role Permissions</h2>
    <div class="table-container">
      <table>
        <thead>
          <tr>
            <th>Role</th>
            <th>Permissions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="role in roles" :key="role.id">
            <td class="role-name">{{ role.role_name }}</td>
            <td class="permissions-cell">
              <div class="permission-item" v-for="permission in permissions" :key="permission.id">
                <label class="checkbox-label">
                  <input 
                    type="checkbox" 
                    :checked="hasPermission(role.id, permission.id)"
                    @change="togglePermission(role.id, permission.id, $event.target.checked)"
                    :disabled="loading"
                  />
                  <span class="checkbox-text">{{ permission.permission_name }}</span>
                </label>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
    <div v-if="error" class="error-message">{{ error }}</div>
  </div>
</template>

<style scoped>
.role-permissions-container {
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

.table-container {
  overflow-x: auto;
  margin: 20px 0;
  border-radius: 6px;
  border: 1px solid #eaeaea;
}

table {
  border-collapse: collapse;
  width: 100%;
  background: white;
}

th, td {
  border: 1px solid #eaeaea;
  padding: 12px 16px;
  text-align: left;
}

th {
  background-color: #f8f9fa;
  color: #2c3e50;
  font-weight: 600;
  white-space: nowrap;
}

.role-name {
  font-weight: 500;
  color: #2c3e50;
}

.permissions-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  min-width: 300px;
  padding: 8px;
}

.permission-item {
  background: #f8f9fa;
  border-radius: 6px;
  padding: 6px 12px;
  transition: background-color 0.2s;
}

.permission-item:hover {
  background: #edf2f7;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  user-select: none;
}

.checkbox-text {
  color: #4a5568;
  font-size: 14px;
}

input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: #3b82f6;
}

input[type="checkbox"]:disabled {
  cursor: not-allowed;
  opacity: 0.6;
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
  .role-permissions-container {
    padding: 16px;
  }

  th, td {
    padding: 8px;
  }

  .permissions-cell {
    gap: 8px;
  }
}
</style>