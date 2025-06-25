<script setup>
import { ref, onMounted, computed } from 'vue'
import { useApi } from '../composables/useApi'

const clients = ref([])
const { callApi, uploadFile, getFileUrl, error } = useApi()
const showAddDialog = ref(false)
const showEditDialog = ref(false)
const editingClient = ref(null)
const user = ref(null)
const validationError = ref('')
const selectedFile = ref(null)
const editSelectedFile = ref(null)
const isSubmitting = ref(false)
const isLoading = ref(false)

// Check if user is admin by getting role from localStorage
const isAdmin = computed(() => user.value?.role_id === 1)

// Add helper function to check if file is an image
const isImageFile = (path) => {
  if (!path) return false
  const extension = path.split('.').pop().toLowerCase()
  return ['jpg', 'jpeg', 'png', 'gif', 'webp'].includes(extension)
}

const validateEmail = (email) => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  return emailRegex.test(email)
}

const newClient = ref({
  name: '',
  address: '',
  email: '',
  mobiles: '',
  id_no: '',
  is_client: true,
  is_broker: false,
  notes: '',
})

// Add filter and sort refs
const filters = ref({
  name: '',
  email: '',
  mobile: '',
  idNo: '',
  isBroker: null,
})

const sortConfig = ref({
  key: 'name',
  direction: 'asc',
})

// Add computed property for filtered and sorted clients
const filteredAndSortedClients = computed(() => {
  let result = [...clients.value]

  // Apply filters
  result = result.filter((client) => {
    if (
      filters.value.name &&
      !client.name?.toLowerCase().includes(filters.value.name.toLowerCase())
    ) {
      return false
    }
    if (
      filters.value.email &&
      !client.email?.toLowerCase().includes(filters.value.email.toLowerCase())
    ) {
      return false
    }
    if (filters.value.mobile && !client.mobiles?.includes(filters.value.mobile)) {
      return false
    }
    if (filters.value.idNo && !client.id_no?.includes(filters.value.idNo)) {
      return false
    }
    if (filters.value.isBroker !== null && client.is_broker !== filters.value.isBroker) {
      return false
    }
    return true
  })

  // Apply sorting
  result.sort((a, b) => {
    let aVal = a[sortConfig.value.key]
    let bVal = b[sortConfig.value.key]

    // Handle null values
    if (aVal === null) aVal = ''
    if (bVal === null) bVal = ''

    // Convert to strings for comparison
    aVal = String(aVal).toLowerCase()
    bVal = String(bVal).toLowerCase()

    if (sortConfig.value.direction === 'asc') {
      return aVal > bVal ? 1 : aVal < bVal ? -1 : 0
    } else {
      return aVal < bVal ? 1 : aVal > bVal ? -1 : 0
    }
  })

  return result
})

// Add function to handle sorting
const handleSort = (key) => {
  if (sortConfig.value.key === key) {
    // If already sorting by this key, toggle direction
    sortConfig.value.direction = sortConfig.value.direction === 'asc' ? 'desc' : 'asc'
  } else {
    // If sorting by a new key, set it with ascending direction
    sortConfig.value.key = key
    sortConfig.value.direction = 'asc'
  }
}

// Add function to clear filters
const clearFilters = () => {
  filters.value = {
    name: '',
    email: '',
    mobile: '',
    idNo: '',
    isBroker: null,
  }
}

