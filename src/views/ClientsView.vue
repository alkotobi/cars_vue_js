<script setup>
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { useApi } from '../composables/useApi'
import TaskForm from '../components/car-stock/TaskForm.vue'

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

// Add dropdown state for task buttons
const openTaskDropdown = ref(null)

// Add TaskForm modal state
const showTaskForm = ref(false)
const selectedClientForTask = ref(null)
const selectedClientForActions = ref(null)

// Add teleport dropdown state
const teleportDropdown = ref({
  isOpen: false,
  clientId: null,
  position: { x: 0, y: 0 },
  buttonElement: null,
})

// Add task creation methods
const openTaskForClient = (client) => {
  console.log('openTaskForClient called with client:', client)

  // Close dropdown
  openTaskDropdown.value = null
  closeTeleportDropdown()

  // Set the selected client and show the task form
  selectedClientForTask.value = client
  showTaskForm.value = true
}

// Add methods for dropdown actions
const handleEditClient = (clientId) => {
  editClient(getClientById(clientId))
  closeTeleportDropdown()
}

const handleDeleteClient = (clientId) => {
  deleteClient(getClientById(clientId))
  closeTeleportDropdown()
}

const handleViewDetails = () => {
  closeTeleportDropdown()
}

const toggleTaskDropdown = (clientId, event) => {
  console.log('toggleTaskDropdown called with clientId:', clientId)

  if (openTaskDropdown.value === clientId) {
    openTaskDropdown.value = null
    closeTeleportDropdown()
  } else {
    openTeleportDropdown(clientId, event)
  }
}

// Add teleport dropdown functions
const openTeleportDropdown = (clientId, event) => {
  const button = event.currentTarget
  const rect = button.getBoundingClientRect()
  const dropdownWidth = 200 // Width of the dropdown
  const windowWidth = window.innerWidth
  const padding = 20 // Padding from screen edge

  // Calculate initial position
  let x = rect.left
  let y = rect.bottom + window.scrollY

  // Check if dropdown would go off-screen to the right
  if (x + dropdownWidth + padding > windowWidth) {
    // Position to the left of the button instead
    x = rect.right - dropdownWidth
  }

  // Check if dropdown would go off-screen to the left
  if (x < padding) {
    x = padding
  }

  teleportDropdown.value = {
    isOpen: true,
    clientId: clientId,
    position: { x, y },
    buttonElement: button,
  }

  // Close the regular dropdown
  openTaskDropdown.value = null
}

const closeTeleportDropdown = () => {
  teleportDropdown.value.isOpen = false
  teleportDropdown.value.clientId = null
}

// Add click outside handler
const handleClickOutside = (event) => {
  if (teleportDropdown.value.isOpen) {
    const dropdown = document.querySelector('.teleport-dropdown')
    const button = teleportDropdown.value.buttonElement

    if (dropdown && !dropdown.contains(event.target) && button && !button.contains(event.target)) {
      closeTeleportDropdown()
    }
  }
}

// Add scroll handler to close dropdown
const handleScroll = () => {
  if (teleportDropdown.value.isOpen) {
    closeTeleportDropdown()
  }
}

// Helper function to get client by ID
const getClientById = (clientId) => {
  return clients.value.find((client) => client.id === clientId)
}

// Close dropdown when clicking outside
const closeTaskDropdown = () => {
  openTaskDropdown.value = null
}

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

