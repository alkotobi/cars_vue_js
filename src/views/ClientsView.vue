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
})

const fetchClients = async () => {
  isLoading.value = true
  try {
    const result = await callApi({
      query: `
        SELECT * FROM clients where is_broker = 0
        ORDER BY name ASC
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
        INSERT INTO clients (name, address, email, mobiles, id_no, is_broker)
        VALUES (?, ?, ?, ?, ?, 0)
      `,
      params: [
        newClient.value.name,
        newClient.value.address,
        newClient.value.email,
        newClient.value.mobiles,
        newClient.value.id_no,
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

  // If no file is currently attached and no new file is selected
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
    // Update client basic info
    const result = await callApi({
      query: `
      UPDATE clients 
      SET name = ?, address = ?, email = ?, mobiles = ?, id_no = ?, is_broker = 0
      WHERE id = ?
    `,
      params: [
        editingClient.value.name,
        editingClient.value.address,
        editingClient.value.email,
        editingClient.value.mobiles,
        editingClient.value.id_no,
        editingClient.value.id,
      ],
    })

    if (result.success) {
      // If there's a new file selected, upload it using the new uploadFile method
      if (editSelectedFile.value) {
        try {
          const filename = `${editingClient.value.id}.${editSelectedFile.value.name.split('.').pop()}`
          const uploadResult = await uploadFile(editSelectedFile.value, 'ids', filename)

          if (uploadResult.success) {
            // Use the relative path returned from uploadFile
            await callApi({
              query: 'UPDATE clients SET id_copy_path = ? WHERE id = ?',
              params: [uploadResult.relativePath, editingClient.value.id],
            })
          }
        } catch (err) {
          console.error('Error uploading file:', err)
          error.value = 'Client updated but failed to upload new ID document'
        }
      }

      showEditDialog.value = false
      editingClient.value = null
      editSelectedFile.value = null
      validationError.value = ''
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
      <button @click="showAddDialog = true" class="add-btn" :disabled="!isAdmin">
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

    <div class="content" v-else>
      <table class="clients-table">
        <thead>
          <tr>
            <th><i class="fas fa-user"></i> Name</th>
            <th><i class="fas fa-map-marker-alt"></i> Address</th>
            <th><i class="fas fa-envelope"></i> Email</th>
            <th><i class="fas fa-phone"></i> Mobile</th>
            <th><i class="fas fa-id-card"></i> ID No</th>
            <th><i class="fas fa-file-alt"></i> ID Document</th>
            <th><i class="fas fa-cog"></i> Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="client in clients" :key="client.id">
            <td>{{ client.name }}</td>
            <td>{{ client.address }}</td>
            <td>{{ client.email }}</td>
            <td>{{ client.mobiles }}</td>
            <td>{{ client.id_no }}</td>
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
              <button @click="editClient(client)" class="btn edit-btn">
                <i class="fas fa-edit"></i>
                Edit
              </button>
              <button v-if="isAdmin" @click="deleteClient(client)" class="btn delete-btn">
                <i class="fas fa-trash"></i>
                Delete
              </button>
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
          <button @click="addClient" class="btn save-btn" :disabled="isSubmitting || !isAdmin">
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

          <div class="file-upload">
            <label for="edit-id-document">
              <i class="fas fa-file-upload"></i>
              ID Document: <span class="required">*</span>
            </label>
            <div v-if="editingClient.id_copy_path" class="current-file">
              <i class="fas fa-file"></i>
              Current: <a :href="getFileUrl(editingClient.id_copy_path)" target="_blank">View ID</a>
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
            <span v-if="editSelectedFile" class="selected-file">
              <i class="fas fa-check"></i>
              {{ editSelectedFile.name }}
            </span>
          </div>

          <div v-if="validationError" class="error-message">
            <i class="fas fa-exclamation-circle"></i>
            {{ validationError }}
          </div>
        </div>

        <div class="dialog-actions">
          <button @click="updateClient" class="btn save-btn" :disabled="isSubmitting || !isAdmin">
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
}

.dialog {
  background-color: white;
  padding: 24px;
  border-radius: 8px;
  min-width: 400px;
  max-width: 600px;
  width: 100%;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.dialog-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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
}

.close-btn:hover {
  color: #ef4444;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 20px;
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
</style>