const fetchClients = async () => {
  isLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT 
          c.*,
          COUNT(cs.id) as cars_count
        FROM clients c
        LEFT JOIN cars_stock cs ON c.id = cs.id_client
        WHERE c.is_client = 1
        GROUP BY c.id
        ORDER BY c.name ASC
    `,
      params: [],
    })
    if (result.success) {
      clients.value = result.data
    }
  } catch (err) {
    console.error('Error fetching clients:', err)
    error.value = 'Failed to load clients'
  } finally {
    isLoading.value = false
  }
}

const handleFileChange = (event, isEdit = false) => {
  const file = event.target.files[0]
  if (isEdit) {
    editSelectedFile.value = file
  } else {
    selectedFile.value = file
  }
}

const addClient = async () => {
  if (isSubmitting.value) return // Prevent double submission

  // Clear previous validation errors
  validationError.value = ''

  // Required field validation
  if (!newClient.value.mobiles) {
    validationError.value = 'Mobile number is required'
    return
  }

  if (!newClient.value.id_no) {
    validationError.value = 'ID number is required'
    return
  }

  if (!selectedFile.value) {
    validationError.value = 'ID Document is required'
    return
  }

  // Email validation only if provided
  if (newClient.value.email && !validateEmail(newClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  try {
    isSubmitting.value = true
    // First insert the client to get the ID
    const result = await callApi({
      query: `
        INSERT INTO clients (name, address, email, mobiles, id_no, is_broker, is_client, notes)
        VALUES (?, ?, ?, ?, ?, ?, 1, ?)
    `,
      params: [
        newClient.value.name,
        newClient.value.address,
        newClient.value.email,
        newClient.value.mobiles,
        newClient.value.id_no,
        newClient.value.is_broker ? 1 : 0,
        newClient.value.notes,
      ],
    })

    console.log('Insert result:', result)
    console.log('Insert result type:', typeof result)
    console.log('Full result object:', JSON.stringify(result, null, 2))

    if (result.success) {
      const clientId = result.lastInsertId
      console.log('Client ID ***************:', clientId)
      // If there's a file selected, upload it
      if (selectedFile.value) {
        try {
          const filename = `${clientId}.${selectedFile.value.name.split('.').pop()}`
          console.log('Uploading file with filename:', filename)
          const uploadResult = await uploadFile(selectedFile.value, 'ids', filename)
          console.log('Upload result:', uploadResult)

          if (uploadResult.success) {
            // Use the relativePath returned from uploadFile
            console.log('Updating client with file path:', uploadResult.relativePath)
            const updateResult = await callApi({
              query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
              params: [uploadResult.relativePath, clientId],
            })
            console.log('Update result:', updateResult)
          }
        } catch (err) {
          console.error('Error uploading file:', err)
          error.value = 'Client created but failed to upload ID document'
        }
      }

      // Reset form and close dialog
      showAddDialog.value = false
      validationError.value = ''
      selectedFile.value = null
      newClient.value = {
        name: '',
        address: '',
        email: '',
        mobiles: '',
        id_no: '',
        is_client: true,
        is_broker: false,
        notes: '',
      }
      await fetchClients()
    } else {
      error.value = result.error
      console.error('Error adding client:', result.error)
    }
  } catch (err) {
    error.value = err.message
    console.error('Error in add client process:', err)
  } finally {
    isSubmitting.value = false
  }
}

const editClient = (client) => {
  editingClient.value = { ...client }
  editSelectedFile.value = null
  showEditDialog.value = true
}

const updateClient = async () => {
  if (isSubmitting.value) return // Prevent double submission

  // Clear previous validation errors
  validationError.value = ''

  // Required field validation
  if (!editingClient.value.mobiles) {
    validationError.value = 'Mobile number is required'
    return
  }

  if (!editingClient.value.id_no) {
    validationError.value = 'ID number is required'
    return
  }

  // Check if there's either an existing file or a new file selected
  if (!editingClient.value.id_copy_path && !editSelectedFile.value) {
    validationError.value = 'ID Document is required'
    return
  }

  // Email validation only if provided
  if (editingClient.value.email && !validateEmail(editingClient.value.email)) {
    validationError.value = 'Please enter a valid email address'
    return
  }

  try {
    isSubmitting.value = true

    console.log('=== UPDATE CLIENT DEBUG ===')
    console.log('Client ID:', editingClient.value.id)
    console.log('Current file path:', editingClient.value.id_copy_path)
    console.log('New file selected:', editSelectedFile.value)

    // If there's a new file selected, upload it first
    if (editSelectedFile.value) {
      try {
        const filename = `${editingClient.value.id}.${editSelectedFile.value.name.split('.').pop()}`
        console.log('Uploading new file with filename:', filename)

        const uploadResult = await uploadFile(editSelectedFile.value, 'ids', filename)
        console.log('Upload result:', uploadResult)

        if (uploadResult.success) {
          // Update the client with the new file path
          editingClient.value.id_copy_path = uploadResult.relativePath
          console.log('Updated client file path to:', editingClient.value.id_copy_path)
        } else {
          console.error('Upload failed:', uploadResult)
          error.value = 'Failed to upload new ID document'
          return
        }
      } catch (err) {
        console.error('Error uploading file:', err)
        error.value = 'Failed to upload new ID document'
        return
      }
    }

    console.log('Final file path for database update:', editingClient.value.id_copy_path)

    // Update client basic info (including the file path if it was updated)
    const result = await callApi({
      query: `
      UPDATE clients 
      SET name = ?, address = ?, email = ?, mobiles = ?, id_no = ?, is_broker = ?, is_client = 1, notes = ?, id_copy_path = ?
      WHERE id = ?
    `,
      params: [
        editingClient.value.name,
        editingClient.value.address,
        editingClient.value.email,
        editingClient.value.mobiles,
        editingClient.value.id_no,
        editingClient.value.is_broker ? 1 : 0,
        editingClient.value.notes,
        editingClient.value.id_copy_path,
        editingClient.value.id,
      ],
    })

    console.log('Database update result:', result)

    if (result.success) {
      console.log('Client updated successfully')
      showEditDialog.value = false
      editingClient.value = null
      editSelectedFile.value = null
      validationError.value = ''
      await fetchClients()
    } else {
      console.error('Database update failed:', result.error)
      error.value = result.error
    }
  } catch (err) {
    console.error('Error in update client process:', err)
    error.value = err.message
  } finally {
    isSubmitting.value = false
  }
}

const deleteClient = async (client) => {
  if (confirm('Are you sure you want to delete this client?')) {
    const result = await callApi({
      query: 'DELETE FROM clients WHERE id = ?',
      params: [client.id],
    })
    if (result.success) {
      await fetchClients()
    }
  }
}

const handleImageClick = (path) => {
  if (path) {
    window.open(getFileUrl(path), '_blank')
  }
}

const clearEditFile = () => {
  editSelectedFile.value = null
}

onMounted(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    user.value = JSON.parse(userStr)
    fetchClients()
  }
})
</script>

<template>
  <div class="clients-view">
    <div class="header">
      <h2>
        <i class="fas fa-users"></i>
        Clients Management
      </h2>
      <button @click="showAddDialog = true" class="add-btn">
        <i class="fas fa-plus"></i>
        Add Client
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-state">
      <i class="fas fa-spinner fa-spin"></i>
      Loading clients...
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-message">
      <i class="fas fa-exclamation-circle"></i>
      {{ error }}
    </div>

    <!-- Filters Section -->
    <div v-if="!isLoading && !error" class="filters-section">
      <div class="filters-header">
        <h4>
          <i class="fas fa-filter"></i>
          Filters
        </h4>
        <button @click="clearFilters" class="clear-filters-btn">
          <i class="fas fa-times"></i>
          Clear Filters
        </button>
      </div>

      <div class="filters-grid">
        <div class="filter-group">
          <label>
            <i class="fas fa-user"></i>
            Name
          </label>
          <input type="text" v-model="filters.name" placeholder="Filter by name..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-envelope"></i>
            Email
          </label>
          <input type="text" v-model="filters.email" placeholder="Filter by email..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-phone"></i>
            Mobile
          </label>
          <input type="text" v-model="filters.mobile" placeholder="Filter by mobile..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-id-card"></i>
            ID Number
          </label>
          <input type="text" v-model="filters.idNo" placeholder="Filter by ID..." />
        </div>

        <div class="filter-group">
          <label>
            <i class="fas fa-user-tie"></i>
            Broker Status
          </label>
          <select v-model="filters.isBroker">
            <option :value="null">All</option>
            <option :value="true">Broker</option>
            <option :value="false">Not Broker</option>
          </select>
        </div>
      </div>
    </div>

    <div class="content" v-if="!isLoading && !error">
      <div v-if="filteredAndSortedClients.length === 0" class="no-results">
        <i class="fas fa-search fa-2x"></i>
        <p>No clients match the current filters</p>
      </div>

      <table v-else class="clients-table">
        <thead>
          <tr>
            <th @click="handleSort('name')" class="sortable">
              <i class="fas fa-user"></i>
              Name
              <i
                v-if="sortConfig.key === 'name'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('address')" class="sortable">
              <i class="fas fa-map-marker-alt"></i>
              Address
              <i
                v-if="sortConfig.key === 'address'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('email')" class="sortable">
              <i class="fas fa-envelope"></i>
              Email
              <i
                v-if="sortConfig.key === 'email'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('mobiles')" class="sortable">
              <i class="fas fa-phone"></i>
              Mobile
              <i
                v-if="sortConfig.key === 'mobiles'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('id_no')" class="sortable">
              <i class="fas fa-id-card"></i>
              ID No
              <i
                v-if="sortConfig.key === 'id_no'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th @click="handleSort('cars_count')" class="sortable">
              <i class="fas fa-car"></i>
              Cars
              <i
                v-if="sortConfig.key === 'cars_count'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th><i class="fas fa-file-alt"></i> ID Document</th>
            <th @click="handleSort('is_broker')" class="sortable">
              <i class="fas fa-user-tag"></i>
              Status
              <i
                v-if="sortConfig.key === 'is_broker'"
                :class="['fas', sortConfig.direction === 'asc' ? 'fa-sort-up' : 'fa-sort-down']"
              >
              </i>
            </th>
            <th><i class="fas fa-sticky-note"></i> Notes</th>
            <th><i class="fas fa-cog"></i> Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="client in filteredAndSortedClients" :key="client.id">
            <td>{{ client.name }}</td>
            <td>{{ client.address }}</td>
            <td>{{ client.email }}</td>
            <td>{{ client.mobiles }}</td>
            <td>{{ client.id_no }}</td>
            <td class="cars-count">
              <span class="badge cars" :class="{ 'has-cars': client.cars_count > 0 }">
                <i class="fas fa-car"></i>
                {{ client.cars_count }}
              </span>
            </td>
            <td class="id-document-cell">
              <div
                v-if="client.id_copy_path && isImageFile(client.id_copy_path)"
                class="image-preview"
                @click="handleImageClick(client.id_copy_path)"
              >
                <img :src="getFileUrl(client.id_copy_path)" :alt="'ID for ' + client.name" />
              </div>
              <a
                v-else-if="client.id_copy_path"
                :href="getFileUrl(client.id_copy_path)"
                target="_blank"
                class="document-link"
              >
                <i class="fas fa-file-download"></i>
                View Document
              </a>
              <span v-else class="no-document">
                <i class="fas fa-times-circle"></i>
                No ID
              </span>
            </td>
            <td>
              <div class="status-badges">
                <span class="badge client">
                  <i class="fas fa-user"></i>
                  Client
                </span>
                <span v-if="client.is_broker" class="badge broker">
                  <i class="fas fa-user-tie"></i>
                  Broker
                </span>
              </div>
            </td>
            <td class="notes-cell">
              <span v-if="client.notes" class="notes-content" :title="client.notes">
                {{
                  client.notes.length > 50 ? client.notes.substring(0, 50) + '...' : client.notes
                }}
              </span>
              <span v-else class="no-notes">No notes</span>
            </td>
            <td class="actions">
              <button @click="editClient(client)" class="action-btn edit">
                <i class="fas fa-edit"></i>
                Edit
              </button>
              <button v-if="isAdmin" @click="deleteClient(client)" class="action-btn delete">
                <i class="fas fa-trash"></i>
                Delete
              </button>
              <router-link :to="`/clients/${client.id}`" class="action-btn details" target="_blank">
                <i class="fas fa-info-circle"></i>
                Details
              </router-link>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Add Client Dialog -->
    <div v-if="showAddDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-user-plus"></i>
            Add New Client
          </h3>
          <button @click="showAddDialog = false" class="close-btn" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div class="form-group">
          <div class="input-group">
            <i class="fas fa-user input-icon"></i>
            <input
              v-model="newClient.name"
              placeholder="Name"
              class="input-field"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-map-marker-alt input-icon"></i>
            <input
              v-model="newClient.address"
              placeholder="Address"
              class="input-field"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-envelope input-icon"></i>
            <input
              v-model="newClient.email"
              placeholder="Email (Optional)"
              class="input-field"
              :class="{ error: validationError && newClient.email }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-phone input-icon"></i>
            <input
              v-model="newClient.mobiles"
              placeholder="Mobile *"
              class="input-field"
              :class="{ error: validationError && !newClient.mobiles }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-id-card input-icon"></i>
            <input
              v-model="newClient.id_no"
              placeholder="ID No *"
              class="input-field"
              :class="{ error: validationError && !newClient.id_no }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="checkbox-field">
            <input
              type="checkbox"
              id="is-broker"
              v-model="newClient.is_broker"
              :disabled="isSubmitting"
            />
            <label for="is-broker">Is Broker</label>
          </div>

          <div class="textarea-group">
            <label for="notes">
              <i class="fas fa-sticky-note"></i>
              Notes
            </label>
            <textarea
              id="notes"
              v-model="newClient.notes"
              placeholder="Add notes about the client..."
              rows="3"
              class="textarea-field"
              :disabled="isSubmitting"
            ></textarea>
          </div>

          <div class="file-upload">
            <label for="id-document">
              <i class="fas fa-file-upload"></i>
              ID Document: <span class="required">*</span>
            </label>
            <input
              type="file"
              id="id-document"
              @change="handleFileChange($event)"
              accept="image/*,.pdf"
              class="file-input"
              :class="{ error: validationError && !selectedFile }"
              :disabled="isSubmitting"
            />
            <span v-if="selectedFile" class="selected-file">
              <i class="fas fa-check"></i>
              {{ selectedFile.name }}
            </span>
          </div>

          <div v-if="validationError" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ validationError }}
          </div>
        </div>

        <div class="dialog-actions">
          <button @click="addClient" class="btn save-btn" :disabled="isSubmitting">
            <i class="fas fa-save"></i>
            <span v-if="isSubmitting">
              <i class="fas fa-spinner fa-spin"></i>
              Saving...
            </span>
            <span v-else>Add</span>
          </button>
          <button @click="showAddDialog = false" class="btn cancel-btn" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
            Cancel
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Client Dialog -->
    <div v-if="showEditDialog" class="dialog-overlay">
      <div class="dialog">
        <div class="dialog-header">
          <h3>
            <i class="fas fa-user-edit"></i>
            Edit Client
          </h3>
          <button @click="showEditDialog = false" class="close-btn" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
          </button>
        </div>

        <div v-if="error" class="error-message">
          <i class="fas fa-exclamation-circle"></i>
          {{ error }}
        </div>

        <div class="form-group">
          <div class="input-group">
            <i class="fas fa-user input-icon"></i>
            <input
              v-model="editingClient.name"
              placeholder="Name"
              class="input-field"
              :readonly="!isAdmin"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-map-marker-alt input-icon"></i>
            <input
              v-model="editingClient.address"
              placeholder="Address"
              class="input-field"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-envelope input-icon"></i>
            <input
              v-model="editingClient.email"
              placeholder="Email (Optional)"
              class="input-field"
              :class="{ error: validationError && editingClient.email }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-phone input-icon"></i>
            <input
              v-model="editingClient.mobiles"
              placeholder="Mobile *"
              class="input-field"
              :class="{ error: validationError && !editingClient.mobiles }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="input-group">
            <i class="fas fa-id-card input-icon"></i>
            <input
              v-model="editingClient.id_no"
              placeholder="ID No *"
              class="input-field"
              :class="{ error: validationError && !editingClient.id_no }"
              :disabled="isSubmitting"
            />
          </div>

          <div class="checkbox-field">
            <input
              type="checkbox"
              id="edit-is-broker"
              v-model="editingClient.is_broker"
              :disabled="isSubmitting"
            />
            <label for="edit-is-broker">Is Broker</label>
          </div>

          <div class="textarea-group">
            <label for="edit-notes">
              <i class="fas fa-sticky-note"></i>
              Notes
            </label>
            <textarea
              id="edit-notes"
              v-model="editingClient.notes"
              placeholder="Add notes about the client..."
              rows="3"
              class="textarea-field"
              :disabled="isSubmitting"
            ></textarea>
          </div>

          <div class="file-upload">
            <label for="edit-id-document">
              <i class="fas fa-file-upload"></i>
              ID Document: <span class="required">*</span>
            </label>
            <div v-if="editingClient.id_copy_path && !editSelectedFile" class="current-file">
              <i class="fas fa-file"></i>
              Current: <a :href="getFileUrl(editingClient.id_copy_path)" target="_blank">View ID</a>
            </div>
            <div v-if="editSelectedFile" class="new-file">
              <i class="fas fa-upload"></i>
              New file selected: {{ editSelectedFile.name }}
              <button
                type="button"
                @click="clearEditFile"
                class="clear-file-btn"
                :disabled="isSubmitting"
              >
                <i class="fas fa-times"></i>
                Clear
              </button>
            </div>
            <input
              type="file"
              id="edit-id-document"
              @change="handleFileChange($event, true)"
              accept="image/*,.pdf"
              class="file-input"
              :class="{
                error: validationError && !editingClient.id_copy_path && !editSelectedFile,
              }"
              :disabled="isSubmitting"
            />
          </div>

          <div v-if="validationError" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ validationError }}
          </div>
        </div>

        <div class="dialog-actions">
          <button @click="updateClient" class="btn save-btn" :disabled="isSubmitting">
            <i class="fas fa-save"></i>
            <span v-if="isSubmitting">
              <i class="fas fa-spinner fa-spin"></i>
              Saving...
            </span>
            <span v-else>Save</span>
          </button>
          <button @click="showEditDialog = false" class="btn cancel-btn" :disabled="isSubmitting">
            <i class="fas fa-times"></i>
            Cancel
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.clients-view {
  padding: 20px;
  width: 1000px;
}

.loading-state {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.loading-state i {
  margin-right: 8px;
  color: #3b82f6;
}

.error-message {
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  color: #dc2626;
  padding: 12px;
  border-radius: 4px;
  margin: 12px 0;
  display: flex;
  align-items: center;
}

.error-message i {
  margin-right: 8px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header h2 {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header h2 i {
  color: #3b82f6;
}

.add-btn {
  padding: 8px 16px;
  background-color: #10b981;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background-color 0.2s;
}

.add-btn:hover {
  background-color: #059669;
}

.add-btn:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.clients-table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 20px;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  border-radius: 8px;
  overflow: hidden;
}

.clients-table th,
.clients-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e5e7eb;
}

.clients-table th {
  background-color: #f8f9fa;
  font-weight: 600;
  color: #374151;
}

.clients-table th i {
  margin-right: 8px;
  color: #6b7280;
}

.clients-table tbody tr:hover {
  background-color: #f9fafb;
}

.btn {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  margin-right: 8px;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  transition: all 0.2s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.edit-btn {
  background-color: #3b82f6;
  color: white;
}

.edit-btn:hover {
  background-color: #2563eb;
}

.delete-btn {
  background-color: #ef4444;
  color: white;
}

.delete-btn:hover {
  background-color: #dc2626;
}

.save-btn {
  background-color: #10b981;
  color: white;
}

.save-btn:hover {
  background-color: #059669;
}

.cancel-btn {
  background-color: #6b7280;
  color: white;
}

.cancel-btn:hover {
  background-color: #4b5563;
}

.dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1000;
  padding: 20px;
  overflow-y: auto;
}

.dialog {
  background-color: white;
  padding: 24px;
  border-radius: 8px;
  min-width: 400px;
  max-width: 600px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-shrink: 0;
}

.dialog-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  color: #1f2937;
}

.dialog-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  font-size: 1.2em;
  transition: color 0.2s;
  flex-shrink: 0;
}

.close-btn:hover {
  color: #ef4444;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
  overflow-y: auto;
  flex: 1;
  min-height: 0;
}

.input-group {
  position: relative;
}

.input-icon {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: #6b7280;
}

.input-field {
  width: 100%;
  padding: 8px 12px 8px 36px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1em;
  transition: border-color 0.2s;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.input-field.error {
  border-color: #ef4444;
}

.input-field[readonly] {
  background-color: #f3f4f6;
  cursor: not-allowed;
}

.file-upload {
  margin-top: 8px;
}

.file-upload label {
  display: block;
  margin-bottom: 8px;
  font-weight: 500;
  color: #374151;
}

.file-upload label i {
  margin-right: 8px;
  color: #6b7280;
}

.file-input {
  width: 100%;
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  background-color: white;
}

.selected-file {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  color: #059669;
  font-size: 0.9em;
}

.current-file {
  margin-bottom: 8px;
  color: #4b5563;
  display: flex;
  align-items: center;
  gap: 8px;
}

.current-file a {
  color: #3b82f6;
  text-decoration: none;
}

.current-file a:hover {
  text-decoration: underline;
}

.dialog-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  flex-shrink: 0;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.id-document-cell {
  width: 120px;
  text-align: center;
}

.image-preview {
  width: 100px;
  height: 60px;
  margin: 0 auto;
  cursor: pointer;
  overflow: hidden;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  transition: transform 0.2s;
}

.image-preview:hover {
  transform: scale(1.05);
}

.image-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.document-link {
  color: #3b82f6;
  text-decoration: none;
  font-size: 0.9em;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  transition: color 0.2s;
}

.document-link:hover {
  color: #2563eb;
  text-decoration: underline;
}

.no-document {
  color: #6b7280;
  font-size: 0.9em;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.no-document i {
  color: #ef4444;
}

.required {
  color: #ef4444;
  margin-left: 4px;
}

.input-field.error {
  border-color: #ef4444;
  background-color: #fef2f2;
}

.file-input.error {
  border-color: #ef4444;
  background-color: #fef2f2;
}

.checkbox-field {
  margin-top: 8px;
}

.checkbox-field input {
  margin-right: 8px;
}

.checkbox-field label {
  margin-right: 16px;
}

.status-badges {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.badge.client {
  background-color: #dcfce7;
  color: #059669;
}

.badge.broker {
  background-color: #e0f2fe;
  color: #0284c7;
}

.badge.cars {
  background-color: #f1f5f9;
  color: #64748b;
  font-weight: 500;
  min-width: 40px;
  justify-content: center;
}

.badge.cars.has-cars {
  background-color: #dbeafe;
  color: #2563eb;
}

.cars-count {
  text-align: center;
}

.textarea-group {
  margin-top: 16px;
}

.textarea-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  color: #374151;
  font-weight: 500;
}

.textarea-group label i {
  color: #6b7280;
}

.textarea-field {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 1em;
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
  transition: all 0.2s;
}

.textarea-field:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.notes-cell {
  max-width: 200px;
}

.notes-content {
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: #374151;
  font-size: 0.9em;
}

.no-notes {
  color: #9ca3af;
  font-size: 0.9em;
  font-style: italic;
}

.filters-section {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 16px;
  margin: 16px 0;
  border: 1px solid #e5e7eb;
}

.filters-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.filters-header h4 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
  display: flex;
  align-items: center;
  gap: 8px;
}

.clear-filters-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  background-color: #ef4444;
  color: white;
  cursor: pointer;
  font-size: 0.875rem;
  transition: background-color 0.2s;
}

.clear-filters-btn:hover {
  background-color: #dc2626;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.filter-group label {
  display: flex;
  align-items: center;
  gap: 6px;
  color: #374151;
  font-weight: 500;
  font-size: 0.875rem;
}

.filter-group input,
.filter-group select {
  padding: 8px;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  font-size: 0.875rem;
  transition: border-color 0.2s;
}

.filter-group input:focus,
.filter-group select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.1);
}

.no-results {
  text-align: center;
  padding: 2rem;
  color: #6b7280;
  background-color: #f9fafb;
  border-radius: 8px;
  margin-top: 1rem;
}

.no-results i {
  margin-bottom: 1rem;
  color: #9ca3af;
}

.sortable {
  cursor: pointer;
  user-select: none;
  position: relative;
  padding-right: 24px !important;
}

.sortable i:last-child {
  position: absolute;
  right: 8px;
  top: 50%;
  transform: translateY(-50%);
  opacity: 0.5;
}

.sortable:hover {
  background-color: #f3f4f6;
}

.sortable i:last-child {
  opacity: 1;
}

.action-btn.details {
  background-color: #e3f2fd;
  color: #1976d2;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 0.9em;
  border: none;
  cursor: pointer;
  transition: background-color 0.2s;
}

.action-btn.details:hover {
  background-color: #bbdefb;
}

/* Responsive styles for mobile devices */
@media (max-width: 768px) {
  .clients-view {
    padding: 10px;
    width: 100%;
  }

  .dialog-overlay {
    padding: 10px;
  }

  .dialog {
    min-width: 300px;
    max-width: 100%;
    max-height: 95vh;
    padding: 16px;
  }

  .dialog-header {
    margin-bottom: 16px;
  }

  .dialog-header h3 {
    font-size: 1.1rem;
  }

  .form-group {
    gap: 12px;
    margin-bottom: 16px;
  }

  .dialog-actions {
    flex-direction: column;
    gap: 8px;
    margin-top: 16px;
  }

  .dialog-actions .btn {
    width: 100%;
    justify-content: center;
    padding: 12px;
    font-size: 1rem;
  }

  .input-field {
    padding: 12px 12px 12px 40px;
    font-size: 16px; /* Prevents zoom on iOS */
  }

  .textarea-field {
    font-size: 16px; /* Prevents zoom on iOS */
  }

  .filters-grid {
    grid-template-columns: 1fr;
  }

  .clients-table {
    font-size: 0.9rem;
  }

  .clients-table th,
  .clients-table td {
    padding: 8px;
  }
}

@media (max-width: 480px) {
  .dialog {
    padding: 12px;
  }

  .dialog-header h3 {
    font-size: 1rem;
  }

  .input-field {
    padding: 10px 10px 10px 35px;
  }

  .btn {
    padding: 8px 12px;
    font-size: 0.9rem;
  }
}

.new-file {
  margin-bottom: 8px;
  color: #059669;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 4px;
}

.new-file i {
  color: #059669;
}

.clear-file-btn {
  margin-left: auto;
  padding: 4px 8px;
  background-color: #ef4444;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.8rem;
  transition: background-color 0.2s;
}

.clear-file-btn:hover:not(:disabled) {
  background-color: #dc2626;
}

.clear-file-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