const getCurrentUser = () => {
  try {
    const userStr = localStorage.getItem('user')
    if (userStr) {
      user.value = JSON.parse(userStr)
    }
  } catch (err) {
    console.error('Error getting current user:', err)
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
  console.log('File selected:', file ? file.name : 'No file')

  if (isEdit) {
    editSelectedFile.value = file
    console.log('Edit file set to:', editSelectedFile.value ? editSelectedFile.value.name : 'null')
  } else {
    selectedFile.value = file
    console.log('Add file set to:', selectedFile.value ? selectedFile.value.name : 'null')
  }
}

// Debug function to test file upload
const debugFileUpload = async (file, clientId) => {
  console.log('=== DEBUG FILE UPLOAD ===')
  console.log('File:', file)
  console.log('Client ID:', clientId)

  try {
    const filename = `${clientId}.${file.name.split('.').pop()}`
    console.log('Generated filename:', filename)

    const uploadResult = await uploadFile(file, 'ids', filename)
    console.log('Upload result:', uploadResult)
    console.log('Upload result type:', typeof uploadResult)
    console.log('Upload result keys:', Object.keys(uploadResult))
    console.log('Relative path:', uploadResult.relativePath)
    console.log('File path:', uploadResult.file_path)

    return uploadResult
  } catch (error) {
    console.error('Debug upload error:', error)
    throw error
  }
}

// Debug function to test database update
const debugDatabaseUpdate = async (filePath, clientId) => {
  console.log('=== DEBUG DATABASE UPDATE ===')
  console.log('File path to save:', filePath)
  console.log('Client ID:', clientId)

  try {
    const updateResult = await callApi({
      query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
      params: [filePath, clientId],
    })

    console.log('Database update result:', updateResult)

    // Verify the update by fetching the client
    const verifyResult = await callApi({
      query: 'SELECT id_copy_path FROM clients WHERE id = ?',
      params: [clientId],
    })

    console.log('Verification result:', verifyResult)
    console.log('Stored file path:', verifyResult.data?.[0]?.id_copy_path)

    return updateResult
  } catch (error) {
    console.error('Debug database update error:', error)
    throw error
  }
}

// Test function for file upload
const testFileUpload = async () => {
  if (!editSelectedFile.value || !editingClient.value) {
    alert('Please select a file and have a client loaded for editing')
    return
  }

  console.log('=== TESTING FILE UPLOAD ===')
  try {
    const uploadResult = await debugFileUpload(editSelectedFile.value, editingClient.value.id)

    if (uploadResult.success) {
      const dbResult = await debugDatabaseUpdate(uploadResult.relativePath, editingClient.value.id)

      if (dbResult.success) {
        alert('Test successful! File uploaded and database updated. Check console for details.')
        // Refresh the client data to show the new file
        await fetchClients()
      } else {
        alert('Test failed: Database update failed. Check console for details.')
      }
    } else {
      alert('Test failed: File upload failed. Check console for details.')
    }
  } catch (error) {
    alert('Test failed with error: ' + error.message)
    console.error('Test error:', error)
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
          const uploadResult = await debugFileUpload(selectedFile.value, clientId)

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
  console.log('Editing client:', client)
  editingClient.value = { ...client }
  editSelectedFile.value = null

  // Reset the file input element
  const fileInput = document.getElementById('edit-id-document')
  if (fileInput) {
    fileInput.value = ''
  }

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
  const hasExistingFile = editingClient.value.id_copy_path
  const hasNewFile = editSelectedFile.value

  if (!hasExistingFile && !hasNewFile) {
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

    // Update client basic info first
    const result = await callApi({
      query: `
      UPDATE clients 
      SET name = ?, address = ?, email = ?, mobiles = ?, id_no = ?, is_broker = ?, is_client = 1, notes = ?
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
        editingClient.value.id,
      ],
    })

    if (result.success) {
      // If there's a new file selected, upload it and update the path
      if (editSelectedFile.value) {
        try {
          const uploadResult = await debugFileUpload(editSelectedFile.value, editingClient.value.id)

          if (uploadResult.success) {
            console.log('File uploaded successfully, updating client record')
            // Update the client record with the new file path
            const updateFileResult = await debugDatabaseUpdate(
              uploadResult.relativePath,
              editingClient.value.id,
            )

            if (!updateFileResult.success) {
              console.error('Failed to update client with new file path:', updateFileResult.error)
              error.value = 'Client updated but failed to save new ID document path'
            }
          } else {
            console.error('File upload failed:', uploadResult.error)
            error.value = 'Client updated but failed to upload new ID document'
          }
        } catch (err) {
          console.error('Error uploading file:', err)
          error.value = 'Client updated but failed to upload new ID document'
        }
      }

      // Reset form and close dialog
      showEditDialog.value = false
      editingClient.value = null
      editSelectedFile.value = null
      validationError.value = ''

      // Refresh the clients list to show updated data
      await fetchClients()
    } else {
      error.value = result.error
      console.error('Error updating client:', result.error)
    }
  } catch (err) {
    error.value = err.message
    console.error('Error in update client process:', err)
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

onMounted(async () => {
  await getCurrentUser()
  await fetchClients()

  // Add event listeners for teleport dropdown
  document.addEventListener('click', handleClickOutside)
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  // Remove event listeners
  document.removeEventListener('click', handleClickOutside)
  window.removeEventListener('scroll', handleScroll)
})

// Task form event handlers
const handleTaskSave = () => {
  console.log('Task saved for client:', selectedClientForTask.value)
  showTaskForm.value = false
  selectedClientForTask.value = null
}

const handleTaskCancel = () => {
  console.log('Task form canceled')
  showTaskForm.value = false
  selectedClientForTask.value = null
}
</script>

<template>
  <div class="clients-view" @click="closeTaskDropdown">
    <!-- Header Section -->
    <div class="header-section">
      <div class="header-content">
        <div class="header-info">
          <h1 class="page-title">
            <i class="fas fa-users"></i>
            Clients Management
          </h1>
        </div>
      </div>
      <div class="header-actions">
        <button @click="showAddDialog = true" class="btn-primary">
          <i class="fas fa-plus"></i>
          Add New Client
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="loading-container">
      <div class="loading-spinner">
        <i class="fas fa-spinner fa-spin"></i>
      </div>
      <p class="loading-text">Loading clients...</p>
    </div>

    <!-- Error Message -->
    <div v-if="error" class="error-container">
      <div class="error-content">
        <i class="fas fa-exclamation-triangle"></i>
        <div class="error-text">
          <h4>Error Loading Clients</h4>
          <p>{{ error }}</p>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div v-if="!isLoading && !error" class="main-content">
      <!-- Filters Section -->
      <div class="filters-section">
        <div class="filters-header">
          <div class="filters-title">
            <i class="fas fa-filter"></i>
            <h3>Search & Filters</h3>
          </div>
          <button @click="clearFilters" class="btn-secondary">
            <i class="fas fa-times"></i>
            Clear All
          </button>
        </div>

        <div class="filters-grid">
          <div class="filter-group">
            <label>
              <i class="fas fa-user"></i>
              Name
            </label>
            <input
              type="text"
              v-model="filters.name"
              placeholder="Search by client name..."
              class="filter-input"
            />
          </div>

          <div class="filter-group">
            <label>
              <i class="fas fa-envelope"></i>
              Email
            </label>
            <input
              type="text"
              v-model="filters.email"
              placeholder="Search by email..."
              class="filter-input"
            />
          </div>

          <div class="filter-group">
            <label>
              <i class="fas fa-phone"></i>
              Mobile
            </label>
            <input
              type="text"
              v-model="filters.mobile"
              placeholder="Search by mobile..."
              class="filter-input"
            />
          </div>

          <div class="filter-group">
            <label>
              <i class="fas fa-id-card"></i>
              ID Number
            </label>
            <input
              type="text"
              v-model="filters.idNo"
              placeholder="Search by ID..."
              class="filter-input"
            />
          </div>

          <div class="filter-group">
            <label>
              <i class="fas fa-user-tie"></i>
              Type
            </label>
            <select v-model="filters.isBroker" class="filter-select">
              <option :value="null">All Clients</option>
              <option :value="true">Brokers Only</option>
              <option :value="false">Regular Clients</option>
            </select>
          </div>
        </div>
      </div>

      <!-- Results Section -->
      <div class="results-section">
        <div v-if="filteredAndSortedClients.length === 0" class="empty-state">
          <div class="empty-icon">
            <i class="fas fa-search"></i>
          </div>
          <h3>No clients found</h3>
          <p>Try adjusting your search criteria or add a new client</p>
          <button @click="showAddDialog = true" class="btn-primary">
            <i class="fas fa-plus"></i>
            Add First Client
          </button>
        </div>

        <div v-else class="clients-container">
          <div class="table-header">
            <h3>Client List ({{ filteredAndSortedClients.length }} results)</h3>
          </div>

          <div class="table-wrapper">
            <table class="clients-table">
              <thead>
                <tr>
                  <th @click="handleSort('name')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-user"></i>
                      <span>Name</span>
                      <i
                        v-if="sortConfig.key === 'name'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th @click="handleSort('email')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-envelope"></i>
                      <span>Email</span>
                      <i
                        v-if="sortConfig.key === 'email'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th @click="handleSort('mobiles')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-phone"></i>
                      <span>Mobile</span>
                      <i
                        v-if="sortConfig.key === 'mobiles'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th @click="handleSort('id_no')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-id-card"></i>
                      <span>ID Number</span>
                      <i
                        v-if="sortConfig.key === 'id_no'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th @click="handleSort('cars_count')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-car"></i>
                      <span>Cars</span>
                      <i
                        v-if="sortConfig.key === 'cars_count'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th>
                    <div class="th-content">
                      <i class="fas fa-file-alt"></i>
                      <span>ID Document</span>
                    </div>
                  </th>
                  <th @click="handleSort('is_broker')" class="sortable">
                    <div class="th-content">
                      <i class="fas fa-user-tag"></i>
                      <span>Status</span>
                      <i
                        v-if="sortConfig.key === 'is_broker'"
                        :class="['sort-icon', sortConfig.direction === 'asc' ? 'asc' : 'desc']"
                      >
                      </i>
                    </div>
                  </th>
                  <th>
                    <div class="th-content">
                      <i class="fas fa-sticky-note"></i>
                      <span>Notes</span>
                    </div>
                  </th>
                  <th>
                    <div class="th-content">
                      <i class="fas fa-cog"></i>
                      <span>Actions</span>
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="client in filteredAndSortedClients" :key="client.id" class="client-row">
                  <td class="client-name">
                    <div class="name-content">
                      <div class="name-text">{{ client.name || 'Unnamed Client' }}</div>
                      <div class="address-text" v-if="client.address">{{ client.address }}</div>
                    </div>
                  </td>
                  <td class="client-email">
                    <span v-if="client.email" class="email-link">
                      <i class="fas fa-envelope"></i>
                      {{ client.email }}
                    </span>
                    <span v-else class="no-email">No email</span>
                  </td>
                  <td class="client-mobile">
                    <span class="mobile-number">
                      <i class="fas fa-phone"></i>
                      {{ client.mobiles }}
                    </span>
                  </td>
                  <td class="client-id">
                    <span class="id-number">{{ client.id_no }}</span>
                  </td>
                  <td class="cars-count">
                    <div class="cars-badge" :class="{ 'has-cars': client.cars_count > 0 }">
                      <i class="fas fa-car"></i>
                      <span>{{ client.cars_count || 0 }}</span>
                    </div>
                  </td>
                  <td class="id-document-cell">
                    <div
                      v-if="client.id_copy_path && isImageFile(client.id_copy_path)"
                      class="document-preview"
                      @click="handleImageClick(client.id_copy_path)"
                    >
                      <img :src="getFileUrl(client.id_copy_path)" :alt="'ID for ' + client.name" />
                      <div class="preview-overlay">
                        <i class="fas fa-eye"></i>
                      </div>
                    </div>
                    <a
                      v-else-if="client.id_copy_path"
                      :href="getFileUrl(client.id_copy_path)"
                      target="_blank"
                      class="document-link"
                    >
                      <i class="fas fa-file-download"></i>
                      <span>View</span>
                    </a>
                    <span v-else class="no-document">
                      <i class="fas fa-times-circle"></i>
                      <span>No ID</span>
                    </span>
                  </td>
                  <td class="client-status">
                    <div class="status-badges">
                      <span class="status-badge client">
                        <i class="fas fa-user"></i>
                        Client
                      </span>
                      <span v-if="client.is_broker" class="status-badge broker">
                        <i class="fas fa-user-tie"></i>
                        Broker
                      </span>
                    </div>
                  </td>
                  <td class="client-notes">
                    <div v-if="client.notes" class="notes-content" :title="client.notes">
                      {{
                        client.notes.length > 50
                          ? client.notes.substring(0, 50) + '...'
                          : client.notes
                      }}
                    </div>
                    <span v-else class="no-notes">No notes</span>
                  </td>
                  <td class="client-actions">
                    <div class="actions-group">
                      <!-- Task Button -->
                      <div class="task-dropdown-container">
                        <button
                          @click.stop="openTaskForClient(client)"
                          class="action-btn task"
                          title="Create task"
                        >
                          <i class="fas fa-tasks"></i>
                        </button>
                      </div>

                      <!-- Actions Dropdown Button -->
                      <div class="task-dropdown-container">
                        <button
                          @click.stop="toggleTaskDropdown(client.id, $event)"
                          class="action-btn actions"
                          title="Actions"
                        >
                          <i class="fas fa-ellipsis-v"></i>
                        </button>
                      </div>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- Global Actions Dropdown Menu -->
    <div
      v-if="openTaskDropdown"
      class="global-dropdown-menu"
      :style="{
        left: dropdownPosition.x + 'px',
        top: dropdownPosition.y + 'px',
      }"
      @click.stop
    >
      <button @click="editClient(selectedClientForActions)" class="task-dropdown-item">
        <i class="fas fa-edit"></i>
        Edit Client
      </button>
      <router-link
        :to="`/clients/${selectedClientForActions?.id}`"
        class="task-dropdown-item"
        target="_blank"
      >
        <i class="fas fa-info-circle"></i>
        View Details
      </router-link>
      <button
        v-if="isAdmin"
        @click="deleteClient(selectedClientForActions)"
        class="task-dropdown-item delete"
      >
        <i class="fas fa-trash"></i>
        Delete Client
      </button>
    </div>

    <!-- Add Client Dialog -->
    <div v-if="showAddDialog" class="modal-overlay" @click="showAddDialog = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
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

        <form @submit.prevent="addClient" class="modal-form">
          <div class="form-grid">
            <div class="form-group">
              <label for="name">
                <i class="fas fa-user"></i>
                Full Name *
              </label>
              <input
                id="name"
                v-model="newClient.name"
                type="text"
                placeholder="Enter client's full name"
                class="form-input"
                :class="{ error: validationError && !newClient.name }"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="mobile">
                <i class="fas fa-phone"></i>
                Mobile Number *
              </label>
              <input
                id="mobile"
                v-model="newClient.mobiles"
                type="tel"
                placeholder="Enter mobile number"
                class="form-input"
                :class="{ error: validationError && !newClient.mobiles }"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="id_no">
                <i class="fas fa-id-card"></i>
                ID Number *
              </label>
              <input
                id="id_no"
                v-model="newClient.id_no"
                type="text"
                placeholder="Enter ID number"
                class="form-input"
                :class="{ error: validationError && !newClient.id_no }"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="email">
                <i class="fas fa-envelope"></i>
                Email Address
              </label>
              <input
                id="email"
                v-model="newClient.email"
                type="email"
                placeholder="Enter email address (optional)"
                class="form-input"
                :class="{ error: validationError && newClient.email }"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-group full-width">
              <label for="address">
                <i class="fas fa-map-marker-alt"></i>
                Address
              </label>
              <input
                id="address"
                v-model="newClient.address"
                type="text"
                placeholder="Enter client's address"
                class="form-input"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-group">
              <label for="is_broker" class="checkbox-label">
                <input
                  id="is_broker"
                  type="checkbox"
                  v-model="newClient.is_broker"
                  :disabled="isSubmitting"
                  class="checkbox-input"
                />
                <span class="checkbox-custom"></span>
                <i class="fas fa-user-tie"></i>
                Is Broker
              </label>
            </div>

            <div class="form-group full-width">
              <label for="notes">
                <i class="fas fa-sticky-note"></i>
                Notes
              </label>
              <textarea
                id="notes"
                v-model="newClient.notes"
                placeholder="Add any additional notes about this client..."
                rows="3"
                class="form-textarea"
                :disabled="isSubmitting"
              ></textarea>
            </div>

            <div class="form-group full-width">
              <label for="id-document">
                <i class="fas fa-file-upload"></i>
                ID Document *
              </label>
              <div class="file-upload-area">
                <input
                  type="file"
                  id="id-document"
                  @change="handleFileChange($event)"
                  accept="image/*,.pdf"
                  class="file-input"
                  :class="{ error: validationError && !selectedFile }"
                  :disabled="isSubmitting"
                  required
                />
                <div class="file-upload-content">
                  <i class="fas fa-cloud-upload-alt"></i>
                  <p>Click to upload or drag and drop</p>
                  <span>Supports: JPG, PNG, PDF (Max 5MB)</span>
                </div>
              </div>
              <div v-if="selectedFile" class="selected-file">
                <i class="fas fa-check-circle"></i>
                <span>{{ selectedFile.name }}</span>
              </div>
            </div>
          </div>

          <div v-if="validationError" class="validation-error">
            <i class="fas fa-exclamation-triangle"></i>
            {{ validationError }}
          </div>

          <div class="modal-actions">
            <button
              type="button"
              @click="showAddDialog = false"
              class="btn-secondary"
              :disabled="isSubmitting"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button type="submit" class="btn-primary" :disabled="isSubmitting">
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-save"></i>
              {{ isSubmitting ? 'Adding...' : 'Add Client' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit Client Dialog -->
    <div v-if="showEditDialog" class="modal-overlay" @click="showEditDialog = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
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

        <form @submit.prevent="updateClient" class="modal-form">
          <div class="form-grid">
            <div class="form-group">
              <label for="edit-name">
                <i class="fas fa-user"></i>
                Full Name *
              </label>
              <input
                id="edit-name"
                v-model="editingClient.name"
                type="text"
                placeholder="Enter client's full name"
                class="form-input"
                :readonly="!isAdmin"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="edit-mobile">
                <i class="fas fa-phone"></i>
                Mobile Number *
              </label>
              <input
                id="edit-mobile"
                v-model="editingClient.mobiles"
                type="tel"
                placeholder="Enter mobile number"
                class="form-input"
                :class="{ error: validationError && !editingClient.mobiles }"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="edit-id_no">
                <i class="fas fa-id-card"></i>
                ID Number *
              </label>
              <input
                id="edit-id_no"
                v-model="editingClient.id_no"
                type="text"
                placeholder="Enter ID number"
                class="form-input"
                :class="{ error: validationError && !editingClient.id_no }"
                :disabled="isSubmitting"
                required
              />
            </div>

            <div class="form-group">
              <label for="edit-email">
                <i class="fas fa-envelope"></i>
                Email Address
              </label>
              <input
                id="edit-email"
                v-model="editingClient.email"
                type="email"
                placeholder="Enter email address (optional)"
                class="form-input"
                :class="{ error: validationError && editingClient.email }"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-group full-width">
              <label for="edit-address">
                <i class="fas fa-map-marker-alt"></i>
                Address
              </label>
              <input
                id="edit-address"
                v-model="editingClient.address"
                type="text"
                placeholder="Enter client's address"
                class="form-input"
                :disabled="isSubmitting"
              />
            </div>

            <div class="form-group">
              <label for="edit-is_broker" class="checkbox-label">
                <input
                  id="edit-is_broker"
                  type="checkbox"
                  v-model="editingClient.is_broker"
                  :disabled="isSubmitting"
                  class="checkbox-input"
                />
                <span class="checkbox-custom"></span>
                <i class="fas fa-user-tie"></i>
                Is Broker
              </label>
            </div>

            <div class="form-group full-width">
              <label for="edit-notes">
                <i class="fas fa-sticky-note"></i>
                Notes
              </label>
              <textarea
                id="edit-notes"
                v-model="editingClient.notes"
                placeholder="Add any additional notes about this client..."
                rows="3"
                class="form-textarea"
                :disabled="isSubmitting"
              ></textarea>
            </div>

            <div class="form-group full-width">
              <label for="edit-id-document">
                <i class="fas fa-file-upload"></i>
                ID Document
              </label>

              <div v-if="editingClient.id_copy_path" class="current-file">
                <div class="current-file-info">
                  <i class="fas fa-file"></i>
                  <span>Current document:</span>
                  <a
                    :href="getFileUrl(editingClient.id_copy_path)"
                    target="_blank"
                    class="file-link"
                  >
                    View Current ID
                  </a>
                </div>
              </div>

              <div class="file-upload-area">
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
                <div class="file-upload-content">
                  <i class="fas fa-cloud-upload-alt"></i>
                  <p>Upload new document (optional)</p>
                  <span>Will replace current document</span>
                </div>
              </div>

              <div v-if="editSelectedFile" class="selected-file">
                <i class="fas fa-check-circle"></i>
                <span>{{ editSelectedFile.name }}</span>
              </div>
            </div>
          </div>

          <div v-if="validationError" class="validation-error">
            <i class="fas fa-exclamation-triangle"></i>
            {{ validationError }}
          </div>

          <div class="modal-actions">
            <button
              type="button"
              @click="showEditDialog = false"
              class="btn-secondary"
              :disabled="isSubmitting"
            >
              <i class="fas fa-times"></i>
              Cancel
            </button>
            <button type="submit" class="btn-primary" :disabled="isSubmitting">
              <i v-if="isSubmitting" class="fas fa-spinner fa-spin"></i>
              <i v-else class="fas fa-save"></i>
              {{ isSubmitting ? 'Saving...' : 'Save Changes' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Task Form Modal -->
    <TaskForm
      v-if="selectedClientForTask"
      :entity-data="selectedClientForTask"
      entity-type="client"
      :is-visible="showTaskForm"
      @save="handleTaskSave"
      @cancel="handleTaskCancel"
    />

    <!-- Teleport Dropdown Menu -->
    <teleport to="body">
      <div
        v-if="teleportDropdown.isOpen"
        class="teleport-dropdown"
        :style="{
          position: 'absolute',
          left: teleportDropdown.position.x + 'px',
          top: teleportDropdown.position.y + 'px',
          zIndex: 9999,
        }"
      >
        <ul class="teleport-dropdown-menu">
          <li>
            <button @click="handleEditClient(teleportDropdown.clientId)" class="dropdown-item">
              <i class="fas fa-edit"></i>
              <span>Edit Client</span>
            </button>
          </li>
          <li>
            <router-link
              :to="`/clients/${getClientById(teleportDropdown.clientId)?.id}`"
              class="dropdown-item"
              target="_blank"
              @click="handleViewDetails"
            >
              <i class="fas fa-info-circle"></i>
              <span>View Details</span>
            </router-link>
          </li>
          <li>
            <button
              @click="openTaskForClient(getClientById(teleportDropdown.clientId))"
              class="dropdown-item"
            >
              <i class="fas fa-plus"></i>
              <span>Add New Task</span>
            </button>
          </li>
          <li v-if="isAdmin">
            <button
              @click="handleDeleteClient(teleportDropdown.clientId)"
              class="dropdown-item delete"
            >
              <i class="fas fa-trash"></i>
              <span>Delete Client</span>
            </button>
          </li>
        </ul>
      </div>
    </teleport>
  </div>
</template>

<style scoped>
.clients-view {
  width: 100%;
  margin: 0;
  padding: 24px;
  background-color: #f8fafc;
  min-height: 100vh;
}

.loading-container {
  text-align: center;
  padding: 2rem;
  color: #666;
}

.loading-spinner {
  margin-bottom: 8px;
}

.loading-text {
  font-size: 1.2rem;
  font-weight: 500;
}

.error-container {
  background-color: #fee2e2;
  border: 1px solid #ef4444;
  color: #dc2626;
  padding: 12px;
  border-radius: 4px;
  margin: 12px 0;
  display: flex;
  align-items: center;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.error-text {
  font-size: 1rem;
}

.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-info {
  display: flex;
  flex-direction: column;
}

.page-title {
  font-size: 1.5rem;
  font-weight: 600;
}

.page-subtitle {
  font-size: 1rem;
  color: #6b7280;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-primary {
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

.btn-primary:hover {
  background-color: #059669;
}

.btn-primary:disabled {
  background-color: #9ca3af;
  cursor: not-allowed;
}

.stats-section {
  background-color: #f9fafb;
  border-radius: 8px;
  padding: 12px;
  margin: 12px 0;
  border: 1px solid #e5e7eb;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.stat-card {
  background-color: white;
  border-radius: 8px;
  padding: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.stat-icon {
  font-size: 1.5rem;
  margin-bottom: 6px;
}

.stat-content {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-number {
  font-size: 1.25rem;
  font-weight: 600;
}

.stat-label {
  font-size: 0.8rem;
  color: #6b7280;
}

.main-content {
  padding: 20px;
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

.filters-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.filters-title h3 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
}

.btn-secondary {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 4px;
  font-size: 1.2em;
  transition: color 0.2s;
}

.btn-secondary:hover {
  color: #ef4444;
}

.filters-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
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

.results-section {
  margin-top: 20px;
}

.clients-container {
  background-color: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.table-header h3 {
  margin: 0;
  font-size: 1rem;
  color: #374151;
}

.clients-table {
  width: 100%;
  border-collapse: collapse;
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

.client-row {
  cursor: pointer;
}

.client-name {
  width: 20%;
}

.client-email {
  width: 20%;
}

.client-mobile {
  width: 15%;
}

.client-id {
  width: 10%;
}

.cars-count {
  width: 10%;
}

.id-document-cell {
  width: 15%;
}

.client-status {
  width: 10%;
}

.client-notes {
  width: 20%;
}

.client-actions {
  width: 10%;
}

.name-content {
  display: flex;
  flex-direction: column;
}

.name-text {
  font-weight: 500;
}

.address-text {
  color: #6b7280;
  font-size: 0.9em;
}

.email-link {
  color: #3b82f6;
  text-decoration: none;
}

.no-email {
  color: #6b7280;
  font-size: 0.9em;
}

.mobile-number {
  font-weight: 500;
}

.id-number {
  font-weight: 500;
}

.cars-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.cars-badge.has-cars {
  background-color: #dbeafe;
  color: #2563eb;
}

.document-preview {
  width: 100px;
  height: 60px;
  margin: 0 auto;
  cursor: pointer;
  overflow: hidden;
  border-radius: 4px;
  border: 1px solid #d1d5db;
  transition: transform 0.2s;
}

.document-preview:hover {
  transform: scale(1.05);
}

.document-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  transition: opacity 0.2s;
}

.preview-overlay:hover {
  opacity: 1;
}

.preview-overlay i {
  color: white;
  font-size: 1.5rem;
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

.selected-file {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 8px;
  padding: 8px 12px;
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
  color: #059669;
  font-size: 0.9em;
}

.selected-file .file-name {
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 8px;
}

.selected-file .file-info {
  font-size: 0.8em;
  color: #047857;
  font-style: italic;
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
  z-index: 1000;
  padding: 20px;
  overflow-y: auto;
}

.modal-content {
  background: white;
  border-radius: 16px;
  padding: 32px;
  min-width: 600px;
  max-width: 900px;
  width: 100%;
  max-height: 90vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px rgba(0, 0, 0, 0.15);
  animation: modalSlideIn 0.3s ease-out;
}

@keyframes modalSlideIn {
  from {
    opacity: 0;
    transform: translateY(-20px) scale(0.95);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #e5e7eb;
}

.modal-header h3 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.modal-header h3 i {
  color: #3b82f6;
}

.close-btn {
  background: none;
  border: none;
  color: #6b7280;
  cursor: pointer;
  padding: 8px;
  border-radius: 6px;
  transition: color 0.2s;
  flex-shrink: 0;
}

.close-btn:hover {
  background: #f3f4f6;
  color: #ef4444;
}

.modal-form {
  flex: 1;
  overflow-y: auto;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-bottom: 24px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-group.full-width {
  grid-column: span 2;
}

.form-group label {
  font-weight: 600;
  color: #374151;
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  gap: 8px;
}

.form-group label i {
  color: #6b7280;
}

.form-input,
.form-textarea {
  padding: 12px 16px;
  border: 2px solid #e5e7eb;
  border-radius: 8px;
  font-size: 1rem;
  transition: all 0.2s ease;
  background: white;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.form-input.error,
.form-textarea.error {
  border-color: #ef4444;
  background: #fef2f2;
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
  font-family: inherit;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  padding: 12px;
  border-radius: 8px;
  transition: background-color 0.2s ease;
}

.checkbox-label:hover {
  background: #f9fafb;
}

.checkbox-input {
  width: 18px;
  height: 18px;
  accent-color: #3b82f6;
}

.checkbox-label i {
  color: #6b7280;
}

.file-upload-area {
  position: relative;
  border: 2px dashed #d1d5db;
  border-radius: 8px;
  padding: 32px 16px;
  text-align: center;
  transition: all 0.2s ease;
  background: #f9fafb;
}

.file-upload-area:hover {
  border-color: #3b82f6;
  background: #eff6ff;
}

.file-input {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  opacity: 0;
  cursor: pointer;
}

.file-upload-content {
  pointer-events: none;
}

.file-upload-content i {
  font-size: 2rem;
  color: #6b7280;
  margin-bottom: 12px;
}

.file-upload-content p {
  font-size: 1rem;
  font-weight: 600;
  color: #374151;
  margin: 0 0 8px 0;
}

.file-upload-content span {
  font-size: 0.9rem;
  color: #6b7280;
}

.selected-file {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: linear-gradient(135deg, #dcfce7 0%, #bbf7d0 100%);
  border: 1px solid #bbf7d0;
  border-radius: 8px;
  color: #166534;
  font-weight: 500;
}

.selected-file i {
  color: #16a34a;
}

.current-file {
  background: #f3f4f6;
  padding: 12px 16px;
  border-radius: 8px;
  margin-bottom: 16px;
}

.current-file-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.9rem;
  color: #374151;
}

.file-link {
  color: #3b82f6;
  text-decoration: none;
  font-weight: 500;
}

.file-link:hover {
  text-decoration: underline;
}

.validation-error {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border: 1px solid #fecaca;
  color: #dc2626;
  padding: 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 12px;
  font-weight: 500;
}

.validation-error i {
  margin-right: 8px;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 16px;
  padding-top: 24px;
  border-top: 1px solid #e5e7eb;
  margin-top: auto;
}

.modal-actions .btn-primary,
.modal-actions .btn-secondary {
  padding: 12px 24px;
  font-size: 1rem;
  font-weight: 600;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.modal-actions .btn-primary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.modal-actions .btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
}

.modal-actions .btn-secondary {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
}

.modal-actions .btn-secondary:hover {
  background: #e5e7eb;
  border-color: #9ca3af;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
}

.empty-icon {
  margin-bottom: 24px;
}

.empty-icon i {
  color: #d1d5db;
  font-size: 4rem;
}

.empty-state h3 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #374151;
  margin: 0 0 12px 0;
}

.empty-state p {
  font-size: 1.1rem;
  color: #6b7280;
  margin: 0 0 24px 0;
}

.loading-container {
  text-align: center;
  padding: 60px 20px;
}

.loading-spinner {
  margin-bottom: 16px;
}

.loading-spinner i {
  font-size: 2rem;
  color: #3b82f6;
}

.loading-text {
  font-size: 1.2rem;
  font-weight: 500;
  color: #6b7280;
}

.error-container {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  border: 1px solid #fecaca;
  border-radius: 12px;
  padding: 24px;
  margin: 24px 0;
}

.error-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.error-content i {
  font-size: 1.5rem;
  color: #dc2626;
}

.error-text h4 {
  font-size: 1.2rem;
  font-weight: 600;
  color: #991b1b;
  margin: 0 0 8px 0;
}

.error-text p {
  font-size: 1rem;
  color: #7f1d1d;
  margin: 0;
}

/* Responsive styles for mobile devices */
@media (max-width: 768px) {
  .clients-view {
    padding: 10px;
    width: 100%;
  }

  .header-section {
    flex-direction: column;
    gap: 16px;
    align-items: flex-start;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .filters-grid {
    grid-template-columns: 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .form-group.full-width {
    grid-column: span 1;
  }

  .clients-table {
    font-size: 0.9rem;
  }

  .clients-table th,
  .clients-table td {
    padding: 8px;
  }

  .modal-overlay {
    padding: 10px;
  }

  .modal-content {
    min-width: 300px;
    max-width: 100%;
    max-height: 95vh;
    padding: 16px;
  }

  .modal-header {
    margin-bottom: 16px;
  }

  .modal-header h3 {
    font-size: 1.1rem;
  }

  .modal-form {
    gap: 12px;
    margin-bottom: 16px;
  }

  .modal-actions {
    flex-direction: column;
    gap: 8px;
    margin-top: 16px;
  }

  .modal-actions .btn {
    width: 100%;
    justify-content: center;
    padding: 12px;
    font-size: 1rem;
  }

  .form-input {
    padding: 12px 12px 12px 40px;
    font-size: 16px; /* Prevents zoom on iOS */
  }

  .form-textarea {
    font-size: 16px; /* Prevents zoom on iOS */
  }
}

@media (max-width: 480px) {
  .modal-content {
    padding: 12px;
  }

  .modal-header h3 {
    font-size: 1rem;
  }

  .form-input {
    padding: 10px 10px 10px 35px;
  }

  .btn {
    padding: 8px 12px;
    font-size: 0.9rem;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }
}

/* Additional modern styling */
.clients-view {
  max-width: 1400px;
  margin: 0 auto;
  padding: 24px;
  background-color: #f8fafc;
  min-height: 100vh;
}

.header-section {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 32px;
  border-radius: 16px;
  margin-bottom: 24px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 8px 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.page-title i {
  color: #fbbf24;
}

.page-subtitle {
  font-size: 1.1rem;
  opacity: 0.9;
  margin: 0;
}

.btn-primary {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 8px;
  padding: 12px 24px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.3s ease;
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(16, 185, 129, 0.4);
}

.btn-primary:disabled {
  background: #9ca3af;
  transform: none;
  box-shadow: none;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f3f4f6;
  color: #374151;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 10px 20px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.btn-secondary:hover {
  background: #e5e7eb;
  border-color: #9ca3af;
}

.stats-section {
  background: white;
  border-radius: 16px;
  padding: 24px;
  margin-bottom: 24px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border: 1px solid #e5e7eb;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.stat-card {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  transition: all 0.3s ease;
  border: 1px solid #e2e8f0;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.stat-icon {
  font-size: 2.5rem;
  margin-bottom: 16px;
  color: #3b82f6;
}

.stat-icon.brokers {
  color: #8b5cf6;
}

.stat-icon.cars {
  color: #10b981;
}

.stat-icon.active {
  color: #f59e0b;
}

.stat-number {
  font-size: 2rem;
  font-weight: 700;
  color: #1f2937;
  margin: 0 0 8px 0;
}

.stat-label {
  font-size: 0.9rem;
  color: #6b7280;
  font-weight: 500;
  margin: 0;
}

.main-content {
  background: white;
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.05);
  border: 1px solid #e5e7eb;
}

.filters-section {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  border: 1px solid #e2e8f0;
}

.filters-title h3 {
  font-size: 1.2rem;
  font-weight: 600;
  color: #1f2937;
}

.filters-title i {
  color: #3b82f6;
  font-size: 1.1rem;
}

.filter-input,
.filter-select {
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  padding: 12px 16px;
  font-size: 0.9rem;
  transition: all 0.2s ease;
  width: 100%;
}

.filter-input:focus,
.filter-select:focus {
  outline: none;
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.results-section {
  background: white;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.clients-container {
  padding: 0;
}

.table-header {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  padding: 20px 24px;
  border-bottom: 1px solid #e5e7eb;
}

.table-header h3 {
  font-size: 1.1rem;
  font-weight: 600;
  color: #1f2937;
  margin: 0;
}

.table-wrapper {
  overflow-x: auto;
}

.clients-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
}

.clients-table th {
  background: #f8fafc;
  padding: 16px 12px;
  font-weight: 600;
  color: #374151;
  border-bottom: 2px solid #e5e7eb;
  position: sticky;
  top: 0;
  z-index: 10;
}

.th-content {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.sortable {
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s ease;
}

.sortable:hover {
  background-color: #f1f5f9;
}

.sort-icon {
  position: absolute;
  right: 0;
  font-size: 0.8rem;
  opacity: 0.7;
}

.sort-icon.asc::before {
  content: '▲';
}

.sort-icon.desc::before {
  content: '▼';
}

.clients-table td {
  padding: 16px 12px;
  border-bottom: 1px solid #f1f5f9;
  vertical-align: middle;
}

.client-row {
  transition: all 0.2s ease;
}

.client-row:hover {
  background-color: #f8fafc;
  transform: translateX(4px);
}

.name-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.name-text {
  font-weight: 600;
  color: #1f2937;
  font-size: 1rem;
}

.address-text {
  color: #6b7280;
  font-size: 0.85rem;
}

.email-link {
  color: #3b82f6;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
  transition: color 0.2s ease;
}

.email-link:hover {
  color: #2563eb;
  text-decoration: underline;
}

.no-email {
  color: #9ca3af;
  font-style: italic;
  font-size: 0.9rem;
}

.mobile-number {
  font-weight: 600;
  color: #1f2937;
  display: flex;
  align-items: center;
  gap: 6px;
}

.id-number {
  font-weight: 600;
  color: #1f2937;
  font-family: 'Courier New', monospace;
  background: #f3f4f6;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.9rem;
}

.cars-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border-radius: 20px;
  font-size: 0.9rem;
  font-weight: 600;
  background: #f3f4f6;
  color: #6b7280;
  min-width: 50px;
  justify-content: center;
}

.cars-badge.has-cars {
  background: linear-gradient(135deg, #dbeafe 0%, #bfdbfe 100%);
  color: #1d4ed8;
  box-shadow: 0 2px 4px rgba(59, 130, 246, 0.2);
}

.document-preview {
  width: 80px;
  height: 50px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  position: relative;
  border: 2px solid #e5e7eb;
  transition: all 0.3s ease;
}

.actions-group {
  display: flex;
  gap: 8px;
  justify-content: center;
  position: relative;
  z-index: 1000;
}

.action-btn {
  padding: 8px;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  font-size: 0.9rem;
  min-width: 36px;
  height: 36px;
}
.action-btn.edit {
  background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  color: white;
}
.action-btn.edit:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
}
.action-btn.details {
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  text-decoration: none;
}
.action-btn.details:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
}
.action-btn.delete {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
}
.action-btn.delete:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(239, 68, 68, 0.3);
}

.action-btn.task {
  background: linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%);
  color: white;
}

.action-btn.task:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);
}

.action-btn.actions {
  background: linear-gradient(135deg, #6b7280 0%, #4b5563 100%);
  color: white;
}

.action-btn.actions:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(107, 114, 128, 0.3);
}

.task-dropdown-container {
  position: relative;
  z-index: 10000;
}

.task-dropdown-menu {
  position: absolute;
  top: 100%;
  right: 0;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  padding: 8px;
  z-index: 99999;
  min-width: 200px;
  margin-top: 8px;
  animation: dropdownSlideIn 0.2s ease-out;
}

@keyframes dropdownSlideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.task-dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  color: #374151;
  font-size: 0.9rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.task-dropdown-item:hover {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #1f2937;
  transform: translateX(4px);
}

.task-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.task-dropdown-item:hover i {
  color: #3b82f6;
}

.task-dropdown-item.delete {
  color: #dc2626;
}

.task-dropdown-item.delete:hover {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #991b1b;
}

.task-dropdown-item.delete:hover i {
  color: #dc2626;
}

/* Add separator before delete button */
.task-dropdown-item.delete {
  border-top: 1px solid #e5e7eb;
  margin-top: 4px;
  padding-top: 16px;
}

.global-dropdown-menu {
  position: absolute;
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  padding: 8px;
  z-index: 99999;
  min-width: 200px;
  animation: dropdownSlideIn 0.2s ease-out;
}

@keyframes dropdownSlideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.task-dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  color: #374151;
  font-size: 0.9rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.task-dropdown-item:hover {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #1f2937;
  transform: translateX(4px);
}

.task-dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.task-dropdown-item:hover i {
  color: #3b82f6;
}

.task-dropdown-item.delete {
  color: #dc2626;
}

.task-dropdown-item.delete:hover {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #991b1b;
}

.task-dropdown-item.delete:hover i {
  color: #dc2626;
}

/* Add separator before delete button */
.task-dropdown-item.delete {
  border-top: 1px solid #e5e7eb;
  margin-top: 4px;
  padding-top: 16px;
}

/* Teleport Dropdown Styles */
.teleport-dropdown {
  background: white;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
  padding: 8px;
  min-width: 200px;
}

.teleport-dropdown-menu {
  list-style: none;
  margin: 0;
  padding: 0;
}

.teleport-dropdown-menu li {
  margin: 0;
}

.teleport-dropdown-menu .dropdown-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  background: none;
  border: none;
  color: #374151;
  font-size: 0.9rem;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s ease;
  text-decoration: none;
}

.teleport-dropdown-menu .dropdown-item:hover {
  background: linear-gradient(135deg, #f8fafc 0%, #f1f5f9 100%);
  color: #1f2937;
  transform: translateX(4px);
}

.teleport-dropdown-menu .dropdown-item i {
  width: 16px;
  text-align: center;
  color: #6b7280;
}

.teleport-dropdown-menu .dropdown-item:hover i {
  color: #3b82f6;
}

.teleport-dropdown-menu .dropdown-item.delete {
  color: #dc2626;
}

.teleport-dropdown-menu .dropdown-item.delete:hover {
  background: linear-gradient(135deg, #fef2f2 0%, #fee2e2 100%);
  color: #991b1b;
}

.teleport-dropdown-menu .dropdown-item.delete:hover i {
  color: #dc2626;
}

/* Add separator before delete button */
.teleport-dropdown-menu .dropdown-item.delete {
  border-top: 1px solid #e5e7eb;
  margin-top: 4px;
  padding-top: 16px;
}
</style>
