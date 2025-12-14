<template>
  <div class="databases-component">
    <div class="header-section">
      <h2>Databases Management</h2>
      <button @click="openCreateModal" class="btn-primary">
        <i class="fas fa-plus"></i>
        Add Database
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

    <!-- Toolbar -->
    <div v-if="!loading && databases.length > 0" class="toolbar">
      <div class="toolbar-left">
        <button @click="updateStructure" class="btn-toolbar" :disabled="selectedDatabases.length === 0">
          <i class="fas fa-sync-alt"></i>
          Update Structure
        </button>
        <button @click="runSql" class="btn-toolbar" :disabled="selectedDatabases.length === 0">
          <i class="fas fa-code"></i>
          Run SQL
        </button>
      </div>
      <div class="toolbar-right">
        <span v-if="selectedDatabases.length > 0" class="selected-count">
          {{ selectedDatabases.length }} database(s) selected
        </span>
      </div>
    </div>

    <!-- Table -->
    <div v-if="!loading && databases.length > 0" class="table-container">
      <table class="databases-table">
        <thead>
          <tr>
            <th>
              <input
                type="checkbox"
                :checked="allSelected"
                @change="toggleSelectAll"
                class="checkbox-select-all"
              />
            </th>
            <th>ID</th>
            <th>DB Code</th>
            <th>DB Name</th>
            <th>DB Host Start</th>
            <th>DB Host End</th>
            <th>DB Host Cost/Month</th>
            <th>Serv Host Start</th>
            <th>Serv Host End</th>
            <th>Serv Host Cost/Month</th>
            <th>Files Dir</th>
            <th>JS Dir</th>
            <th>Status</th>
            <th>Version</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="db in databases" :key="db.id">
            <td>
              <input
                type="checkbox"
                :value="db.id"
                v-model="selectedDatabases"
                class="checkbox-db"
              />
            </td>
            <td>{{ db.id }}</td>
            <td>{{ db.db_code }}</td>
            <td>{{ db.db_name }}</td>
            <td>{{ formatDate(db.db_host_start) }}</td>
            <td>{{ formatDate(db.db_host_end) }}</td>
            <td>{{ formatCurrency(db.db_host_cost_per_month) }}</td>
            <td>{{ formatDate(db.serv_host_start) }}</td>
            <td>{{ formatDate(db.serv_host_end) }}</td>
            <td>{{ formatCurrency(db.serv_host_cost_per_month) }}</td>
            <td>{{ db.files_dir }}</td>
            <td>{{ db.js_dir }}</td>
            <td>
              <span :class="['status-badge', db.is_created == 1 ? 'status-created' : 'status-pending']">
                {{ db.is_created == 1 ? 'Created' : 'Pending' }}
              </span>
            </td>
            <td>
              <span v-if="db.is_created == 1 && db.version !== null && db.version !== undefined" class="version-badge">
                v{{ db.version }}
              </span>
              <span v-else-if="db.is_created == 1" class="version-badge version-unknown">
                N/A
              </span>
              <span v-else class="version-badge version-pending">
                -
              </span>
            </td>
            <td class="actions-cell">
              <button
                @click="createTables(db)"
                :disabled="db.is_created == 1 || creatingTables"
                class="btn-create-tables"
                :title="db.is_created == 1 ? 'Tables already created' : 'Create tables from setup.sql'"
              >
                <i class="fas fa-database"></i>
                {{ creatingTables && creatingTablesId === db.id ? 'Creating...' : 'Create Tables' }}
              </button>
              <button @click="openEditModal(db)" class="btn-edit" title="Edit">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="confirmDelete(db)" class="btn-delete" title="Delete">
                <i class="fas fa-trash"></i>
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && databases.length === 0" class="empty-state">
      <i class="fas fa-database"></i>
      <p>No databases found. Click "Add Database" to create one.</p>
    </div>

    <!-- Create/Edit Modal -->
    <div v-if="showModal" class="modal-overlay" @click="closeModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ editingDatabase ? 'Edit Database' : 'Add Database' }}</h3>
          <button @click="closeModal" class="modal-close">
            <i class="fas fa-times"></i>
          </button>
        </div>
        <form @submit.prevent="saveDatabase" class="modal-form">
          <div class="form-row">
            <div v-if="editingDatabase" class="form-group">
              <label>DB Code</label>
              <input
                v-model="formData.db_code"
                type="text"
                readonly
                disabled
                class="readonly-input"
              />
              <small class="help-text">Auto-generated (cannot be changed)</small>
            </div>
            <div class="form-group" :class="{ 'full-width': !editingDatabase }">
              <label>DB Name *</label>
              <input
                v-model="formData.db_name"
                type="text"
                required
                placeholder="e.g., merhab_cars"
              />
              <small v-if="!editingDatabase" class="help-text"
                >DB Code will be auto-generated from DB Name</small
              >
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>DB Host Start</label>
              <input v-model="formData.db_host_start" type="date" />
            </div>
            <div class="form-group">
              <label>DB Host End</label>
              <input v-model="formData.db_host_end" type="date" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>DB Host Cost/Month</label>
              <input
                v-model="formData.db_host_cost_per_month"
                type="number"
                step="0.01"
                placeholder="0.00"
              />
            </div>
            <div class="form-group">
              <label>Serv Host Start</label>
              <input v-model="formData.serv_host_start" type="date" />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Serv Host End</label>
              <input v-model="formData.serv_host_end" type="date" />
            </div>
            <div class="form-group">
              <label>Serv Host Cost/Month</label>
              <input
                v-model="formData.serv_host_cost_per_month"
                type="number"
                step="0.01"
                placeholder="0.00"
              />
            </div>
          </div>

          <div class="form-row">
            <div class="form-group">
              <label>Files Directory</label>
              <input
                v-model="formData.files_dir"
                type="text"
                placeholder="e.g., /var/www/files"
                @blur="formatUnixPath('files_dir')"
              />
              <small class="help-text">Unix path format (must start with /)</small>
            </div>
            <div class="form-group">
              <label>JS Directory</label>
              <input
                v-model="formData.js_dir"
                type="text"
                placeholder="e.g., /var/www/js"
                @blur="formatUnixPath('js_dir')"
              />
              <small class="help-text">Unix path format (must start with /)</small>
            </div>
          </div>

          <div class="modal-actions">
            <button type="button" @click="closeModal" class="btn-cancel">Cancel</button>
            <button type="submit" :disabled="saving" class="btn-save">
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
            Are you sure you want to delete database <strong>{{ databaseToDelete?.db_code }}</strong
            >?
          </p>
          <p class="warning-text">This action cannot be undone.</p>
        </div>
        <div class="modal-actions">
          <button @click="cancelDelete" class="btn-cancel">Cancel</button>
          <button @click="deleteDatabase" :disabled="deleting" class="btn-delete-confirm">
            <i v-if="deleting" class="fas fa-spinner fa-spin"></i>
            {{ deleting ? 'Deleting...' : 'Delete' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Version Input Modal -->
    <div v-if="showVersionModal" class="modal-overlay" @click="cancelVersionInput">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Enter Database Version</h3>
        </div>
        <div class="modal-body">
          <p>
            Please enter the version number for database <strong>{{ databaseToCreateTables?.db_code }}</strong>:
          </p>
          <div class="form-group">
            <label for="version-input">Version:</label>
            <input
              id="version-input"
              type="number"
              v-model="versionInput"
              placeholder="Enter version number"
              min="1"
              class="form-input"
              @keyup.enter="confirmCreateTables"
            />
          </div>
        </div>
        <div class="modal-actions">
          <button @click="cancelVersionInput" class="btn-cancel">Cancel</button>
          <button @click="confirmCreateTables" :disabled="!versionInput || creatingTables" class="btn-primary">
            <i v-if="creatingTables" class="fas fa-spinner fa-spin"></i>
            {{ creatingTables ? 'Creating...' : 'Create Tables' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'

const databases = ref([])
const loading = ref(false)
const error = ref('')
const successMessage = ref('')
const showModal = ref(false)
const showDeleteModal = ref(false)
const editingDatabase = ref(null)
const databaseToDelete = ref(null)
const saving = ref(false)
const deleting = ref(false)
const creatingTables = ref(false)
const creatingTablesId = ref(null)
const showVersionModal = ref(false)
const databaseToCreateTables = ref(null)
const versionInput = ref('')
const selectedDatabases = ref([])

const formData = ref({
  db_code: '',
  db_name: '',
  db_host_start: '',
  db_host_end: '',
  db_host_cost_per_month: '',
  serv_host_start: '',
  serv_host_end: '',
  serv_host_cost_per_month: '',
  files_dir: '',
  js_dir: '',
})

// Get API base URL
const getApiBaseUrl = () => {
  const hostname = window.location.hostname
  const isLocalhost = hostname === 'localhost' || hostname === '127.0.0.1'
  return isLocalhost ? 'http://localhost:8000/api' : 'https://www.merhab.com/api'
}

// Fetch all databases
const fetchDatabases = async () => {
  loading.value = true
  error.value = ''
  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php?action=get_databases`)
    const result = await response.json()

    if (result.success) {
      databases.value = result.data || []
    } else {
      error.value = result.message || 'Failed to fetch databases'
    }
  } catch (err) {
    error.value = 'An error occurred while fetching databases'
    console.error(err)
  } finally {
    loading.value = false
  }
}

// Open create modal
const openCreateModal = () => {
  editingDatabase.value = null
  resetForm()
  showModal.value = true
}

// Open edit modal
const openEditModal = (db) => {
  editingDatabase.value = db
  formData.value = {
    db_code: db.db_code || '',
    db_name: db.db_name || '',
    db_host_start: db.db_host_start ? db.db_host_start.split(' ')[0] : '',
    db_host_end: db.db_host_end ? db.db_host_end.split(' ')[0] : '',
    db_host_cost_per_month: db.db_host_cost_per_month || '',
    serv_host_start: db.serv_host_start ? db.serv_host_start.split(' ')[0] : '',
    serv_host_end: db.serv_host_end ? db.serv_host_end.split(' ')[0] : '',
    serv_host_cost_per_month: db.serv_host_cost_per_month || '',
    files_dir: db.files_dir || '',
    js_dir: db.js_dir || '',
  }
  showModal.value = true
}

// Close modal
const closeModal = () => {
  showModal.value = false
  editingDatabase.value = null
  resetForm()
}

// Reset form
const resetForm = () => {
  formData.value = {
    db_code: '',
    db_name: '',
    db_host_start: '',
    db_host_end: '',
    db_host_cost_per_month: '',
    serv_host_start: '',
    serv_host_end: '',
    serv_host_cost_per_month: '',
    files_dir: '',
    js_dir: '',
  }
}

// Save database (create or update)
const saveDatabase = async () => {
  saving.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const payload = {
      action: editingDatabase.value ? 'update_database' : 'create_database',
      ...formData.value,
    }

    if (editingDatabase.value) {
      payload.id = editingDatabase.value.id
    } else {
      // Remove db_code when creating (it will be auto-generated on backend)
      delete payload.db_code
    }

    // Convert empty strings to null for optional fields
    const cleanPayload = { ...payload }
    Object.keys(cleanPayload).forEach((key) => {
      if (cleanPayload[key] === '') {
        cleanPayload[key] = null
      }
    })

    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(cleanPayload),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = editingDatabase.value
        ? 'Database updated successfully'
        : 'Database created successfully'
      closeModal()
      await fetchDatabases()
      setTimeout(() => {
        successMessage.value = ''
      }, 3000)
    } else {
      error.value = result.message || 'Failed to save database'
    }
  } catch (err) {
    error.value = 'An error occurred while saving database'
    console.error(err)
  } finally {
    saving.value = false
  }
}

// Confirm delete
const confirmDelete = (db) => {
  databaseToDelete.value = db
  showDeleteModal.value = true
}

// Cancel delete
const cancelDelete = () => {
  showDeleteModal.value = false
  databaseToDelete.value = null
}

// Delete database
// Create tables from setup.sql
const createTables = (db) => {
  if (db.is_created == 1) {
    return
  }

  databaseToCreateTables.value = db
  versionInput.value = ''
  showVersionModal.value = true
}

const cancelVersionInput = () => {
  showVersionModal.value = false
  databaseToCreateTables.value = null
  versionInput.value = ''
}

const confirmCreateTables = async () => {
  if (!databaseToCreateTables.value || !versionInput.value || parseInt(versionInput.value) <= 0) {
    return
  }

  creatingTables.value = true
  creatingTablesId.value = databaseToCreateTables.value.id
  error.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'create_tables',
        id: databaseToCreateTables.value.id,
        version: parseInt(versionInput.value),
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = result.message || 'Tables created successfully'
      showVersionModal.value = false
      databaseToCreateTables.value = null
      versionInput.value = ''
      await fetchDatabases()
      setTimeout(() => {
        successMessage.value = ''
      }, 5000)
    } else {
      error.value = result.message || 'Failed to create tables'
    }
  } catch (err) {
    error.value = 'An error occurred while creating tables'
    console.error(err)
  } finally {
    creatingTables.value = false
    creatingTablesId.value = null
  }
}

const deleteDatabase = async () => {
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
        action: 'delete_database',
        id: databaseToDelete.value.id,
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = 'Database deleted successfully'
      cancelDelete()
      await fetchDatabases()
      setTimeout(() => {
        successMessage.value = ''
      }, 3000)
    } else {
      error.value = result.message || 'Failed to delete database'
    }
  } catch (err) {
    error.value = 'An error occurred while deleting database'
    console.error(err)
  } finally {
    deleting.value = false
  }
}

// Format date
const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

// Format currency
const formatCurrency = (amount) => {
  if (!amount && amount !== 0) return '-'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

// Select all functionality
const allSelected = computed(() => {
  return databases.value.length > 0 && selectedDatabases.value.length === databases.value.length
})

const toggleSelectAll = () => {
  if (allSelected.value) {
    selectedDatabases.value = []
  } else {
    selectedDatabases.value = databases.value.map(db => db.id)
  }
}

// Toolbar button handlers (to be implemented later)
const updateStructure = () => {
  console.log('Update Structure clicked for databases:', selectedDatabases.value)
  // Functionality will be implemented later
}

const runSql = () => {
  console.log('Run SQL clicked for databases:', selectedDatabases.value)
  // Functionality will be implemented later
}

// Format Unix path on blur
const formatUnixPath = (field) => {
  if (!formData.value[field]) return

  let path = formData.value[field]

  // Replace backslashes with forward slashes
  path = path.replace(/\\/g, '/')

  // Remove double slashes
  path = path.replace(/\/+/g, '/')

  // Ensure it starts with /
  if (!path.startsWith('/')) {
    path = '/' + path
  }

  // Remove trailing slash (unless it's root)
  if (path.length > 1 && path.endsWith('/')) {
    path = path.slice(0, -1)
  }

  formData.value[field] = path
}

onMounted(() => {
  fetchDatabases()
})
</script>

<style scoped>
.databases-component {
  position: relative;
  width: 100%;
  overflow-x: auto;
  padding: 0;
  z-index: 1;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.header-section h2 {
  margin: 0;
  color: #2c3e50;
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

.databases-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
}

.databases-table th {
  background-color: #f5f7fa;
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #2c3e50;
  border-bottom: 2px solid #e4e7ed;
}

.databases-table td {
  padding: 1rem;
  border-bottom: 1px solid #e4e7ed;
  color: #606266;
}

.databases-table tr:hover {
  background-color: #f5f7fa;
}

/* Checkbox Styles */
.checkbox-select-all,
.checkbox-db {
  width: 18px;
  height: 18px;
  cursor: pointer;
  accent-color: #409eff;
}

.checkbox-select-all {
  margin: 0;
}

.checkbox-db {
  margin: 0;
}

/* Toolbar Styles */
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  background-color: #f5f7fa;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  margin-bottom: 1rem;
  position: relative;
  z-index: 0;
  width: 100%;
  box-sizing: border-box;
}

.toolbar-left {
  display: flex;
  gap: 0.75rem;
}

.toolbar-right {
  display: flex;
  align-items: center;
}

.btn-toolbar {
  background-color: #409eff;
  color: white;
  border: none;
  padding: 0.625rem 1.25rem;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.3s ease;
}

.btn-toolbar:hover:not(:disabled) {
  background-color: #66b1ff;
  transform: translateY(-1px);
}

.btn-toolbar:disabled {
  background-color: #c0c4cc;
  cursor: not-allowed;
  opacity: 0.6;
  pointer-events: none;
}

.btn-toolbar i {
  font-size: 0.875rem;
}

.selected-count {
  color: #606266;
  font-size: 0.9rem;
  font-weight: 500;
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.btn-create-tables,
.btn-edit,
.btn-delete {
  padding: 0.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  white-space: nowrap;
}

.btn-create-tables {
  background-color: #67c23a;
  color: white;
  padding: 0.5rem 0.75rem;
  min-width: 120px;
}

.btn-create-tables:hover:not(:disabled) {
  background-color: #85ce61;
}

.btn-create-tables:disabled {
  background-color: #c0c4cc;
  cursor: not-allowed;
  opacity: 0.6;
}

.btn-edit,
.btn-delete {
  width: 32px;
  height: 32px;
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

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.version-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
  background-color: #409eff;
  color: white;
}

.version-badge.version-unknown {
  background-color: #909399;
  color: white;
}

.version-badge.version-pending {
  background-color: #e4e7ed;
  color: #909399;
  text-transform: uppercase;
}

.status-created {
  background-color: #f0f9eb;
  color: #67c23a;
}

.status-pending {
  background-color: #fef0f0;
  color: #f56c6c;
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
  cursor: pointer;
  color: #909399;
  padding: 0.5rem;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.modal-close:hover {
  background-color: #f5f7fa;
  color: #606266;
}

.modal-form {
  padding: 1.5rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
}

.form-group label {
  margin-bottom: 0.5rem;
  color: #606266;
  font-weight: 500;
}

.form-group input {
  padding: 0.75rem;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.form-group input:focus {
  outline: none;
  border-color: #409eff;
  box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
}

.form-group input.readonly-input,
.form-group input:disabled {
  background-color: #f5f7fa;
  color: #909399;
  cursor: not-allowed;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.help-text {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: #909399;
  font-style: italic;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e4e7ed;
}

.btn-cancel,
.btn-save {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.3s ease;
}

.btn-cancel {
  background-color: #909399;
  color: white;
}

.btn-cancel:hover {
  background-color: #a6a9ad;
}

.btn-save {
  background-color: #409eff;
  color: white;
}

.btn-save:hover:not(:disabled) {
  background-color: #66b1ff;
}

.btn-save:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.delete-modal {
  max-width: 500px;
}

.modal-body {
  padding: 1.5rem;
}

.modal-body p {
  margin: 0.5rem 0;
  color: #606266;
}

.warning-text {
  color: #f56c6c;
  font-weight: 500;
}

.btn-delete-confirm {
  background-color: #f56c6c;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  transition: all 0.3s ease;
}

.btn-delete-confirm:hover:not(:disabled) {
  background-color: #f78989;
}

.btn-delete-confirm:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .form-row {
    grid-template-columns: 1fr;
  }

  .header-section {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .table-container {
    overflow-x: scroll;
  }
}
</style>
