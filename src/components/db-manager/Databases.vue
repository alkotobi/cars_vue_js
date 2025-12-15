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
        <button @click="openUpdateVersionModal" class="btn-toolbar" :disabled="selectedDatabases.length === 0">
          <i class="fas fa-tag"></i>
          Update Version
        </button>
        <button @click="openUploadCodeModal" class="btn-toolbar" :disabled="selectedDatabases.length === 0">
          <i class="fas fa-upload"></i>
          Upload Code Files
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
                :title="db.is_created == 1 ? 'Already created' : 'Create tables, directories, and db_code.json'"
              >
                <i class="fas fa-database"></i>
                {{ creatingTables && creatingTablesId === db.id ? 'Creating...' : 'Create' }}
              </button>
              <button @click="openEditModal(db)" class="btn-edit" title="Edit">
                <i class="fas fa-edit"></i>
              </button>
              <button @click="openJsonModal(db)" class="btn-json" title="Edit db_code.json">
                <i class="fas fa-code"></i>
                JSON
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
            {{ creatingTables ? 'Creating...' : 'Create' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Update Version Modal -->
    <div v-if="showUpdateVersionModal" class="modal-overlay" @click="cancelUpdateVersion">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Update Version</h3>
        </div>
        <div class="modal-body">
          <div v-if="versionCheckError" class="error-message" style="margin-bottom: 1rem;">
            <i class="fas fa-exclamation-circle"></i>
            {{ versionCheckError }}
          </div>
          <div v-else>
            <p>
              Updating version for <strong>{{ selectedDatabases.length }}</strong> selected database(s).
            </p>
            <p v-if="selectedDatabasesVersions.length > 0" class="info-text">
              Current version: <strong>v{{ selectedDatabasesVersions[0] }}</strong>
            </p>
          </div>
          <div class="form-group">
            <label for="new-version-input">New Version *</label>
            <input
              id="new-version-input"
              type="number"
              v-model="newVersionInput"
              placeholder="Enter new version number"
              min="1"
              class="form-input"
              :disabled="!!versionCheckError"
              @keyup.enter="confirmUpdateVersion"
            />
          </div>
        </div>
        <div class="modal-actions">
          <button @click="cancelUpdateVersion" class="btn-cancel">Cancel</button>
          <button @click="confirmUpdateVersion" :disabled="!newVersionInput || updatingVersion || !!versionCheckError" class="btn-primary">
            <i v-if="updatingVersion" class="fas fa-spinner fa-spin"></i>
            {{ updatingVersion ? 'Updating...' : 'Update Version' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Run SQL Modal -->
    <div v-if="showRunSqlModal" class="modal-overlay" @click="cancelRunSql">
      <div class="modal-content run-sql-modal" @click.stop>
        <div class="modal-header">
          <h3>Run SQL</h3>
        </div>
        <div class="modal-body">
          <p>
            Execute SQL on <strong>{{ selectedDatabases.length }}</strong> selected database(s).
          </p>
          <div class="form-group">
            <label for="sql-input">SQL Statement *</label>
            <textarea
              id="sql-input"
              v-model="runSqlInput"
              required
              rows="10"
              class="form-input sql-textarea"
              placeholder="Enter SQL statement to execute"
              :disabled="runningSql"
            ></textarea>
          </div>
          
          <!-- Results -->
          <div v-if="runSqlResults.length > 0" class="sql-results">
            <h4>Results:</h4>
            <div v-for="(result, index) in runSqlResults" :key="index" class="result-item" :class="{ 'result-error': result.error, 'result-success': !result.error }">
              <div class="result-header">
                <strong>{{ result.db_name }}</strong>
                <span v-if="result.error" class="result-status error">Error</span>
                <span v-else class="result-status success">Success</span>
              </div>
              <div v-if="result.error" class="result-message">
                {{ result.error }}
              </div>
              <div v-else-if="result.affected_rows !== undefined" class="result-message">
                Affected rows: {{ result.affected_rows }}
              </div>
              <div v-else-if="result.rows" class="result-message">
                Rows returned: {{ result.rows.length }}
              </div>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="cancelRunSql" class="btn-cancel" :disabled="runningSql">Cancel</button>
          <button @click="confirmRunSql" :disabled="!runSqlInput.trim() || runningSql" class="btn-primary">
            <i v-if="runningSql" class="fas fa-spinner fa-spin"></i>
            {{ runningSql ? 'Running...' : 'Run SQL' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Edit db_code.json Modal -->
    <EditDbCodeJson
      :show="showJsonModal"
      :database="jsonEditingDatabase"
      :api-base-url="getApiBaseUrl()"
      @close="cancelJsonEdit"
      @saved="handleJsonSaved"
    />

    <!-- Upload Code Files Modal -->
    <div v-if="showUploadCodeModal" class="modal-overlay" @click="cancelUploadCode">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Upload Code Files</h3>
        </div>
        <div class="modal-body">
          <p>
            Upload code file(s) to <strong>{{ selectedDatabases.length }}</strong> selected database(s).
          </p>
          <div class="form-group">
            <label for="code-file-input">Select File(s) *</label>
            <div
              class="file-drop-zone"
              :class="{ 'drag-over': isDragOver, 'has-files': selectedCodeFiles.length > 0 }"
              @drop.prevent="handleFileDrop"
              @dragover.prevent="isDragOver = true"
              @dragenter.prevent="isDragOver = true"
              @dragleave.prevent="isDragOver = false"
            >
              <input
                id="code-file-input"
                type="file"
                ref="codeFileInput"
                @change="handleFileSelect"
                multiple
                :disabled="uploadingCode"
                class="file-input-hidden"
              />
              <div class="drop-zone-content">
                <i class="fas fa-cloud-upload-alt"></i>
                <p v-if="selectedCodeFiles.length === 0">
                  <strong>Drag and drop files here</strong><br />
                  or <span class="browse-link">browse</span> to select files
                </p>
                <p v-else>
                  <strong>{{ selectedCodeFiles.length }} file(s) selected</strong><br />
                  <span class="browse-link">Click to change files</span>
                </p>
              </div>
            </div>
            <div v-if="selectedCodeFiles.length > 0" class="selected-files">
              <p><strong>Selected files:</strong></p>
              <ul>
                <li v-for="(file, index) in selectedCodeFiles" :key="index">
                  {{ file.name }} ({{ formatFileSize(file.size) }})
                </li>
              </ul>
            </div>
          </div>
          
          <!-- Upload Results -->
          <div v-if="uploadCodeResults.length > 0" class="upload-results">
            <h4>Progress:</h4>
            <div v-for="(result, index) in uploadCodeResults" :key="index" class="result-item" :class="{ 'result-error': result.error, 'result-success': !result.error && result.progress === 100 }">
              <div class="result-header">
                <strong>{{ result.db_name }}</strong>
                <span v-if="result.error" class="result-status error">Error</span>
                <span v-else-if="result.progress === 100" class="result-status success">Completed</span>
                <span v-else class="result-status progress">In Progress</span>
              </div>
              
              <!-- Progress Bar -->
              <div v-if="!result.error" class="progress-container">
                <div class="progress-bar">
                  <div 
                    class="progress-fill" 
                    :style="{ width: (result.progress || 0) + '%' }"
                  ></div>
                </div>
                <span class="progress-text">{{ result.progress || 0 }}%</span>
              </div>
              
              <!-- Status Message -->
              <div v-if="result.status" class="result-status-message">
                {{ result.status }}
              </div>
              
              <!-- Error Message -->
              <div v-if="result.error" class="result-message">
                {{ result.error }}
              </div>
              
              <!-- Success Message with Files -->
              <div v-else-if="result.files && result.files.length > 0" class="result-message">
                <p>Uploaded {{ result.files.length }} file(s):</p>
                <ul>
                  <li v-for="(file, fileIndex) in result.files" :key="fileIndex">
                    {{ file.name }} → {{ file.path }}
                  </li>
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div class="modal-actions">
          <button @click="cancelUploadCode" class="btn-cancel" :disabled="uploadingCode">Cancel</button>
          <button @click="confirmUploadCode" :disabled="selectedCodeFiles.length === 0 || uploadingCode" class="btn-primary">
            <i v-if="uploadingCode" class="fas fa-spinner fa-spin"></i>
            {{ uploadingCode ? 'Uploading...' : 'Upload' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import EditDbCodeJson from './EditDbCodeJson.vue'

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
const showUpdateVersionModal = ref(false)
const newVersionInput = ref('')
const updatingVersion = ref(false)
const versionCheckError = ref('')
const selectedDatabasesVersions = ref([])
const showRunSqlModal = ref(false)
const runSqlInput = ref('')
const runningSql = ref(false)
const runSqlResults = ref([])
const showUploadCodeModal = ref(false)
const codeFileInput = ref(null)
const selectedCodeFiles = ref([])
const uploadingCode = ref(false)
const uploadCodeResults = ref([])
const isDragOver = ref(false)
const showJsonModal = ref(false)
const jsonEditingDatabase = ref(null)

// JSON file editing functionality
const openJsonModal = (db) => {
  if (!db.js_dir || db.js_dir.trim() === '') {
    error.value = 'JS directory (js_dir) is not configured for this database'
    return
  }
  
  jsonEditingDatabase.value = db
  showJsonModal.value = true
}

const cancelJsonEdit = () => {
  showJsonModal.value = false
  jsonEditingDatabase.value = null
}

const handleJsonSaved = (message) => {
  successMessage.value = message
  setTimeout(() => {
    successMessage.value = ''
  }, 5000)
}

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
// Create tables, directories, and db_code.json from setup.sql
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

// Toolbar button handlers
const updateStructure = async () => {
  if (selectedDatabases.value.length === 0) {
    error.value = 'Please select at least one database'
    return
  }

  // Confirm action
  if (!confirm(`This will update the structure for ${selectedDatabases.value.length} selected database(s). Continue?`)) {
    return
  }

  error.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'update_structure',
        database_ids: selectedDatabases.value,
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = result.message || 'Structure updated successfully'
      await fetchDatabases()
      setTimeout(() => {
        successMessage.value = ''
      }, 5000)
    } else {
      error.value = result.message || 'Failed to update structure'
    }
  } catch (err) {
    error.value = 'An error occurred while updating structure'
    console.error(err)
  }
}

const runSql = () => {
  if (selectedDatabases.value.length === 0) {
    error.value = 'Please select at least one database'
    return
  }
  
  runSqlInput.value = ''
  runSqlResults.value = []
  showRunSqlModal.value = true
}

// Update version functionality
const openUpdateVersionModal = () => {
  if (selectedDatabases.value.length === 0) {
    return
  }

  // Get selected databases with their versions
  const selectedDbs = databases.value.filter(db => selectedDatabases.value.includes(db.id))
  
  // Check if all selected databases are created
  const notCreated = selectedDbs.filter(db => db.is_created != 1)
  if (notCreated.length > 0) {
    error.value = 'All selected databases must be created before updating version'
    return
  }

  // Get versions of selected databases
  const versions = selectedDbs
    .map(db => db.version)
    .filter(v => v !== null && v !== undefined)

  if (versions.length === 0) {
    versionCheckError.value = 'Selected databases do not have version information'
    selectedDatabasesVersions.value = []
  } else {
    // Check if all versions are the same
    const uniqueVersions = [...new Set(versions)]
    if (uniqueVersions.length > 1) {
      versionCheckError.value = `Selected databases have different versions: ${uniqueVersions.join(', ')}. All databases must have the same version to update.`
      selectedDatabasesVersions.value = []
    } else {
      versionCheckError.value = ''
      selectedDatabasesVersions.value = uniqueVersions
    }
  }

  newVersionInput.value = ''
  showUpdateVersionModal.value = true
}

const cancelUpdateVersion = () => {
  showUpdateVersionModal.value = false
  newVersionInput.value = ''
  versionCheckError.value = ''
  selectedDatabasesVersions.value = []
}

// Run SQL functionality
const cancelRunSql = () => {
  showRunSqlModal.value = false
  runSqlInput.value = ''
  runSqlResults.value = []
}

const confirmRunSql = async () => {
  if (!runSqlInput.value.trim()) {
    return
  }

  runningSql.value = true
  error.value = ''
  successMessage.value = ''
  runSqlResults.value = []

  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'run_sql',
        database_ids: selectedDatabases.value,
        sql: runSqlInput.value.trim(),
      }),
    })

    const result = await response.json()

    if (result.success) {
      runSqlResults.value = result.data || []
      successMessage.value = result.message || 'SQL executed successfully'
      setTimeout(() => {
        successMessage.value = ''
      }, 5000)
    } else {
      error.value = result.message || 'Failed to execute SQL'
    }
  } catch (err) {
    error.value = 'An error occurred while executing SQL'
    console.error(err)
  } finally {
    runningSql.value = false
  }
}

const confirmUpdateVersion = async () => {
  if (!newVersionInput.value || parseInt(newVersionInput.value) <= 0) {
    return
  }

  if (versionCheckError.value) {
    return
  }

  updatingVersion.value = true
  error.value = ''
  successMessage.value = ''

  try {
    const response = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        action: 'update_databases_version',
        database_ids: selectedDatabases.value,
        version: parseInt(newVersionInput.value),
      }),
    })

    const result = await response.json()

    if (result.success) {
      successMessage.value = result.message || `Version updated successfully for ${selectedDatabases.value.length} database(s)`
      showUpdateVersionModal.value = false
      newVersionInput.value = ''
      versionCheckError.value = ''
      selectedDatabasesVersions.value = []
      await fetchDatabases()
      setTimeout(() => {
        successMessage.value = ''
      }, 5000)
    } else {
      error.value = result.message || 'Failed to update version'
    }
  } catch (err) {
    error.value = 'An error occurred while updating version'
    console.error(err)
  } finally {
    updatingVersion.value = false
  }
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

