<template>
  <div class="update-db-structure-component">
    <div class="header-section">
      <h2>Update DB Structure</h2>
      <button @click="openCreateModal" class="btn-primary">
        <i class="fas fa-plus"></i>
        Add Update
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <i class="fas fa-spinner fa-spin"></i>
      <span>Loading...</span>
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="success-message">
      <i class="fas fa-check-circle"></i>
      {{ successMessage }}
    </div>

    <!-- Table -->
    <div v-if="!loading && updates.length > 0" class="table-container">
      <table class="updates-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>From Version</th>
            <th>Current Version</th>
            <th>Description</th>
            <th>SQL</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="update in updates" :key="update.id">
            <td>{{ update.id }}</td>
            <td>{{ update.from_version }}</td>
            <td>{{ update.current_version }}</td>
            <td>{{ update.description || '-' }}</td>
            <td class="sql-cell">
              <div class="sql-preview">{{ truncateSql(update.sql) }}</div>
            </td>
            <td class="actions-cell">
              <button @click="openEditModal(update)" class="btn-edit" title="Edit">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="confirmDelete(update)" class="btn-delete" title="Delete">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && updates.length === 0" class="empty-state">
      <i class="fas fa-database"></i>
      <p>No updates found. Click "Add Update" to create one.</p>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ editingUpdate ? 'Edit Update' : 'Add Update' }}</h3>
          <button @click="closeModal" class="modal-close">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <form @submit.prevent="saveUpdate" class="modal-body">
          <div class="form-group">
            <label for="from_version">From Version *</label>
            <input
              id="from_version"
              type="number"
              v-model="formData.from_version"
              required
              min="1"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label for="current_version">Current Version *</label>
            <input
              id="current_version"
              type="number"
              v-model="formData.current_version"
              required
              min="1"
              class="form-input"
            />
          </div>
          <div class="form-group">
            <label for="description">Description</label>
            <textarea
              id="description"
              v-model="formData.description"
              rows="3"
              class="form-input"
              placeholder="Optional description of the update"
            ></textarea>
          </div>
          <div class="form-group">
            <label for="sql">SQL *</label>
            <textarea
              id="sql"
              v-model="formData.sql"
              required
              rows="10"
              class="form-input sql-textarea"
              placeholder="Enter SQL statements for this update"
            ></textarea>
          </div>
          <div class="modal-actions">
            <button type="button" @click="closeModal" class="btn-cancel">Cancel</button>
            <button type="submit" :disabled="saving" class="btn-primary">
              <i v-if="saving" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-save"></i>
              {{ saving ? 'Saving...' : 'Save' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click="cancelDelete">
      <div class="modal-content delete-modal" @click.stop>
        <div class="modal-header">
          <h3>Confirm Delete</h3>
        </div>
        <div class="modal-body">
          <p>
            Are you sure you want to delete update <strong>#{{ updateToDelete?.id }}</strong>?
          </p>
          <p class="warning-text">This action cannot be undone.</p>
        </div>
        <div class="modal-actions">
          <button @click="cancelDelete" class="btn-cancel">Cancel</button>
          <button @click="deleteUpdate" :disabled="deleting" class="btn-delete-confirm">
            <i v-if="deleting" class="fas fa-spinner fa-spin"></i>
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const updates = ref([])
const loading = ref(false)
const error = ref('')
const successMessage = ref('')
const showModal = ref(false)
const showDeleteModal = ref(false)
const editingUpdate = ref(null)
const updateToDelete = ref(null)
const saving = ref(false)
const deleting = ref(false)

const formData = ref({
  from_version: '',
  current_version: '',
  description: '',
  sql: '',
})

// Get API base URL
const getApiBaseUrl = () => {
  const hostname = window.location.hostname
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
  return isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'
}

// Fetch all updates
const fetchUpdates = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php?action=get_db_updates`)
    const result = await response.json()

    if (result.success) {
      updates.value = result.data || []
    } else {
      error.value = result.message || 'Failed to fetch updates'
    }
  } catch (err) {
    error.value = 'An error occurred while fetching updates'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// Open create modal
const openCreateModal = () => {
  editingUpdate.value = null
  resetForm()
  showModal.value = true
}

// Open edit modal
const openEditModal = (update) => {
  editingUpdate.value = update
  formData.value = {
    from_version: update.from_version,
    current_version: update.current_version,
    description: update.description || '',
    sql: update.sql,
  }
  showModal.value = true
}

// Close modal
const closeModal = () => {
  showModal.value = false
  editingUpdate.value = null
  resetForm()
}

// Reset form
const resetForm = () => {
  formData.value = {
    from_version: '',
    current_version: '',
    description: '',
    sql: '',
  }
}

// Save update
const saveUpdate = async () => {
  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const action = editingUpdate.value ? 'update_db_update' : 'create_db_update'
    const body = {
      action,
      ...formData.value,
    }

    if (editingUpdate.value) {
      body.id = editingUpdate.value.id
    }

    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = result.message || 'Update saved successfully'
      showModal.value = false
      resetForm()
      await fetchUpdates()
      setTimeout(() => {
        successMessage.value = ''
      }, 3000)
    } else {
      error.value = result.message || 'Failed to save update'
    }
  } catch (err) {
    error.value = 'An error occurred while saving update'
    console.error(err)
  } finally {
    saving.value = false
  }
}

// Confirm delete
const confirmDelete = (update) => {
  updateToDelete.value = update
  showDeleteModal.value = true
}

// Cancel delete
const cancelDelete = () => {
  showDeleteModal.value = false
  updateToDelete.value = null
}

// Delete update
const deleteUpdate = async () => {
  deleting.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'delete_db_update',
        id: updateToDelete.value.id,
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = result.message || 'Update deleted successfully'
      showDeleteModal.value = false
      updateToDelete.value = null
      await fetchUpdates()
      setTimeout(() => {
        successMessage.value = ''
      }, 3000)
    } else {
      error.value = result.message || 'Failed to delete update'
    }
  } catch (err) {
    error.value = 'An error occurred while deleting update'
    console.error(err)
  } finally {
    deleting.value = false
  }
}

// Truncate SQL for preview
const truncateSql = (sql) => {
  if (!sql) return '-'
  if (sql.length <= 100) return sql
  return sql.substring(0, 100) + '...'
}

onMounted(() => {
  fetchUpdates()
})
</script>

<style scoped>
.update-db-structure-component {
  padding: 0;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.btn-primary {
  background-color: #409eff;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.btn-primary:hover {
  background-color: #66b1ff;
}

.loading-state,
.error-message,
.success-message {
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.loading-state {
  background-color: #f0f9ff;
  color: #409eff;
}

.error-message {
  background-color: #fef0f0;
  color: #f56c6c;
  border: 1px solid #fde2e2;
}

.success-message {
  background-color: #f0f9eb;
  color: #67c23a;
  border: 1px solid #e1f3d8;
}

.table-container {
  overflow-x: auto;
  position: relative;
  z-index: 0;
}

.updates-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
}

.updates-table th {
  background-color: #f5f7fa;
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #2c3e50;
  border-bottom: 2px solid #e4e7ed;
}

.updates-table td {
  padding: 1rem;
  border-bottom: 1px solid #e4e7ed;
  color: #606266;
}

.updates-table tr:hover {
  background-color: #f5f7fa;
}

.sql-cell {
  max-width: 300px;
}

.sql-preview {
  font-family: 'Courier New', monospace;
  font-size: 0.875rem;
  color: #606266;
  word-break: break-word;
}

.actions-cell {
  white-space: nowrap;
}

.btn-edit,
.btn-delete {
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  margin-right: 0.5rem;
  transition: all 0.3s ease;
}

.btn-edit {
  background-color: #409eff;
  color: white;
}

.btn-edit:hover {
  background-color: #66b1ff;
}

.btn-delete {
  background-color: #f56c6c;
  color: white;
}

.btn-delete:hover {
  background-color: #f78989;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #909399;
}

.empty-state i {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 2000;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow-y: auto;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e4e7ed;
}

.modal-header h3 {
  margin: 0;
  color: #2c3e50;
}

.modal-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #909399;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.modal-close:hover {
  background-color: #f5f7fa;
  color: #606266;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  color: #606266;
  font-weight: 500;
}

.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.3s ease;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: #409eff;
  outline: none;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.sql-textarea {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  resize: vertical;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e4e7ed;
}

.btn-cancel {
  background-color: #909399;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.btn-cancel:hover {
  background-color: #a6a9ad;
}

.btn-delete-confirm {
  background-color: #f56c6c;
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.btn-delete-confirm:hover:not(:disabled) {
  background-color: #f78989;
}

.btn-delete-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.warning-text {
  color: #f56c6c;
  font-weight: 500;
  margin-top: 0.5rem;
}
</style>