// Upload Code Files functionality
const openUploadCodeModal = () => {
  if (selectedDatabases.value.length === 0) {
    error.value = 'Please select at least one database'
    return
  }
  
  // Check if all selected databases have js_dir
  const selectedDbs = databases.value.filter(db => selectedDatabases.value.includes(db.id))
  const withoutJsDir = selectedDbs.filter(db => !db.js_dir || db.js_dir.trim() === '')
  if (withoutJsDir.length > 0) {
    error.value = 'All selected databases must have a JS directory (js_dir) configured'
    return
  }
  
  selectedCodeFiles.value = []
  uploadCodeResults.value = []
  isDragOver.value = false
  showUploadCodeModal.value = true
}

const handleFileSelect = (event) => {
  const files = Array.from(event.target.files || [])
  selectedCodeFiles.value = files
  isDragOver.value = false
}

const handleFileDrop = (event) => {
  isDragOver.value = false
  const files = Array.from(event.dataTransfer.files || [])
  if (files.length > 0) {
    selectedCodeFiles.value = files
    // Also update the file input element
    if (codeFileInput.value) {
      const dataTransfer = new DataTransfer()
      files.forEach(file => dataTransfer.items.add(file))
      codeFileInput.value.files = dataTransfer.files
    }
  }
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 Bytes'
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const cancelUploadCode = () => {
  showUploadCodeModal.value = false
  selectedCodeFiles.value = []
  uploadCodeResults.value = []
  isDragOver.value = false
  if (codeFileInput.value) {
    codeFileInput.value.value = ''
  }
}

const confirmUploadCode = async () => {
  if (selectedCodeFiles.value.length === 0) {
    error.value = 'Please select at least one file'
    return
  }
  
  uploadingCode.value = true
  uploadCodeResults.value = []
  error.value = ''
  
  try {
    const selectedDbs = databases.value.filter(db => selectedDatabases.value.includes(db.id))
    
    // Group databases by js_dir to avoid duplicate uploads
    const dbGroupsByJsDir = {}
    selectedDbs.forEach(db => {
      const jsDir = db.js_dir.trim()
      if (!dbGroupsByJsDir[jsDir]) {
        dbGroupsByJsDir[jsDir] = []
      }
      dbGroupsByJsDir[jsDir].push(db)
    })
    
    // Initialize results for all databases
    selectedDbs.forEach(db => {
      uploadCodeResults.value.push({
        db_name: db.db_name,
        error: null,
        files: [],
        progress: 0,
        status: 'preparing',
        js_dir: db.js_dir.trim()
      })
    })
    
    // Process each unique js_dir
    for (const [jsDir, dbGroup] of Object.entries(dbGroupsByJsDir)) {
      // Find result indices for all databases in this group
      const resultIndices = dbGroup.map(db => 
        uploadCodeResults.value.findIndex(r => r.db_name === db.db_name)
      )
      
      // Use the first database in the group for the API call
      const firstDb = dbGroup[0]
      
      try {
        // Step 1: Prepare folder (create if doesn't exist, clear if exists)
        resultIndices.forEach(idx => {
          uploadCodeResults.value[idx].status = 'Preparing folder...'
          uploadCodeResults.value[idx].progress = 0
        })
        
        const prepareResponse = await fetch(`${getApiBaseUrl()}/db_manager_api.php`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            action: 'prepare_upload_folder',
            database_id: firstDb.id,
            js_dir: jsDir,
          }),
        })
        
        const prepareResult = await prepareResponse.json()
        
        if (!prepareResult.success) {
          const errorMsg = prepareResult.message || 'Failed to prepare folder'
          resultIndices.forEach(idx => {
            uploadCodeResults.value[idx].error = errorMsg
            uploadCodeResults.value[idx].status = 'Error'
          })
          continue
        }
        
        // Step 2: Upload files with progress (only once per js_dir)
        resultIndices.forEach(idx => {
          uploadCodeResults.value[idx].status = 'Uploading files...'
        })
        const totalFiles = selectedCodeFiles.value.length
        const uploadedFiles = []
        
        for (let fileIndex = 0; fileIndex < selectedCodeFiles.value.length; fileIndex++) {
          const file = selectedCodeFiles.value[fileIndex]
          
          // Update progress for all databases in this group
          const progress = Math.round(((fileIndex + 1) / totalFiles) * 100)
          resultIndices.forEach(idx => {
            uploadCodeResults.value[idx].progress = progress
            uploadCodeResults.value[idx].status = `Uploading ${fileIndex + 1}/${totalFiles}: ${file.name}`
          })
          
          const formData = new FormData()
          formData.append('file', file)
          
          // Use js_dir as base_directory, remove leading slash if present
          let jsDirPath = jsDir
          if (jsDirPath.startsWith('/')) {
            jsDirPath = jsDirPath.substring(1)
          }
          formData.append('base_directory', jsDirPath)
          formData.append('destination_folder', '') // Upload to root of js_dir
          formData.append('custom_filename', file.name) // Keep original filename
          
          const uploadResponse = await fetch(`${getApiBaseUrl()}/upload.php`, {
            method: 'POST',
            body: formData,
          })
          
          const uploadResult = await uploadResponse.json()
          
          if (uploadResult.success) {
            uploadedFiles.push({
              name: file.name,
              path: uploadResult.file_path || 'Uploaded successfully'
            })
          } else {
            const errorMsg = uploadResult.message || 'Upload failed'
            resultIndices.forEach(idx => {
              uploadCodeResults.value[idx].error = errorMsg
              uploadCodeResults.value[idx].status = 'Error'
            })
            break // Stop uploading other files for this js_dir if one fails
          }
        }
        
        // Mark all databases in this group as complete
        if (uploadedFiles.length > 0 && !resultIndices.some(idx => uploadCodeResults.value[idx].error)) {
          resultIndices.forEach(idx => {
            uploadCodeResults.value[idx].files = uploadedFiles
            uploadCodeResults.value[idx].status = 'Completed'
            uploadCodeResults.value[idx].progress = 100
          })
        }
      } catch (err) {
        const errorMsg = err.message || 'An error occurred during upload'
        resultIndices.forEach(idx => {
          uploadCodeResults.value[idx].error = errorMsg
          uploadCodeResults.value[idx].status = 'Error'
        })
      }
    }
    
    // Check if all uploads were successful
    const allSuccessful = uploadCodeResults.value.every(r => !r.error && r.files.length > 0)
    
    if (allSuccessful) {
      successMessage.value = 'All files uploaded successfully'
      setTimeout(() => {
        successMessage.value = ''
        // Close the modal after successful upload
        cancelUploadCode()
      }, 2000)
    }
  } catch (err) {
    error.value = 'An error occurred while uploading files'
    console.error(err)
  } finally {
    uploadingCode.value = false
  }
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
.btn-delete,
.btn-json {
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

.btn-json {
  background-color: #909399;
  color: white;
  font-size: 0.7rem;
  padding: 0.25rem;
  min-width: 50px;
  width: auto;
  height: 32px;
}

.btn-json:hover {
  background-color: #a6a9ad;
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

.info-text {
  color: #606266;
  font-size: 0.9rem;
  margin-top: 0.5rem;
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

.run-sql-modal {
  max-width: 900px;
}

.sql-textarea {
  font-family: 'Courier New', monospace;
  font-size: 0.9rem;
  resize: vertical;
}

.sql-results {
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e4e7ed;
}

.sql-results h4 {
  margin: 0 0 1rem 0;
  color: #2c3e50;
  font-size: 1rem;
}

.result-item {
  margin-bottom: 1rem;
  padding: 1rem;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.result-item.result-success {
  background-color: #f0f9eb;
  border-color: #e1f3d8;
}

.result-item.result-error {
  background-color: #fef0f0;
  border-color: #fde2e2;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.result-header strong {
  color: #2c3e50;
}

.result-status {
  padding: 0.25rem 0.75rem;
  border-radius: 12px;
  font-size: 0.75rem;
  font-weight: 600;
}

.result-status.success {
  background-color: #67c23a;
  color: white;
}

.result-status.error {
  background-color: #f56c6c;
  color: white;
}

.result-message {
  color: #606266;
  font-size: 0.9rem;
  font-family: 'Courier New', monospace;
  word-break: break-word;
}

.upload-results {
  margin-top: 1.5rem;
}

.upload-results h4 {
  margin-bottom: 1rem;
  color: #2c3e50;
  font-size: 1.1rem;
}

.selected-files {
  margin-top: 1rem;
  padding: 1rem;
  background-color: #f5f7fa;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.selected-files p {
  margin: 0 0 0.5rem 0;
  font-weight: 500;
  color: #2c3e50;
}

.selected-files ul {
  margin: 0;
  padding-left: 1.5rem;
  list-style-type: disc;
}

.selected-files li {
  margin: 0.25rem 0;
  color: #606266;
}

.result-message ul {
  margin: 0.5rem 0;
  padding-left: 1.5rem;
  list-style-type: disc;
}

.result-message li {
  margin: 0.25rem 0;
}

.file-drop-zone {
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  padding: 3rem 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.3s ease;
  background-color: #fafafa;
  position: relative;
}

.file-drop-zone:hover {
  border-color: #409eff;
  background-color: #f0f7ff;
}

.file-drop-zone.drag-over {
  border-color: #409eff;
  background-color: #e6f3ff;
  transform: scale(1.02);
}

.file-drop-zone.has-files {
  border-color: #67c23a;
  background-color: #f0f9eb;
}

.file-input-hidden {
  position: absolute;
  width: 100%;
  height: 100%;
  top: 0;
  left: 0;
  opacity: 0;
  cursor: pointer;
  z-index: 1;
}

.drop-zone-content {
  pointer-events: none;
  z-index: 0;
}

.drop-zone-content i {
  font-size: 3rem;
  color: #909399;
  margin-bottom: 1rem;
  display: block;
}

.file-drop-zone.drag-over .drop-zone-content i,
.file-drop-zone.has-files .drop-zone-content i {
  color: #409eff;
}

.drop-zone-content p {
  margin: 0;
  color: #606266;
  font-size: 1rem;
}

.browse-link {
  color: #409eff;
  text-decoration: underline;
  cursor: pointer;
}

.browse-link:hover {
  color: #66b1ff;
}

.progress-container {
  margin: 1rem 0;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.progress-bar {
  flex: 1;
  height: 24px;
  background-color: #e4e7ed;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #409eff 0%, #66b1ff 100%);
  border-radius: 12px;
  transition: width 0.3s ease;
  position: relative;
}

.progress-fill::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0;
  right: 0;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}

.progress-text {
  font-weight: 600;
  color: #2c3e50;
  min-width: 50px;
  text-align: right;
  font-size: 0.9rem;
}

.result-status.progress {
  background-color: #409eff;
  color: white;
}

.result-status-message {
  margin-top: 0.5rem;
  color: #606266;
  font-size: 0.9rem;
  font-style: italic;
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
